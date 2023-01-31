package use_cases

import (
	"fmt"

	bcmodels "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
)

type RetrieveBankingCircleAccountFundsOptions struct {
	SourceAccounts bcmodels.SourceAccounts
	PaymentAPI     ports.BankingCircleAPI
}

type RetrieveBankingCircleAccountFunds struct {
	RetrieveBankingCircleAccountFundsOptions
}

func NewRetrieveBankingCircleAccountFunds(options RetrieveBankingCircleAccountFundsOptions) RetrieveBankingCircleAccountFunds {
	return RetrieveBankingCircleAccountFunds{
		RetrieveBankingCircleAccountFundsOptions: options,
	}
}

func (c RetrieveBankingCircleAccountFunds) Execute(currency string, highRisk bool) (float64, float64, error) {
	account, ok := c.SourceAccounts.FindAccountNumber(currency, highRisk)
	if !ok {
		return 0, 0, fmt.Errorf("[CheckBankingCircleAccountFunds Execute] account not found for the given currency: %s and highrisk: %t", currency, highRisk)
	}

	accountBalance, err := c.PaymentAPI.CheckAccountBalance(account.AccountID)
	if err != nil {
		return 0, 0, err
	}

	if len(accountBalance.Result) != 1 {
		return 0, 0, fmt.Errorf("[CheckBankingCircleAccountFunds Execute] expected a single entry but got %d", len(accountBalance.Result))
	}

	results := accountBalance.Result[0]

	return results.BeginOfDayAmount + results.IntraDayAmount, account.MaxIntraDayLoan, nil
}
