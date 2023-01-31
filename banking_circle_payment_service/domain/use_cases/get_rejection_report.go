package use_cases

import (
	"time"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	bcStatus "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type GetBankingCircleRejectionReport struct {
	GetBankingCircleRejectionReportOptions
	observer observer
}

type GetBankingCircleRejectionReportOptions struct {
	PaymentAPI         bcStatus.BankingCircleAPI
	StatusCheckDelay   time.Duration
	MaxCheckIterations int
	MetricsClient      ports.MetricsClient
	PaymentNotifier    bcStatus.PaymentNotifier
	Now                func() time.Time
}

func NewGetBankingCircleRejectionReport(options GetBankingCircleRejectionReportOptions) GetBankingCircleRejectionReport {
	if options.Now == nil {
		options.Now = func() time.Time {
			return time.Now().UTC()
		}
	}
	if options.MaxCheckIterations == 0 {
		options.MaxCheckIterations = 10
	}
	return GetBankingCircleRejectionReport{
		GetBankingCircleRejectionReportOptions: options,
		observer:                               NewObserver(options.MetricsClient),
	}
}

func (m GetBankingCircleRejectionReport) Execute(date string) (models.RejectionReport, error) {
	report, err := m.PaymentAPI.GetRejectionReport(date)
	if err != nil {
		return models.RejectionReport{}, err
	}

	return report, nil
}
