//go:generate moq -out mocks/validate_payment_instruction.go -pkg=mocks . Validator

package validation

import (
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type Validator interface {
	ValidateIncomingInstruction(incomingInstruction models.IncomingInstruction) IncomingInstructionValidationResult
}

// IncomingInstructionValidator implements the Validator interface and
// exposes only the ValidatePaymentRequest method.
type IncomingInstructionValidator struct{}

func (prv IncomingInstructionValidator) ValidateIncomingInstruction(incomingInstruction models.IncomingInstruction) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(incomingInstruction, "PaymentInstruction")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}

	merchantCheck := validateMerchant(incomingInstruction.Merchant)
	metadataCheck := validateMetadata(incomingInstruction.Metadata)
	paymentCheck := validatePayment(incomingInstruction.Payment)
	senderContractCheck := validateSenderContract(incomingInstruction)

	// TODO: technical-debt, should not depend on Sender at this level
	if models.IsISLSender(incomingInstruction.Metadata.Sender) {
		merchantCheck = validateIslMerchant(incomingInstruction.Merchant)
	}

	return mergeValidationErrors(merchantCheck, metadataCheck, paymentCheck, senderContractCheck)
}

func validateSenderContract(incomingInstruction models.IncomingInstruction) IncomingInstructionValidationResult {
	if models.IsSaxoSender(incomingInstruction.Metadata.Sender) && incomingInstruction.Payment.Currency.IsoCode == models.ISK {
		return Invalid("SAXO files should not contain ISK currencies")
	}
	return Valid
}

func validateAccount(account models.Account) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(account, "Account")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}

	accountNumberCheck := validateAccountNumber(account.AccountNumber)
	return mergeValidationErrors(accountNumberCheck)
}

func validateMerchant(merchant models.Merchant) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(merchant, "Merchant")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}
	contractNumberCheck := validateContractCode(merchant.ContractNumber)
	merchantNameCheck := validateIsNotEmpty(merchant.Name, "Merchant.Name")
	merchantAccountCheck := validateAccount(merchant.Account)

	return mergeValidationErrors(contractNumberCheck, merchantNameCheck, merchantAccountCheck)
}

func validateIslMerchant(merchant models.Merchant) IncomingInstructionValidationResult {
	iban := merchant.Account.AccountNumber
	ibanCheck := ValidateIcelandicIBAN(iban)

	if len(iban) != 26 {
		return ibanCheck
	}

	accountCheck := Valid
	kennitalaCheck := Valid

	accountNumber := iban[4:16]
	kennitala := iban[16:]

	if merchant.Account.Swift != "" && merchant.Account.Swift != accountNumber {
		// validate accountNumber exists in iban
		accountCheck = Invalid("the Swift field that holds account number for IS Merchants is not the same as account number in IBAN")
	}

	if merchant.RegNumber != "" && merchant.RegNumber != kennitala {
		// validate kennitala exists in iban
		kennitalaCheck = Invalid("RegNumber (kennitala) not the same as kennitala in IBAN")
	}

	return mergeValidationErrors(ibanCheck, kennitalaCheck, accountCheck)
}

func validateMetadata(metadata models.Metadata) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(metadata, "Metadata")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}

	sourceCheck := validateIsNotEmpty(metadata.Source, "Metadata.Source")
	return mergeValidationErrors(sourceCheck)
}

func validateCurrency(currency models.Currency) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(currency, "Currency")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}

	isoNumberCheck := validateCurrencyIsoNumber(currency.IsoNumber)
	isoCodeCheck := validateCurrencyIsoCode(currency.IsoCode)
	return mergeValidationErrors(isoNumberCheck, isoCodeCheck)
}

func validatePayment(payment models.Payment) IncomingInstructionValidationResult {
	isEmptyCheck := validateIsNotEmpty(payment, "Payment")
	if !isEmptyCheck.IsValid() {
		return isEmptyCheck
	}

	executionDateCheck := validateExecutionDate(payment.ExecutionDate)
	senderCheck := validateSender(payment.Sender)
	amountCheck := validateIsNotEmpty(payment.Amount, "Payment.Amount")
	currencyCheck := validateCurrency(payment.Currency)

	return mergeValidationErrors(executionDateCheck, senderCheck, amountCheck, currencyCheck)
}

func validateSender(sender models.Sender) IncomingInstructionValidationResult {
	return validateIsEmpty(sender, "Sender")
}
