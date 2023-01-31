//go:build unit
// +build unit

package ufx_file_listener_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/matryer/is"

	ufl "github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestUfxToPaymentInstructionsConverter_ConvertUfx(t *testing.T) {
	t.Run("handles Icelandic IBAN correctly", func(tt *testing.T) {
		ctx := context.Background()

		// Arrange
		// Given a valid UFX file
		// And an instruction (I1) with Valid SWIFT and Empty IBAN
		// And an instruction (I2) with Invalid SWIFT and Empty IBAN
		// And an instruction (I3) with Valid SWIFT and Valid IBAN
		// And an instruction (I4) with Invalid SWIFT and Valid IBAN

		fileBytes, _ := ioutil.ReadFile("test-ufx-is.xml")
		fileReader := bytes.NewReader(fileBytes)

		ufxFileConverter := ufl.NewUfxToPaymentInstructionsConverter()

		// When we convert the UFX file to payment instructions
		incomingInstructions, err := ufxFileConverter.ConvertUfx(ctx, fileReader, "")

		// Then no error should happen
		assert.NoError(tt, err)

		// And 4 instructions should be parsed
		assert.Len(tt, incomingInstructions, 4)

		// And I1 should have a generated IBAN number as the AccountNumber
		assert.NotNil(tt, incomingInstructions[0].Merchant)
		assert.NotNil(tt, incomingInstructions[0].Merchant.Account)
		assert.Equal(tt, "IS310500260000004000000000", incomingInstructions[0].Merchant.Account.AccountNumber)

		// And I2 should have an empty AccountNumber (because of empty IBAN and invalid SWIFT)
		assert.NotNil(tt, incomingInstructions[1].Merchant)
		assert.NotNil(tt, incomingInstructions[1].Merchant.Account)
		assert.Equal(tt, "", incomingInstructions[1].Merchant.Account.AccountNumber)

		// And I3 should have the incoming IBAN as the AccountNumber
		assert.NotNil(tt, incomingInstructions[2].Merchant)
		assert.NotNil(tt, incomingInstructions[2].Merchant.Account)
		assert.Equal(tt, "IS140159260076545510730339", incomingInstructions[2].Merchant.Account.AccountNumber)

		// And I4 should have the incoming IBAN as the AccountNumber
		assert.NotNil(tt, incomingInstructions[3].Merchant)
		assert.NotNil(tt, incomingInstructions[3].Merchant.Account)
		assert.Equal(tt, "IS140159260076545510730339", incomingInstructions[3].Merchant.Account.AccountNumber)
	})
	t.Run("converts a valid UFX file", func(tt *testing.T) {
		ctx := context.Background()

		// Setup
		fileBytes, _ := ioutil.ReadFile("test-ufx.xml")
		fileReader := bytes.NewReader(fileBytes)
		testFileName := "testfilename"

		ufxFileConverter := ufl.UfxToPaymentInstructionsConverter{}

		expectedExecutionDate, _ := time.Parse("2006-01-02", "2021-06-30")
		expectedMerchant := models.Merchant{
			ContractNumber: "9000000",
			RegNumber:      "4000000000",
			Name:           "Ms Big Shot Merchant",
			Email:          "testemail@testmerchant.is",
			Address: models.Address{
				City:         "TestCity",
				AddressLine1: "Testaddress 9",
				AddressLine2: "240 TestCity",
				Country:      "GBR",
			},
			Account: models.Account{
				AccountNumber:        "GB33BUKB20201555555555",
				Swift:                "GIBAHUHB",
				Country:              "HU",
				SwiftReferenceNumber: "",
				BankCountry:          "HUN",
			},
		}
		expectedPayment := models.Payment{
			Sender: models.Sender{},
			Amount: "10",
			Currency: models.Currency{
				IsoNumber: "978",
				IsoCode:   "EUR",
			},
			ExecutionDate: expectedExecutionDate,
		}
		expectedMetaData := models.Metadata{
			Source:   "Way4",
			FileType: "UFX",
			Filename: testFileName,
			Sender:   "SAXO",
		}

		incomingInstructions, err := ufxFileConverter.ConvertUfx(ctx, fileReader, testFileName)

		assert.NoError(tt, err)
		assert.Len(tt, incomingInstructions, 1)
		assert.Equal(tt, expectedMerchant, incomingInstructions[0].Merchant)
		assert.Equal(tt, expectedPayment, incomingInstructions[0].Payment)
		assert.Equal(tt, expectedMetaData, incomingInstructions[0].Metadata)
	})
}

func TestUfxToPaymentInstructionConverter_FilterCurrency(t *testing.T) {
	is := is.New(t)
	t.Run("Filters the XML for a specified currency", func(t *testing.T) {
		originalFileBytes, _ := ioutil.ReadFile("test-ufx-with-unordered-currencies.xml")
		fileReader := bytes.NewReader(originalFileBytes)
		currency := models.HUF

		ufxFileConverter := ufl.UfxToPaymentInstructionsConverter{}

		xmlFile, err := ufxFileConverter.FilterCurrency(fileReader, currency)
		is.NoErr(err)

		expectedHufCurrency := "<Currency>348</Currency>"
		stringifiedFile := string(xmlFile)

		notExpectedGbpCurrency := "<Currency>826</Currency>"

		is.True(isCurrencyInFile(stringifiedFile, expectedHufCurrency))
		is.True(isCurrencyInFile(stringifiedFile, notExpectedGbpCurrency) != true)
	})
}

func isCurrencyInFile(file string, currency string) bool {
	return strings.Contains(file, currency)
}
