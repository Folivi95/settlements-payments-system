//go:build unit
// +build unit

package models_test

import (
	"testing"
	"time"

	"github.com/matryer/is"

	ppEvents "github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestNewPaymentProviderEvent(t *testing.T) {
	var (
		paymentProviderName      = ppEvents.BC
		paymentProviderPaymentID = ppEvents.ProviderPaymentID("a114f5bd-71f0-4980-8b04-6173f415610e")
		now                      = time.Now()
	)

	t.Run("happy scenarios", func(t *testing.T) {
		t.Run("Returning object must be of type: Failure", func(t *testing.T) {
			is := is.New(t)

			failureReason := &ppEvents.FailureReason{}
			failureReason.Code = "STUCK_PAYMENT"
			failureReason.Message = "Request timed out after x seconds"

			event, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Failure, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), paymentProviderName, paymentProviderPaymentID, "banking_reference", failureReason)

			is.NoErr(err)
			is.Equal(event.Type, ppEvents.Failure)
		})

		t.Run("Returning object must be of type: Submitted", func(t *testing.T) {
			is := is.New(t)
			event, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Submitted, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), paymentProviderName, paymentProviderPaymentID, "banking_reference", nil)
			is.NoErr(err)
			is.Equal(event.Type, ppEvents.Submitted)
		})

		t.Run("Returning object must be of type: Successful", func(t *testing.T) {
			is := is.New(t)
			event, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Processed, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), paymentProviderName, paymentProviderPaymentID, "banking_reference", nil)
			is.NoErr(err)
			is.Equal(event.Type, ppEvents.Processed)
		})
	})

	t.Run("errors scenarios", func(t *testing.T) {
		t.Run("empty provider name", func(t *testing.T) {
			is := is.New(t)
			emptyPaymentProviderName := ppEvents.PaymentProviderName("")
			_, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Submitted, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), emptyPaymentProviderName, paymentProviderPaymentID, "banking_reference", nil)
			is.True(err != nil)
		})

		t.Run("failed events", func(t *testing.T) {
			t.Run("cant have an empty failure object", func(t *testing.T) {
				is := is.New(t)
				_, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Failure, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), paymentProviderName, paymentProviderPaymentID, "banking_reference", &ppEvents.FailureReason{})
				is.True(err != nil)
			})
			t.Run("cant have a nil failure object", func(t *testing.T) {
				is := is.New(t)
				_, err := ppEvents.NewPaymentProviderEvent(now, ppEvents.Failure, ppEvents.NewPaymentInstruction(ppEvents.IncomingInstruction{}), paymentProviderName, paymentProviderPaymentID, "banking_reference", nil)
				is.True(err != nil)
			})
		})
	})
}
