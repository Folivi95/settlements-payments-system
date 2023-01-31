//go:generate moq -out mocks/pp_event_validator_moq.go -pkg mocks . PPEventValidator

package validation

import (
	"fmt"

	paymentproviderevents "github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PPEventValidator interface {
	Validate(paymentproviderevents.PaymentProviderEvent) error
}

type PPEventValidatorFunc func(paymentproviderevents.PaymentProviderEvent) error

func (P PPEventValidatorFunc) Validate(event paymentproviderevents.PaymentProviderEvent) error {
	return P(event)
}

func ValidatePaymentProviderEvent(event paymentproviderevents.PaymentProviderEvent) error {
	var missingFields []string

	if event.PaymentProviderName == "" {
		missingFields = append(missingFields, "PaymentProviderName")
	}

	switch event.Type {
	case paymentproviderevents.Processed:
		if event.PaymentProviderPaymentID == "" {
			missingFields = append(missingFields, "PaymentProviderPaymentID")
		}
	case paymentproviderevents.Failure:
		if event.FailureReason.IsEmpty() {
			missingFields = append(missingFields, "FailureReason")
		}
	default:
		return PaymentProviderInvalidEventTypeError{Event: event}
	}

	if len(missingFields) > 0 {
		return PaymentProviderEventValidationError{
			event:         event,
			missingFields: missingFields,
		}
	}

	return nil
}

type PaymentProviderEventValidationError struct {
	event         paymentproviderevents.PaymentProviderEvent
	missingFields []string
}

func (p PaymentProviderEventValidationError) Error() string {
	return fmt.Sprintf("payment provider event is invalid, missing fields: %v, event: %+v", p.missingFields, p.event)
}

type PaymentProviderInvalidEventTypeError struct {
	Event paymentproviderevents.PaymentProviderEvent
}

func (p PaymentProviderInvalidEventTypeError) Error() string {
	return fmt.Sprintf("event type %q is not a valid event type for event %+v", p.Event.Type, p.Event)
}
