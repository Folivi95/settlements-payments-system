//go:generate moq -out mocks/get_banking_circle_rejection_report_moq.go -pkg mocks . GetBankingCircleRejectionReport

package ports

import "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"

type GetBankingCircleRejectionReport interface {
	Execute(date string) (models.RejectionReport, error)
}
