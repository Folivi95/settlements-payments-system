package use_cases

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client"
	bcmodels "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	bcStatus "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	internalmodels "github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type MakeBankingCirclePayment struct {
	MakeBankingCirclePaymentOptions
	observer observer
}

type MakeBankingCirclePaymentOptions struct {
	PaymentAPI         bcStatus.BankingCircleAPI
	SourceAccounts     []bcmodels.SourceAccount
	MaxCheckIterations int
	MetricsClient      ports.MetricsClient
	PaymentNotifier    bcStatus.PaymentNotifier
	SubmissionNotifier bcStatus.PaymentNotifier
	Now                func() time.Time
}

func NewMakeBankingCirclePayment(options MakeBankingCirclePaymentOptions) MakeBankingCirclePayment {
	if options.Now == nil {
		options.Now = func() time.Time {
			return time.Now().UTC()
		}
	}
	if options.MaxCheckIterations == 0 {
		options.MaxCheckIterations = 10
	}

	return MakeBankingCirclePayment{
		MakeBankingCirclePaymentOptions: options,
		observer:                        NewObserver(options.MetricsClient),
	}
}

// Execute will initiate the Banking Circle payment, then continuously check its status.
// Once the payment has processed (successfully or otherwise), it will send the outcome to
// the Payment Status Notifier via a PaymentProviderEvent.
// Execute will only return an error if it could not do any of the above.
func (m MakeBankingCirclePayment) Execute(ctx context.Context, instruction internalmodels.PaymentInstruction) (paymentID internalmodels.ProviderPaymentID, err error) {
	start := m.Now()

	requestDTO, err := m.createRequestDTO(ctx, instruction)
	if err != nil {
		sendFail := m.sendFailedEvent(ctx, instruction, "", internalmodels.FailureReason{
			Code:    internalmodels.NoSourceAccount,
			Message: err.Error(),
		}, "")
		if sendFail != nil {
			return "", sendFail
		}
		return "", err
	}

	resp, err := m.requestPayment(ctx, instruction, requestDTO)
	if err != nil {
		sendFail := m.sendFailedEvent(ctx, instruction, "", internalmodels.FailureReason{
			Code: internalmodels.TransportFailure,
			Message: TransportError{
				UnderlyingError: err,
				ID:              instruction.ID(),
				ContractNumber:  instruction.ContractNumber(),
			}.Error(),
		}, "")
		if sendFail != nil {
			return "", sendFail
		}
		return "", err
	}

	err = m.sendSubmittedEvent(ctx, instruction, resp.PaymentID, resp.BankingReference, start)
	if err != nil {
		return "", err
	}

	return resp.PaymentID, nil
}

func (m MakeBankingCirclePayment) sendFailedEvent(
	ctx context.Context,
	paymentInstruction internalmodels.PaymentInstruction,
	paymentID internalmodels.ProviderPaymentID,
	reason internalmodels.FailureReason,
	bankingReference internalmodels.BankingReference,
) error {
	m.observer.PaymentIsUnprocessed(ctx, paymentInstruction.ID(), paymentInstruction.IncomingInstruction.Merchant.ContractNumber, reason.Code)
	event, err := internalmodels.NewPaymentProviderEvent(m.Now(), internalmodels.Failure, paymentInstruction, internalmodels.BC, paymentID, bankingReference, &reason)
	if err != nil {
		return err
	}
	return m.PaymentNotifier.SendPaymentStatus(ctx, event)
}

func (m MakeBankingCirclePayment) createRequestDTO(ctx context.Context, instruction internalmodels.PaymentInstruction) (single_payment_endpoint.RequestDto, error) {
	if err := addSourceAccount(&instruction, m.SourceAccounts); err != nil {
		m.observer.CouldntRequestBCRequest(ctx, instruction.ID(), err)
		return single_payment_endpoint.RequestDto{}, err
	}

	reqDto, err := ConvertPaymentInstructionToDto(ctx, &instruction, m.observer)
	if err != nil {
		m.observer.CouldntRequestBCRequest(ctx, instruction.ID(), err)
		return single_payment_endpoint.RequestDto{}, errors.Wrap(err, "convert payment instruction to instruction dto failed")
	}

	return reqDto, nil
}

func (m MakeBankingCirclePayment) requestPayment(ctx context.Context, instruction internalmodels.PaymentInstruction, requestDTO single_payment_endpoint.RequestDto) (single_payment_endpoint.ResponseDto, error) {
	var uidSlice []string
	resp, err := m.PaymentAPI.RequestPayment(ctx, requestDTO, &uidSlice)
	m.observer.RequestedPayment(ctx, instruction.ID(), resp.PaymentID)

	if err != nil {
		switch err.(type) {
		case http_client.UnauthorisedWithBankingCircleError:
			m.observer.RequestPaymentFailedUnauthorized(ctx, instruction, uidSlice, err)
		case http_client.InvalidPaymentRequestError:
			m.observer.RequestPaymentFailedBadRequest(ctx, instruction, uidSlice, err)
		default:
			m.observer.RequestPaymentFailed(ctx, instruction, uidSlice, err)
		}

		return single_payment_endpoint.ResponseDto{}, err
	}

	m.observer.RequestPaymentSucceeded(ctx, instruction, resp)
	return resp, nil
}

func (m MakeBankingCirclePayment) sendSubmittedEvent(ctx context.Context, instruction internalmodels.PaymentInstruction, id internalmodels.ProviderPaymentID, bankingReference internalmodels.BankingReference, start time.Time) error {
	event, err := internalmodels.NewPaymentProviderEvent(start, internalmodels.Submitted, instruction, internalmodels.BC, id, bankingReference, nil)
	if err != nil {
		return err
	}
	return m.SubmissionNotifier.SendPaymentStatus(ctx, event)
}

func addSourceAccount(paymentInstruction *internalmodels.PaymentInstruction, sourceAccounts bcmodels.SourceAccounts) error {
	sourceAccount, accountFound := sourceAccounts.FindAccountNumber(
		string(paymentInstruction.IncomingInstruction.IsoCode()),
		paymentInstruction.IncomingInstruction.Merchant.HighRisk,
	)
	if !accountFound {
		return NoSourceAccountError{
			IsoCode:  string(paymentInstruction.IncomingInstruction.IsoCode()),
			HighRisk: paymentInstruction.IncomingInstruction.Merchant.HighRisk,
		}
	}

	paymentInstruction.SetSourceAccount(sourceAccount.Iban)
	return nil
}
