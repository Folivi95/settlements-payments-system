//go:generate moq -out mocks/track_payment_outcome_moq.go -pkg=mocks . TrackPaymentOutcome

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type TrackPaymentOutcome interface {
	Execute(ctx context.Context, ppEvent models.PaymentProviderEvent) error
}
