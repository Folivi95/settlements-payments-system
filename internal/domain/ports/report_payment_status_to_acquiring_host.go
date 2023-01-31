//go:generate moq -out mocks/acquiring_host_producer_moq.go -pkg=mocks . PaymentExporterProducer

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentExporterProducer interface {
	ReportPaymentStatus(ctx context.Context, ppEvent models.PaymentProviderEvent) error
}
