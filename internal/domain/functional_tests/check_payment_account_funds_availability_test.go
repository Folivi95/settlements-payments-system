//go:build unit
// +build unit

package functional_tests

import (
	"context"
	"testing"

	is "github.com/matryer/is"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_provider"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
)

var dummyMetrics = testdoubles.DummyMetricsClient{}

func TestCheckPaymentAccountFundsAvailability_Execute(t *testing.T) {
	t.Run("Given we receive a request to check available funds and we have enough funds available for that account, then we return true", func(t *testing.T) {
		ctx := context.Background()
		is := is.New(t)

		cases := []struct {
			title        string
			paymentTotal float64
			available    float64
			loan         float64
		}{
			{
				title:        "we have enough balance",
				paymentTotal: 500,
				available:    500,
				loan:         0,
			},
			{
				title:        "we have enough balance using the loan",
				paymentTotal: 1000,
				available:    500,
				loan:         500,
			},
		}
		for _, tc := range cases {
			t.Run(tc.title, func(t *testing.T) {
				stub := &mocks.RetrieveBankingCircleAccountFundsMock{
					ExecuteFunc: func(currency string, highRisk bool) (float64, float64, error) {
						return tc.available, tc.loan, nil
					},
				}
				paymentProviderAdapter := payment_provider.NewPaymentProvider(stub)

				useCase := use_cases.NewCheckPaymentAccountFundsAvailability(dummyMetrics, paymentProviderAdapter, message.NewPrinter(language.English))
				fundsAvailability, err := useCase.Execute(ctx, "EUR", tc.paymentTotal, false)

				is.NoErr(err)
				is.True(fundsAvailability)
			})
		}
	})

	t.Run("Given we receive a request to check available funds and we don't have enough funds available for that account, then we return false", func(t *testing.T) {
		ctx := context.Background()
		is := is.New(t)

		cases := []struct {
			title        string
			paymentTotal float64
			available    float64
			loan         float64
		}{
			{
				title:        "we don't have enough balance",
				paymentTotal: 100000000,
				available:    500,
				loan:         0,
			},
			{
				title:        "we don't have enough balance using the loan",
				paymentTotal: 100000000,
				available:    500,
				loan:         500,
			},
		}
		for _, tc := range cases {
			t.Run(tc.title, func(t *testing.T) {
				stub := &mocks.RetrieveBankingCircleAccountFundsMock{
					ExecuteFunc: func(currency string, highRisk bool) (float64, float64, error) {
						return tc.available, tc.loan, nil
					},
				}
				paymentProviderAdapter := payment_provider.NewPaymentProvider(stub)

				useCase := use_cases.NewCheckPaymentAccountFundsAvailability(dummyMetrics, paymentProviderAdapter, message.NewPrinter(language.English))
				fundsAvailability, err := useCase.Execute(ctx, "EUR", tc.paymentTotal, false)

				is.NoErr(err)
				is.True(!fundsAvailability)
			})
		}
	})
}
