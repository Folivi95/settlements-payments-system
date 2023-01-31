package validation

import (
	"math/big"
	"strings"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func ValidateIcelandicIBAN(iban string) IncomingInstructionValidationResult {
	if len(iban) != 26 {
		return Invalid("iban field for Icelandic account does not contain 26 characters")
	}
	if !strings.HasPrefix(iban, "IS") {
		return Invalid("iban is not Icelandic")
	}

	if !isValidIcelandicChecksum(iban) {
		return Invalid("iban checksum is not valid")
	}

	return Valid
}

func isValidIcelandicChecksum(iban string) bool {
	iban = iban[4:] + iban[:4]
	iban = iban[:22] + models.IsIBANChecksumCode + iban[24:]
	ibanInt := new(big.Int)
	ibanInt.SetString(iban, 10)
	ibanInt.Mod(ibanInt, big.NewInt(97))

	return ibanInt.Int64() == 1
}
