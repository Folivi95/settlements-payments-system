//go:generate moq -out mocks/check_banking_circle_payment_status_moq.go -pkg mocks . CheckBankingCirclePaymentStatus
package ports

import (
	"context"
	"time"

	ppEvent "github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type CheckBankingCirclePaymentStatus interface {
	Execute(ctx context.Context, instruction ppEvent.PaymentInstruction, providerPaymentID ppEvent.ProviderPaymentID, bankingReference ppEvent.BankingReference, start time.Time) error
}
