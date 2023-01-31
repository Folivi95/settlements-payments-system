//go:generate moq -out mocks/payment_provider_request_sender_moq.go -pkg=mocks . PaymentProviderRequestSender

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// PaymentProviderRequestSender allows sending a payment request to a payment provider.
type PaymentProviderRequestSender interface {
	// SendPaymentInstruction will send a payment request to a payment provider.
	// If there is no error, it means the payment request is accepted.
	SendPaymentInstruction(ctx context.Context, paymentInstruction models.PaymentInstruction) error
}
