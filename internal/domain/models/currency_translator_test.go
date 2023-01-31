//go:build unit
// +build unit

package models_test

import (
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestGetCurrencyIsoCode(t *testing.T) {
	cases := []struct {
		description       string
		currencyIsoNumber string
		expectedCurrency  models.CurrencyCode
	}{
		{
			description:       "Returns Currency ISO Code",
			currencyIsoNumber: "826",
			expectedCurrency:  models.GBP,
		},
		{
			description:       "Returns Empty String When Currency ISO Number Is Not Present",
			currencyIsoNumber: "747",
			expectedCurrency:  "",
		},
		{
			description:       "Supports Unpadded Number",
			currencyIsoNumber: "8",
			expectedCurrency:  models.ALL,
		},
		{
			description:       "Supports Padded Numbers",
			currencyIsoNumber: "008",
			expectedCurrency:  models.ALL,
		},
	}

	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			actualCurrency := models.GetCurrencyIsoCode(tt.currencyIsoNumber)
			if actualCurrency != tt.expectedCurrency {
				t.Errorf("got %q, want %q", actualCurrency, tt.expectedCurrency)
			}
		})
	}
}
