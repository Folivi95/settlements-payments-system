package validation

import (
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func validateCurrencyIsoCode(isoCode models.CurrencyCode) IncomingInstructionValidationResult {
	if !validateIsNotEmpty(isoCode, "Currency.IsoCode").IsValid() {
		return Invalid("Currency.IsoCode should not be empty")
	}
	return Valid
}

func validateCurrencyIsoNumber(isoNumber string) IncomingInstructionValidationResult {
	if !validateIsNotEmpty(isoNumber, "Currency.IsoNumber").IsValid() {
		return Invalid("Currency.IsoNumber should not be empty")
	}
	return Valid
}

func validateContractCode(contractCode string) IncomingInstructionValidationResult {
	if !validateIsNotEmpty(contractCode, "ContractCode").IsValid() {
		return Invalid("ContractCode should not be empty")
	}

	if len(contractCode) < 7 {
		return Invalid("ContractCode should be at least 7 characters")
	}

	return Valid
}

func validateAccountNumber(accountNumber string) IncomingInstructionValidationResult {
	if !validateIsNotEmpty(accountNumber, "AccountNumber").IsValid() {
		return Invalid("AccountNumber should not be empty")
	}

	return Valid
}

func validateExecutionDate(executionDate time.Time) IncomingInstructionValidationResult {
	if !validateIsNotEmpty(executionDate, "Payment.ExecutionDate").IsValid() {
		return Invalid("Payment.ExecutionDate should not be empty")
	}

	return Valid
}
