//go:generate moq -out mocks/make_payment_moq.go -pkg=mocks . MakePayment

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// MakePayment is a use case that will try to make a payment as specified in the supplied request.
// If no error is returned, then the request to make payment is accepted.
type MakePayment interface {
	Execute(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error)
}
