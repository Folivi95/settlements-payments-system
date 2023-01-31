//go:generate moq -out mocks/get_payment_instruction_from_repo_moq.go -pkg mocks . GetPaymentInstructionFromRepo
package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// GetPaymentInstructionFromRepo deals with persistence of PaymentInstruction objects.
type GetPaymentInstructionFromRepo interface {
	Get(ctx context.Context, paymentInstructionID models.PaymentInstructionID) (models.PaymentInstruction, error)
	GetFromCorrelationID(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error)
}
