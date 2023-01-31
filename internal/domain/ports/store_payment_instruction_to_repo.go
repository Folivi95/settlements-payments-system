//go:generate moq -out mocks/store_payment_instruction_to_repo_moq.go -pkg mocks . StorePaymentInstructionToRepo

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type StorePaymentInstructionToRepo interface {
	Store(ctx context.Context, instruction models.PaymentInstruction) error
	UpdatePayment(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error
}
