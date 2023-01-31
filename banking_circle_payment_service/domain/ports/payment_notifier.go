//go:generate moq -out mocks/payment_notifier_moq.go -pkg mocks . PaymentNotifier

package ports

import (
	"context"

	ppEvents "github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentNotifier interface {
	SendPaymentStatus(ctx context.Context, event ppEvents.PaymentProviderEvent) error
}
