package utils

import (
	"fmt"
	"math/big"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// see: https://bank.codes/iban/structure/iceland/
func GenerateIcelandicIBAN(kennitala string, accountNumber string) (string, error) {
	if len(kennitala) != 10 {
		return "", fmt.Errorf("kennitala must be 10 characters long")
	}

	if len(accountNumber) != 12 {
		return "", fmt.Errorf("account number must be 12 characters long")
	}

	checkSum := generateIcelandicIBANChecksum(kennitala, accountNumber)

	return "IS" + checkSum + accountNumber + kennitala, nil
}

// see: https://en.wikipedia.org/wiki/International_Bank_Account_Number#Generating_IBAN_check_digits
func generateIcelandicIBANChecksum(kennitala string, accountNumber string) string {
	// Replace the two check digits by 00 (e.g., GB00 for the UK).
	iban := "IS00" + accountNumber + kennitala

	// Move the four initial characters to the end of the string.
	iban = iban[4:] + iban[:4]

	// Replace the letters in the string with digits (IS -> 1828)
	iban = iban[:22] + models.IsIBANChecksumCode + iban[24:]

	// Convert the string to an integer
	ibanInt := new(big.Int)
	ibanInt.SetString(iban, 10)

	// Calculate mod-97
	ibanInt.Mod(ibanInt, big.NewInt(97))

	// Subtract the remainder from 98
	checkSum := ibanInt.Sub(big.NewInt(98), ibanInt)

	return fmt.Sprintf("%02d", checkSum.Int64())
}

// ISO 3166 alpha-2 country code
// Swift Code consists of 8 to 11 characters
// Example: AAAABBCCDDD
// AAAA -> Bank code
// BB -> Alpha2 country code
// CC -> Location code
// DDD -> Branch code, optional.
func ExtractCountryCode(swiftCode string) string {
	if len(swiftCode) < 6 {
		return ""
	}
	return swiftCode[4:6]
}
