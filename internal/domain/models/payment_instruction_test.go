//go:build unit
// +build unit

package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
)

func TestPaymentInstruction(t *testing.T) {
	t.Run("submit a PI for processing updates its state", func(t *testing.T) {
		paymentInstruction, _, err := testhelpers.ValidPaymentInstruction()
		require.NoError(t, err)

		paymentInstruction.SubmitForProcessing()
		assert.Equal(t, paymentInstruction.GetStatus(), models.SubmittedForProcessing)
		assert.Equal(t, paymentInstruction.Version(), 2)
		assert.Equal(t, len(paymentInstruction.Events()), 2)
		assert.Equal(t, paymentInstruction.Events()[1].Type, models.DomainSubmittedToPaymentProvider)
	})

	t.Run("generate ID based from contents of PI", func(t *testing.T) {
		paymentInstruction := models.NewPaymentInstruction(models.IncomingInstruction{
			Merchant: models.Merchant{
				ContractNumber: "800900",
			},
			Payment: models.Payment{
				Amount:        "234.56",
				ExecutionDate: time.Date(2021, time.August, 12, 0, 0, 0, 0, time.UTC),
			},
		})

		// when
		uniqueID := paymentInstruction.BusinessID()

		// then
		assert.Equal(t, uniqueID, "800900#2021-08-12#234.56")
	})

	t.Run("routes the PI to the correct banking provider", func(t *testing.T) {
		testData := []struct {
			sender           string
			source           string
			currency         models.Currency
			expectedProvider models.PaymentProviderType
			description      string
		}{
			{"ISB", "Way4", models.Currency{IsoCode: "USD", IsoNumber: "840"}, models.Islandsbanki, "ISB from Way4 USD"},
			{"ISB", "Way4", models.Currency{IsoCode: "GBP", IsoNumber: "826"}, models.Islandsbanki, "ISB from Way4 GBP"},
			{"ISB", "", models.Currency{IsoCode: "EUR", IsoNumber: "978"}, models.Islandsbanki, "ISB from Way4 EUR"},
			{"RB", "Way4", models.Currency{IsoCode: "ISK", IsoNumber: "352"}, models.Islandsbanki, "RB from Way4 ISK"},
			{"RB_12", "Way4", models.Currency{IsoCode: "ISK", IsoNumber: "352"}, models.Islandsbanki, "RB from Way4 ISK"},
			{"SAXO", "Way4", models.Currency{IsoCode: "EUR", IsoNumber: "978"}, models.BankingCircle, "SAXO from Way4 EUR"},
			{"SAXO", "", models.Currency{IsoCode: "HUF", IsoNumber: "348"}, models.BankingCircle, "SAXO HUF"},
			{"", "Solanteq", models.Currency{IsoCode: "ISK", IsoNumber: "352"}, models.Islandsbanki, "from Solanteq ISK"},
			{"", "Solanteq", models.Currency{IsoCode: "EUR", IsoNumber: "978"}, models.BankingCircle, "from Solanteq EUR"},
		}

		for _, datum := range testData {
			var (
				incomingInstruction = testhelpers.NewIncomingInstructionBuilder().WithMetadataSender(datum.sender).WithMetadataSource(datum.source).WithCurrency(datum.currency).Build()
				paymentInstruction  = testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(incomingInstruction).Build()
			)
			paymentInstruction.RouteToPaymentProvider()
			assert.Equal(t, datum.expectedProvider, paymentInstruction.PaymentProvider(), datum.description)
		}
	})

	t.Run("fixes a payment instruction with IBAN in the incorrect format", func(t *testing.T) {
		var (
			incorrectAccountNumber = "gb 33BUKb202 0155555 5555"
			incomingInstruction    = testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber(incorrectAccountNumber).Build()
			paymentInstruction     = models.NewPaymentInstruction(incomingInstruction)
		)

		assert.Equal(t, "GB33BUKB20201555555555", paymentInstruction.IncomingInstruction.Merchant.Account.AccountNumber)
	})
}
