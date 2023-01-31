package payment_provider

import (
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentProvider struct {
	retrieveBalanceUseCase ports.RetrieveBankingCircleAccountFunds
}

func NewPaymentProvider(retrieveBalanceUseCase ports.RetrieveBankingCircleAccountFunds) *PaymentProvider {
	return &PaymentProvider{
		retrieveBalanceUseCase: retrieveBalanceUseCase,
	}
}

func (p PaymentProvider) RetrieveBalanceForCurrency(currency models.CurrencyCode, highRisk bool) (float64, float64, error) {
	return p.retrieveBalanceUseCase.Execute(string(currency), highRisk)
}
