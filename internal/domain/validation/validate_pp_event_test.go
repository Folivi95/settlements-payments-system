//go:build unit
// +build unit

package validation_test

import (
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

func TestValidatePaymentProviderEvent(t *testing.T) {
	t.Run("invalid successful/failed payments", func(t *testing.T) {
		cases := []struct {
			Name  string
			Event models.PaymentProviderEvent
		}{
			{
				Name: "invalid successful payments: missing provider name",
				Event: models.PaymentProviderEvent{
					CreatedOn:                time.Time{},
					Type:                     models.Processed,
					PaymentInstruction:       models.PaymentInstruction{},
					PaymentProviderName:      "",
					PaymentProviderPaymentID: "123",
					FailureReason:            models.FailureReason{},
				},
			},
			{
				Name: "invalid successful payments: missing payment provider id",
				Event: models.PaymentProviderEvent{
					CreatedOn:                time.Time{},
					Type:                     models.Processed,
					PaymentInstruction:       models.PaymentInstruction{},
					PaymentProviderName:      "Hi",
					PaymentProviderPaymentID: "",
					FailureReason:            models.FailureReason{},
				},
			},
			{
				Name: "invalid failed payments: missing failure reason",
				Event: models.PaymentProviderEvent{
					CreatedOn:                time.Time{},
					Type:                     models.Failure,
					PaymentInstruction:       models.PaymentInstruction{},
					PaymentProviderName:      "bc",
					PaymentProviderPaymentID: "123",
					FailureReason: models.FailureReason{
						Code:    "",
						Message: "",
					},
				},
			},
			{
				Name: "invalid failed payments: missing provider name",
				Event: models.PaymentProviderEvent{
					CreatedOn:                time.Time{},
					Type:                     models.Failure,
					PaymentInstruction:       models.PaymentInstruction{},
					PaymentProviderName:      "",
					PaymentProviderPaymentID: "123",
					FailureReason: models.FailureReason{
						Code:    "xxx",
						Message: "",
					},
				},
			},
		}

		for _, testcase := range cases {
			t.Run(testcase.Name, func(t *testing.T) {
				if validation.ValidatePaymentProviderEvent(testcase.Event) == nil {
					t.Errorf("expected an error for %v, but didnt get one", testcase.Event)
				}
			})
		}
	})

	t.Run("unknown event types", func(t *testing.T) {
		is := is.New(t)
		event := models.PaymentProviderEvent{
			CreatedOn:                time.Time{},
			Type:                     models.PaymentProviderEventType("Bob"),
			PaymentInstruction:       models.PaymentInstruction{},
			PaymentProviderName:      "BC",
			PaymentProviderPaymentID: "123",
			FailureReason: models.FailureReason{
				Code:    "xxx",
				Message: "",
			},
		}
		err := validation.ValidatePaymentProviderEvent(event)
		is.True(err != nil)
		invalidTypeErr, isInvalidTypeErr := err.(validation.PaymentProviderInvalidEventTypeError)
		is.True(isInvalidTypeErr)
		is.Equal(invalidTypeErr.Event, event)
	})
}
