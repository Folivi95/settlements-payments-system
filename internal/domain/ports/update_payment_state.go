package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type UpdatePaymentState interface {
	Execute(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error
}
