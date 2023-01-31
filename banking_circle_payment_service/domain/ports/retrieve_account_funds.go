//go:generate moq -out mocks/retrieve_banking_circle_account_funds_moq.go -pkg mocks . RetrieveBankingCircleAccountFunds

package ports

type RetrieveBankingCircleAccountFunds interface {
	Execute(currency string, highRisk bool) (float64, float64, error)
}
