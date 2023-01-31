//go:build unit
// +build unit

package banking_circle_payment_service_test

import (
	"fmt"
	"testing"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	. "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/use_cases"
)

func TestBankingCircleRetrieveAccountFundsUseCase_Execute(t *testing.T) {
	is := is.New(t)
	mockSourceAccounts := models.SourceAccounts{
		models.SourceAccount{
			Currency:   "EUR",
			IsHighRisk: false,
			AccountDetails: models.AccountDetails{
				Iban:            "IBAN_EUR_HR",
				AccountID:       "64879db2-a8ba-34ee-c0d2381ebe4a",
				MaxIntraDayLoan: 100,
			},
		},
		models.SourceAccount{
			Currency:   "GPB",
			IsHighRisk: false,
			AccountDetails: models.AccountDetails{
				Iban:      "IBAN_EUR",
				AccountID: "0e89a45e-9c7a-34ee-0b85d2abaf4e",
			},
		},
	}

	t.Run("happy path", func(t *testing.T) {
		t.Run("call to BC account balance returns available funds when intra day amount is a positive number ", func(t *testing.T) {
			beginOfDayAmount := float64(1000)
			intraDayAmount := float64(200)
			expectedAmount := float64(1200)
			currencyCode := "EUR"
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckAccountBalanceFunc: func(accountId string) (models.AccountBalance, error) {
				return models.AccountBalance{
					Result: []models.Balance{
						{
							Currency:         "EUR",
							BeginOfDayAmount: beginOfDayAmount,
							IntraDayAmount:   intraDayAmount,
						},
					},
				}, nil
			}}

			checkBCAccountFundsUseCase := NewRetrieveBankingCircleAccountFunds(
				RetrieveBankingCircleAccountFundsOptions{
					SourceAccounts: mockSourceAccounts,
					PaymentAPI:     mockBankingCircleAPIClient,
				},
			)

			actualAmount, intraDayLoan, err := checkBCAccountFundsUseCase.Execute(currencyCode, false)

			is.NoErr(err)
			is.Equal(len(mockBankingCircleAPIClient.CheckAccountBalanceCalls()), 1)
			is.Equal(expectedAmount, actualAmount)
			is.Equal(float64(100), intraDayLoan)
		})
		t.Run("call to BC account balance returns available funds when intra day amount is a negative number ", func(t *testing.T) {
			beginOfDayAmount := float64(1000)
			intraDayAmount := float64(-200)
			expectedAmount := float64(800)
			currencyCode := "EUR"
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckAccountBalanceFunc: func(accountId string) (models.AccountBalance, error) {
				return models.AccountBalance{
					Result: []models.Balance{
						{
							Currency:         "EUR",
							BeginOfDayAmount: beginOfDayAmount,
							IntraDayAmount:   intraDayAmount,
						},
					},
				}, nil
			}}

			checkBCAccountFundsUseCase := NewRetrieveBankingCircleAccountFunds(
				RetrieveBankingCircleAccountFundsOptions{
					SourceAccounts: mockSourceAccounts,
					PaymentAPI:     mockBankingCircleAPIClient,
				},
			)

			actualAmount, _, err := checkBCAccountFundsUseCase.Execute(currencyCode, false)

			is.NoErr(err)
			is.Equal(len(mockBankingCircleAPIClient.CheckAccountBalanceCalls()), 1)
			is.Equal(expectedAmount, actualAmount)
		})
	})

	t.Run("unhappy path", func(t *testing.T) {
		t.Run("given we get more then one balance response, then return an error", func(t *testing.T) {
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckAccountBalanceFunc: func(accountID string) (models.AccountBalance, error) {
				return models.AccountBalance{Result: []models.Balance{
					{
						Currency:         "",
						BeginOfDayAmount: 0,
						IntraDayAmount:   0,
					},
					{
						Currency:         "",
						BeginOfDayAmount: 0,
						IntraDayAmount:   0,
					},
				}}, nil
			}}

			checkBCAccountFundsUseCase := NewRetrieveBankingCircleAccountFunds(
				RetrieveBankingCircleAccountFundsOptions{
					SourceAccounts: mockSourceAccounts,
					PaymentAPI:     mockBankingCircleAPIClient,
				},
			)

			_, _, err := checkBCAccountFundsUseCase.Execute("EUR", false)

			isError := err != nil

			is.True(isError)
			is.Equal(len(mockBankingCircleAPIClient.CheckAccountBalanceCalls()), 1)
		})
		t.Run("given we get man empty balance response, then return an error", func(t *testing.T) {
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckAccountBalanceFunc: func(accountId string) (models.AccountBalance, error) {
				return models.AccountBalance{Result: []models.Balance{}}, nil
			}}

			checkBCAccountFundsUseCase := NewRetrieveBankingCircleAccountFunds(
				RetrieveBankingCircleAccountFundsOptions{
					SourceAccounts: mockSourceAccounts,
					PaymentAPI:     mockBankingCircleAPIClient,
				},
			)

			_, _, err := checkBCAccountFundsUseCase.Execute("EUR", false)

			isError := err != nil

			is.True(isError)
			is.Equal(len(mockBankingCircleAPIClient.CheckAccountBalanceCalls()), 1)
		})
		t.Run("given we get an error response from BC http call, then we return the same error", func(t *testing.T) {
			dummyError := fmt.Errorf("oops")

			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckAccountBalanceFunc: func(accountId string) (models.AccountBalance, error) {
				return models.AccountBalance{Result: []models.Balance{}}, dummyError
			}}

			checkBCAccountFundsUseCase := NewRetrieveBankingCircleAccountFunds(
				RetrieveBankingCircleAccountFundsOptions{
					SourceAccounts: mockSourceAccounts,
					PaymentAPI:     mockBankingCircleAPIClient,
				},
			)

			_, _, err := checkBCAccountFundsUseCase.Execute("EUR", false)

			is.Equal(err, dummyError)
			is.Equal(len(mockBankingCircleAPIClient.CheckAccountBalanceCalls()), 1)
		})
	})
}
