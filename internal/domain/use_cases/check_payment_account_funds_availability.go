package use_cases

import (
	"context"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"
	"golang.org/x/text/currency"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

const (
	missingAccountFunds = "app_settlements_provider_missing_account_funds"
	usingLoan           = "app_settlements_provider_missing_account_funds_using_loan"
)

type CheckPaymentAccountFundsAvailability struct {
	metricsClient                  ports.MetricsClient
	printer                        *message.Printer
	paymentProviderRetrieveBalance ports.PaymentProviderRetrieveBalance
}

func NewCheckPaymentAccountFundsAvailability(
	metricsClient ports.MetricsClient,
	paymentProviderRetrieveBalance ports.PaymentProviderRetrieveBalance,
	printer *message.Printer,
) *CheckPaymentAccountFundsAvailability {
	return &CheckPaymentAccountFundsAvailability{
		metricsClient:                  metricsClient,
		paymentProviderRetrieveBalance: paymentProviderRetrieveBalance,
		printer:                        printer,
	}
}

func (c CheckPaymentAccountFundsAvailability) Execute(ctx context.Context, currencyCode models.CurrencyCode, amount float64, highRisk bool) (bool, error) {
	amountFormatted := c.formatCurrencyAmount(currencyCode, amount)

	balance, maxIntraDayLoan, err := c.paymentProviderRetrieveBalance.RetrieveBalanceForCurrency(currencyCode, highRisk)
	if err != nil {
		zapctx.Error(ctx, "[CheckPaymentAccountFundsAvailability] (Execute) error occurred when retrieving balance for currency",
			zap.String("currency", string(currencyCode)),
			zap.String("amount_needed", amountFormatted),
		)

		return false, err
	}

	balanceFormatted := c.formatCurrencyAmount(currencyCode, balance)
	intraDayLoanFormatted := c.formatCurrencyAmount(currencyCode, maxIntraDayLoan)

	if amount > balance+maxIntraDayLoan {
		zapctx.Error(ctx, "[CheckPaymentAccountFundsAvailability.Execute] Not enough funds",
			zap.String("balance", balanceFormatted),
			zap.String("currency", string(currencyCode)),
			zap.Bool("high_risk", highRisk),
			zap.String("amount_needed", amountFormatted),
		)

		c.metricsClient.Count(ctx, missingAccountFunds, 1, []string{string(currencyCode)})
		return false, nil
	}
	if maxIntraDayLoan > 0 && amount > balance && amount < balance+maxIntraDayLoan {
		zapctx.Info(ctx, "[CheckPaymentAccountFundsAvailability.Execute] Using intra day loan to pay",
			zap.String("amount_needed", amountFormatted),
			zap.String("currency", string(currencyCode)),
			zap.Bool("high_risk", highRisk),
			zap.String("account_balance", balanceFormatted),
			zap.String("max_loan", intraDayLoanFormatted),
		)

		c.metricsClient.Count(ctx, usingLoan, 1, []string{string(currencyCode)})
		return true, nil
	}

	zapctx.Info(ctx, "[CheckPaymentAccountFundsAvailability.Execute] Account balance to do payments is OK.",
		zap.String("amount_needed", amountFormatted),
		zap.String("currency", string(currencyCode)),
		zap.Bool("high_risk", highRisk),
		zap.String("account_balance", balanceFormatted),
		zap.String("loan", intraDayLoanFormatted),
	)

	return true, nil
}

// formatCurrencyAmount formats a currency amount for the locale associated with the currency code.
// e.g. with input 1000000.0666 it outputs 1,000,000.07 for GBP.
func (c CheckPaymentAccountFundsAvailability) formatCurrencyAmount(currencyCode models.CurrencyCode, amount float64) string {
	cur, _ := currency.ParseISO(string(currencyCode)) // ignoring error as if an invalid currencyCode is provided, this method will not be called.
	scale, _ := currency.Cash.Rounding(cur)           // fractional digits
	amountFormatter := number.Decimal(amount, number.Scale(scale))
	return c.printer.Sprint(amountFormatter)
}
