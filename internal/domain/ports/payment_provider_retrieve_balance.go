//go:generate moq -out mocks/payment_provider_retrieve_balance_moq.go -pkg=mocks . PaymentProviderRetrieveBalance

package ports

import "github.com/saltpay/settlements-payments-system/internal/domain/models"

// PaymentProviderRetrieveBalance.
type PaymentProviderRetrieveBalance interface {
	// RetrieveBalanceForCurrency will fetch the balance for a certain currency and high risk value
	RetrieveBalanceForCurrency(currency models.CurrencyCode, highRisk bool) (float64, float64, error)
}
