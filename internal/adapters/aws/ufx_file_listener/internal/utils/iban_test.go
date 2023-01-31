package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/internal/utils"
)

func TestGenerateIcelandicIBAN(t *testing.T) {
	t.Run("WhenInvalidKennitala", func(tt *testing.T) {
		// Act
		actualIban, actualErr := utils.GenerateIcelandicIBAN("short", "123456789012")

		// Arrange
		assert.Empty(tt, actualIban)
		assert.EqualError(tt, actualErr, "kennitala must be 10 characters long")
	})
	t.Run("WhenInvalidAccount", func(tt *testing.T) {
		// Act
		actualIban, actualErr := utils.GenerateIcelandicIBAN("1234567890", "short")

		// Arrange
		assert.Empty(tt, actualIban)
		assert.EqualError(tt, actualErr, "account number must be 12 characters long")
	})
	t.Run("WhenSuccessful", func(tt *testing.T) {
		// Act
		actualIban, actualErr := utils.GenerateIcelandicIBAN("5510730339", "015926007654")
		expectedIban := "IS140159260076545510730339"

		// Arrange
		assert.NoError(tt, actualErr)
		assert.Equal(tt, expectedIban, actualIban)
	})
	t.Run("WhenSuccessfulWithOneLongChecksum", func(tt *testing.T) {
		// Act
		actualIban, actualErr := utils.GenerateIcelandicIBAN("6903942199", "018626000615")
		expectedIban := "IS040186260006156903942199"

		// Arrange
		assert.NoError(tt, actualErr)
		assert.Equal(tt, expectedIban, actualIban)
	})
}

func TestExtractCountryCode(t *testing.T) {
	t.Run("WhenSwiftTooShort", func(tt *testing.T) {
		// Act
		actualCode := utils.ExtractCountryCode("short")

		// Arrange
		assert.Empty(tt, actualCode)
	})
	t.Run("WhenSuccessful", func(tt *testing.T) {
		// Act
		actualCode := utils.ExtractCountryCode("GIBAHUHB")

		// Arrange
		assert.Equal(tt, "HU", actualCode)
	})
}
