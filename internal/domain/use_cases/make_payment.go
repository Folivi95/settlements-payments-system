package use_cases

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store/postgresql"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

const (
	PaymentsCount                       = "app_settlements_payments_count"
	PaymentsAmount                      = "app_settlements_payments_amount"
	AccountNumberIncorrectFormatCounter = "app_settlements_payments_incorrect_account_number"
)

type MakePayment struct {
	metricsClient                ports.MetricsClient
	paymentProviderRequestSender ports.PaymentProviderRequestSender
	aggregatePaymentStore        ports.StorePaymentInstructionToRepo
	paymentInstructionValidator  validation.Validator
}

func NewMakePayment(
	metricsClient ports.MetricsClient,
	paymentProviderRequestSender ports.PaymentProviderRequestSender,
	paymentInstructionRepo ports.StorePaymentInstructionToRepo,
	paymentRequestValidator validation.Validator,
) *MakePayment {
	return &MakePayment{
		metricsClient:                metricsClient,
		paymentProviderRequestSender: paymentProviderRequestSender,
		aggregatePaymentStore:        paymentInstructionRepo,
		paymentInstructionValidator:  paymentRequestValidator,
	}
}

func (m MakePayment) Execute(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
	validationRes := m.paymentInstructionValidator.ValidateIncomingInstruction(incomingInstruction)
	paymentInstruction := models.NewPaymentInstruction(incomingInstruction)

	if incomingInstruction.AccountNumber() != paymentInstruction.IncomingInstruction.AccountNumber() {
		zapctx.Warn(ctx, "flow_step #6: account number (IBAN) is in an incorrect format - it has spaces and/or lower case letters. But payment is still going ahead. Please contact CR to fix the account number.",
			zap.String("account number", incomingInstruction.AccountNumber()),
			zap.String("merchant_contract_number", incomingInstruction.Merchant.ContractNumber),
		)
		m.metricsClient.Count(ctx, AccountNumberIncorrectFormatCounter, 1, []string{string(paymentInstruction.IncomingInstruction.Payment.Currency.IsoCode)})
	}

	if !validationRes.IsValid() {
		paymentInstruction.Rejected(incomingInstruction, validationRes.Error())
		err := m.aggregatePaymentStore.Store(ctx, paymentInstruction)
		if err != nil {
			return "", err
		}

		zapctx.Error(ctx, "flow_step #6: incoming instruction is NOT valid",
			zap.String("merchant_contract_number", incomingInstruction.Merchant.ContractNumber),
			zap.Error(validationRes),
		)

		return "", validationRes
	}

	zapctx.Debug(ctx, "flow_step #6: payment instruction is valid",
		zap.String("id", string(paymentInstruction.ID())),
		zap.String("merchant_contract_number", incomingInstruction.Merchant.ContractNumber),
	)

	if paymentInstruction.IncomingInstruction.Payment.Currency.IsoCode == models.HUF {
		amount, err := strconv.ParseFloat(paymentInstruction.IncomingInstruction.Payment.Amount, 64)
		if err != nil {
			zapctx.Error(ctx, "Not able to convert amount into float64", zap.Error(err))
			return "", err
		}
		roundedAmount := math.Ceil(amount)
		paymentInstruction.IncomingInstruction.Payment.Amount = strconv.FormatFloat(roundedAmount, 'f', -1, 64)
	}

	// TODO: SubmitForProcessing should happen only after payment instruction
	// Route the PI to the correct payment provider
	paymentInstruction.RouteToPaymentProvider()
	paymentInstruction.SubmitForProcessing()

	err := m.aggregatePaymentStore.Store(ctx, paymentInstruction)
	if err != nil {
		switch {
		case errors.Is(err, postgresql.ErrDuplicate):
			paymentInstruction.SetStatus(models.Failed)
			paymentInstruction.AddEvent(models.PaymentInstructionEvent{
				Type:      models.DomainProcessingFailed,
				CreatedOn: time.Now(),
				Details: models.DomainProcessingFailedEventDetails{
					FailureReason: models.PIFailureReason{
						Code:    models.RejectedPayment,
						Message: "duplicated payment instruction",
					},
				},
			})
			errFromDB := m.aggregatePaymentStore.Store(ctx, paymentInstruction)
			if errFromDB != nil {
				return "", fmt.Errorf("unable to store duplicated payment as failed %w, %s", errFromDB, err.Error())
			}
			return "", err
		default:
			return "", err
		}
	}

	err = m.paymentProviderRequestSender.SendPaymentInstruction(ctx, paymentInstruction)
	if err != nil {
		return "", err
	}

	zapctx.Debug(ctx, "flow_step #7: payment instruction routed",
		zap.String("id", string(paymentInstruction.ID())),
		zap.String("merchant_contract_number", incomingInstruction.Merchant.ContractNumber),
		zap.String("payment_provider", string(paymentInstruction.PaymentProvider())),
	)

	// send metric with value per currency and increment counter
	amount, err := strconv.ParseFloat(incomingInstruction.Payment.Amount, 64)
	if err != nil {
		return paymentInstruction.ID(), err
	}

	labels := []string{string(paymentInstruction.PaymentProvider()), string(incomingInstruction.Payment.Currency.IsoCode)}
	m.metricsClient.Count(ctx, PaymentsAmount, int64(math.Ceil(amount*10000)/10000), labels)
	m.metricsClient.Count(ctx, PaymentsCount, 1, labels)

	return paymentInstruction.ID(), nil
}
