//go:generate moq -out mocks/make_banking_circle_payment_moq.go -pkg mocks . MakeBankingCirclePayment

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type MakeBankingCirclePayment interface {
	Execute(ctx context.Context, request models.PaymentInstruction) (paymentID models.ProviderPaymentID, err error)
}
