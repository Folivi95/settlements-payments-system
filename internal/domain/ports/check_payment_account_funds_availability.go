//go:generate moq -out mocks/check_payment_account_funds_availability.go -pkg=mocks . CheckPaymentAccountFundsAvailability

package ports

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// CheckPaymentAccountFundsAvailability is a use case that will check if the payment account has enough funds to do the payment.
// If error is returned, then we log the error and return a false.

type CheckPaymentAccountFundsAvailability interface {
	Execute(ctx context.Context, code models.CurrencyCode, amount float64, highRisk bool) (bool, error)
}
