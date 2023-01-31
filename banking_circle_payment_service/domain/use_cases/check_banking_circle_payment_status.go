package use_cases

import (
	"context"
	"time"

	bcStatus "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	internalmodels "github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type CheckBankingCirclePaymentStatus struct {
	CheckBankingCirclePaymentStatusOptions
	observer observer
}

type CheckBankingCirclePaymentStatusOptions struct {
	PaymentAPI         bcStatus.BankingCircleAPI
	StatusCheckDelay   time.Duration
	MaxCheckIterations int
	MetricsClient      ports.MetricsClient
	PaymentNotifier    bcStatus.PaymentNotifier
	Now                func() time.Time
}

func NewCheckBankingCirclePaymentStatus(options CheckBankingCirclePaymentStatusOptions) CheckBankingCirclePaymentStatus {
	if options.Now == nil {
		options.Now = func() time.Time {
			return time.Now().UTC()
		}
	}
	if options.MaxCheckIterations == 0 {
		options.MaxCheckIterations = 10
	}
	return CheckBankingCirclePaymentStatus{
		CheckBankingCirclePaymentStatusOptions: options,
		observer:                               NewObserver(options.MetricsClient),
	}
}

// Execute will initiate the Banking Circle payment, then continuously check its status.
// Once the payment has processed (successfully or otherwise), it will send the outcome to
// the Payment Status Notifier via a PaymentProviderEvent.
// Execute will only return an error if it could not do any of the above.
func (m CheckBankingCirclePaymentStatus) Execute(ctx context.Context, instruction internalmodels.PaymentInstruction, paymentID internalmodels.ProviderPaymentID, bankingReference internalmodels.BankingReference, start time.Time) error {
	status, err := m.loopCheckPaymentStatus(ctx, paymentID, instruction)
	if err != nil {
		return m.sendFailedEvent(ctx, instruction, paymentID, bankingReference, internalmodels.FailureReason{
			Code:    internalmodels.TransportFailure,
			Message: err.Error(),
		})
	}
	processingSucceeded := status == bcStatus.Processed

	if processingSucceeded {
		if err := m.sendProcessedEvent(ctx, instruction, paymentID, bankingReference); err != nil {
			return err
		}
	} else {
		if err := m.sendFailedEvent(ctx, instruction, paymentID, bankingReference, internalmodels.FailureReason{
			Code:    mapStatusToFailureCode(status),
			Message: BankingCircleError{Status: status}.Error(),
		}); err != nil {
			return err
		}

		if status == bcStatus.PendingProcessing {
			return BankingCircleError{Status: status}
		}
	}

	m.observer.FinishedProcessing(ctx, instruction, processingSucceeded, status, time.Since(start))

	return nil
}

func (m CheckBankingCirclePaymentStatus) sendFailedEvent(
	ctx context.Context,
	paymentInstruction internalmodels.PaymentInstruction,
	paymentID internalmodels.ProviderPaymentID,
	bankingReference internalmodels.BankingReference,
	reason internalmodels.FailureReason,
) error {
	m.observer.PaymentIsUnprocessed(ctx, paymentInstruction.ID(), paymentInstruction.IncomingInstruction.Merchant.ContractNumber, reason.Code)
	event, err := internalmodels.NewPaymentProviderEvent(m.Now(), internalmodels.Failure, paymentInstruction, internalmodels.BC, paymentID, bankingReference, &reason)
	if err != nil {
		return err
	}
	return m.PaymentNotifier.SendPaymentStatus(ctx, event)
}

func (m CheckBankingCirclePaymentStatus) sendProcessedEvent(ctx context.Context, paymentInstruction internalmodels.PaymentInstruction, paymentID internalmodels.ProviderPaymentID, bankingReference internalmodels.BankingReference) error {
	event, err := internalmodels.NewPaymentProviderEvent(m.Now(), internalmodels.Processed, paymentInstruction, internalmodels.BC, paymentID, bankingReference, nil)
	if err != nil {
		return err
	}
	return m.PaymentNotifier.SendPaymentStatus(ctx, event)
}

func (m CheckBankingCirclePaymentStatus) loopCheckPaymentStatus(ctx context.Context, bankingCirclePaymentID internalmodels.ProviderPaymentID, instruction internalmodels.PaymentInstruction) (bcStatus.PaymentStatus, error) {
	var finalStatus bcStatus.PaymentStatus

	for i := 0; i < m.MaxCheckIterations; i++ {
		status, err := m.PaymentAPI.CheckPaymentStatus(bankingCirclePaymentID)

		if err != nil {
			m.observer.CheckPaymentFailed(ctx, instruction, err)

			if i == m.MaxCheckIterations-1 {
				m.observer.CheckPaymentStatusTotallyFailed(ctx, instruction.ID(), err, m.MaxCheckIterations)
				return "", TransportError{
					UnderlyingError: err,
					ID:              instruction.ID(),
					ContractNumber:  instruction.ContractNumber(),
				}
			}
		} else {
			m.observer.CheckPaymentSucceeded(ctx, instruction, status)

			if status == bcStatus.MissingFunding {
				m.observer.MissingFunds(ctx, instruction.IncomingInstruction.AccountNumber(), string(instruction.IncomingInstruction.IsoCode()))
			}

			if status != bcStatus.PendingProcessing {
				m.observer.NoLongerPending(ctx, instruction.ID(), instruction.ContractNumber(), status, i)
				return status, nil
			}

			m.observer.StillPending(ctx, instruction.ID(), instruction.ContractNumber(), i)
		}

		time.Sleep(m.StatusCheckDelay)
		finalStatus = status
	}

	return finalStatus, nil
}

func mapStatusToFailureCode(bankingCircleStatus bcStatus.PaymentStatus) internalmodels.PPEventFailureCode {
	switch bankingCircleStatus {
	case bcStatus.PendingProcessing:
		return internalmodels.StuckInPending
	case bcStatus.Rejected:
		return internalmodels.RejectedCode
	case bcStatus.MissingFunding:
		return internalmodels.MissingFunding
	default:
		return internalmodels.UnhandledPaymentProviderStatus
	}
}
