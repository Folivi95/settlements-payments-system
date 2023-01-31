package payment_provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	mocks3 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_provider"
)

func TestPaymentProvider_RetrieveBalanceForCurrency(t *testing.T) {
	t.Run("returns the account balances for a currency", func(t *testing.T) {
		retrieveBalanceUseCaseMock := &mocks3.RetrieveBankingCircleAccountFundsMock{ExecuteFunc: func(currency string, highRisk bool) (float64, float64, error) {
			return 1000000000.00, 1000000000.00, nil
		}}
		paymentProvider := payment_provider.NewPaymentProvider(retrieveBalanceUseCaseMock)
		beginOfDayAmount, intraDayAmount, err := paymentProvider.RetrieveBalanceForCurrency("EUR", false)
		assert.NoError(t, err)
		assert.Equal(t, 1000000000.00, beginOfDayAmount)
		assert.Equal(t, 1000000000.00, intraDayAmount)
	})
}
