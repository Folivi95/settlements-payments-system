//go:generate moq -out mocks/get_payment_instruction_moq.go -pkg=mocks . GetPaymentInstruction

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type GetPaymentInstruction interface {
	Execute(ctx context.Context, paymentInstructionID models.PaymentInstructionID) (models.PaymentInstruction, error)
	RetrieveByCorrelationID(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error)
}
