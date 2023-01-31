//go:build unit
// +build unit

package validation_test

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

func TestIncomingInstructionValidator_ValidateIncomingInstruction(t *testing.T) {
	prv := validation.IncomingInstructionValidator{}

	t.Run("valid PI", func(t *testing.T) {
		is := is.New(t)

		validInstruction := testhelpers.NewIncomingInstructionBuilder().Build()

		validationResult := prv.ValidateIncomingInstruction(validInstruction)

		if !validationResult.IsValid() {
			t.Log(validationResult.GetErrors())
		}

		is.True(validationResult.IsValid())
	})

	t.Run("valid RB PI", func(t *testing.T) {
		is := is.New(t)

		validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
			WithMerchantAccountNumber("IS140159260076545510730339").
			// a valid rb instruction can contain empty account number & kennitala given a valid iban
			WithMerchantAccountSwift("").
			WithMerchantRegNumber("").
			WithMetadataSender("RB").
			WithPaymentCurrencyISONumber(models.CurrenciesToIso["ISK"]).
			WithPaymentCurrencyISOCode(models.ISK).
			Build()

		validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

		if !validationResult.IsValid() {
			t.Log(validationResult.GetErrors())
		}

		is.True(validationResult.IsValid())
	})

	t.Run("valid non-ISK Icelandic PI", func(t *testing.T) {
		validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
			WithMerchantAccountNumber("IS140159260076545510730339").
			WithMerchantAccountSwift("").
			WithMerchantRegNumber("").
			WithMetadataSender("ISB_USD").
			WithPaymentCurrencyISONumber(models.CurrenciesToIso["USD"]).
			WithPaymentCurrencyISOCode(models.USD).
			Build()

		validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

		if !validationResult.IsValid() {
			t.Log(validationResult.GetErrors())
		}

		assert.True(t, validationResult.IsValid())
	})

	t.Run("Empty RB merchant Name should be valid", func(t *testing.T) {
		is := is.New(t)

		validInstruction := testhelpers.NewIncomingInstructionBuilder().
			WithMetadataSender("RB").
			WithMerchantAccountNumber("IS140159260076545510730339").
			WithMerchantAccountSwift("").
			WithMerchantRegNumber("").
			WithPaymentCurrencyISOCode(models.ISK).
			WithMerchantName("").
			Build()

		validationResult := prv.ValidateIncomingInstruction(validInstruction)

		if !validationResult.IsValid() {
			t.Log(validationResult.GetErrors())
		}

		is.True(validationResult.IsValid())
	})

	t.Run("invalid PI", func(t *testing.T) {
		t.Run("invalid contract number", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantContractNumber("999").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty merchant name", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantName("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty account number", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty execution date", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Time{}).Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty payment amount", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithPaymentAmount("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty isonumber", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithPaymentCurrencyISONumber("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty isocode", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithPaymentCurrencyISOCode("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("empty meta source", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMetadataSource("").Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assertAValidationError(t, validationResult)
		})

		t.Run("ISK payment in SAXO file", func(t *testing.T) {
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMetadataSender("SAXO").WithPaymentCurrencyISOCode(models.ISK).Build()

			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			assert.NotNil(t, validationResult)
			assert.True(t, validationResult.Failed())
			assert.Len(t, validationResult.GetErrors(), 1)
			assert.Equal(t, "SAXO files should not contain ISK currencies", validationResult.GetErrors()[0])
		})

		t.Run("a few errors", func(t *testing.T) {
			is := is.New(t)

			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantContractNumber("x").WithPaymentCurrencyISOCode("").Build()
			validationResult := prv.ValidateIncomingInstruction(incomingInstruction)

			is.True(validationResult.Failed())
			is.Equal(len(validationResult.GetErrors()), 2)
		})
	})

	t.Run("invalid RB PI", func(t *testing.T) {
		t.Run("Invalid IS IBAN", func(t *testing.T) {
			validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
				WithMerchantAccountNumber("IS140159260045510730339"). // not the correct length
				WithMetadataSender("RB").
				WithPaymentCurrencyISONumber(models.CurrenciesToIso["ISK"]).
				WithPaymentCurrencyISOCode(models.ISK).
				Build()

			validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

			assert.False(t, validationResult.IsValid(), "incorrect iban passed should not take this as a valid instruction")
		})

		t.Run("iban kennitala does not match merchant provided kennitala", func(t *testing.T) {
			validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
				WithMerchantAccountNumber("IS140159260076545510730339").
				WithMerchantAccountSwift("015926007654"). // account number
				WithMerchantRegNumber("5510730333").      // kennitala (not matching with iban ...33 correct would be ...39)
				WithMetadataSender("RB").
				WithPaymentCurrencyISONumber(models.CurrenciesToIso["ISK"]).
				WithPaymentCurrencyISOCode(models.ISK).
				Build()

			expectedError := validation.Invalid("RegNumber (kennitala) not the same as kennitala in IBAN")

			validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

			assert.False(t, validationResult.IsValid(), "given iban does not match with kennitala")
			assert.Len(t, validationResult.GetErrors(), 1, "only a single validation error should occur")
			assert.Equal(t, expectedError.Error(), validationResult.Error(), "validation error must only be kennitala mismatch")
		})

		t.Run("iban account number does not match merchant provided account number", func(t *testing.T) {
			validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
				WithMerchantAccountNumber("IS140159260076545510730339").
				WithMerchantAccountSwift("015927007654"). // account number mismatch
				WithMerchantRegNumber("5510730339").      // kennitala
				WithMetadataSender("RB").
				WithPaymentCurrencyISONumber(models.CurrenciesToIso["ISK"]).
				WithPaymentCurrencyISOCode(models.ISK).
				Build()

			expectedError := validation.Invalid("the Swift field that holds account number for IS Merchants is not the same as account number in IBAN")

			validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

			assert.False(t, validationResult.IsValid(), "given iban does not match with account number")
			assert.Len(t, validationResult.GetErrors(), 1, "only a single validation error should occur")
			assert.Equal(t, expectedError.Error(), validationResult.Error(), "validation error must only be kennitala mismatch")
		})

		t.Run("iban kennitala & account number does not match merchant provided kennitala & account number", func(t *testing.T) {
			validRBInstruction := testhelpers.NewIncomingInstructionBuilder().
				WithMerchantAccountNumber("IS140159260076545510730339").
				WithMerchantAccountSwift("015927007655"). // account number mismatch
				WithMerchantRegNumber("5510730333").      // kennitala mismatch
				WithMetadataSender("RB").
				WithPaymentCurrencyISONumber(models.CurrenciesToIso["ISK"]).
				WithPaymentCurrencyISOCode(models.ISK).
				Build()

			expectedError := validation.Invalid("RegNumber (kennitala) not the same as kennitala in IBAN; the Swift field that holds account number for IS Merchants is not the same as account number in IBAN")

			validationResult := prv.ValidateIncomingInstruction(validRBInstruction)

			assert.False(t, validationResult.IsValid(), "given iban does not match with kennitala & account number")
			assert.Len(t, validationResult.GetErrors(), 2, "both account number & kennitala validation error should occur")
			assert.Equal(t, expectedError.Error(), validationResult.Error(), "validation error must only be kennitala mismatch")
		})
	})
}

func assertAValidationError(t *testing.T, validationResult validation.IncomingInstructionValidationResult) {
	t.Helper()
	is := is.New(t)
	is.True(validationResult.Failed())
	is.Equal(len(validationResult.GetErrors()), 1)
}
