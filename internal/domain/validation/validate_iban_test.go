//go:build unit
// +build unit

package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

func TestValidateIcelandicIBAN(t *testing.T) {
	t.Run("WhenNot26Long", func(t *testing.T) {
		// Given a short iban
		iban := "short"

		// When we validate being a valid icelandic iban
		actualErr := validation.ValidateIcelandicIBAN(iban)

		// Then we should get an error
		assert.EqualError(t, actualErr, validation.Invalid("iban field for Icelandic account does not contain 26 characters").Error())
	})
	t.Run("WhenCountryCodeIsDifferent", func(t *testing.T) {
		// Given an iban with an incorrect country code
		iban := "AA140159260076545510730339"

		// When we validate being a valid icelandic iban
		actualErr := validation.ValidateIcelandicIBAN(iban)

		// Then we should get an error
		assert.EqualError(t, actualErr, validation.Invalid("iban is not Icelandic").Error())
	})
	t.Run("WhenChecksumIsIncorrect", func(t *testing.T) {
		// Given an iban with an incorrect checksum
		iban := "IS130159260076545510730339"

		// When we validate being a valid icelandic iban
		actualErr := validation.ValidateIcelandicIBAN(iban)

		// Then we should get an error
		assert.EqualError(t, actualErr, validation.Invalid("iban checksum is not valid").Error())
	})
	t.Run("WhenIbanIsValid", func(t *testing.T) {
		// Given a valid icelandic iban
		iban := "IS140159260076545510730339"

		// When we validate being a valid icelandic iban
		actualErr := validation.ValidateIcelandicIBAN(iban)

		// Then we should NOT get an error
		assert.True(t, actualErr.IsValid(), "valid IBAN given should not flag this as invalid")
	})
}
