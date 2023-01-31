//go:generate moq -out mocks/get_payment_report_moq.go -pkg=mocks . GetPaymentReport

package ports

import (
	"context"
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type GetPaymentReport interface {
	GetReport(ctx context.Context, time time.Time) (models.PaymentReport, error)
	GetCurrencyReport(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error)
	GetPaymentByMid(ctx context.Context, mid string, time time.Time) (models.PaymentInstruction, error)
}
