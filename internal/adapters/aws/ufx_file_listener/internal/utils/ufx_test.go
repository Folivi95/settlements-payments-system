//go:build unit
// +build unit

package utils

import (
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/internal/ufx"

	"github.com/stretchr/testify/assert"
)

func TestGetParamValue(t *testing.T) {
	// Arrange
	mockParms := []ufx.Parm{
		{ParmCode: "parm1", Value: "value1"},
		{ParmCode: "parm2", Value: "value2"},
		{ParmCode: "parm3", Value: "value3"},
	}

	t.Run("WhenParamDoesNotExist", func(tt *testing.T) {
		// Act
		actualValue := GetParamValue(mockParms, "nonExisting")

		// Assert
		assert.Equal(tt, "", actualValue)
	})
	t.Run("WhenParamDoesExists", func(tt *testing.T) {
		// Act
		actualValue := GetParamValue(mockParms, "parm2")

		// Assert
		assert.Equal(tt, "value2", actualValue)
	})
}
