//go:build unit
// +build unit

package models_test

import (
	"testing"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
)

func TestSourceAccounts_FindAccountNumber(t *testing.T) {
	mockSourceAccounts := models.SourceAccounts{
		models.SourceAccount{
			Currency:   "EUR",
			IsHighRisk: false,
			AccountDetails: models.AccountDetails{
				Iban:      "IBAN_EUR_HR",
				AccountID: "64879db2-a8ba-34ee-c0d2381ebe4a",
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

	t.Run("check that we can find an account", func(t *testing.T) {
		is := is.New(t)

		expected, ok := mockSourceAccounts.FindAccountNumber("EUR", false)

		is.True(ok)
		is.Equal(expected, mockSourceAccounts[0].AccountDetails)
	})

	t.Run("check that we can't find an account for the given currency", func(t *testing.T) {
		is := is.New(t)

		_, ok := mockSourceAccounts.FindAccountNumber("XZY", false)

		is.True(!ok)
	})

	t.Run("check that we can't find an existing currency but high risk ", func(t *testing.T) {
		is := is.New(t)

		_, ok := mockSourceAccounts.FindAccountNumber("EUR", true)

		is.True(!ok)
	})
}
