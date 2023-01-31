//go:build unit
// +build unit

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsIslSender(t *testing.T) {
	// Act & Assert
	assert.True(t, IsISLSender("RB"))
	assert.True(t, IsISLSender("RB_123"))
	assert.True(t, IsISLSender("ISB_USD"))
	assert.True(t, IsISLSender("ISB_GBP"))
	assert.True(t, IsISLSender("ISB_EUR"))
	assert.True(t, IsISLSender("ISB"))
	assert.False(t, IsISLSender("Else"))
}

func TestIsSaxoSender(t *testing.T) {
	// Act & Assert
	assert.True(t, IsSaxoSender("SAXO"))
	assert.True(t, IsSaxoSender("SAXO_HR"))
	assert.False(t, IsSaxoSender("Else"))
}
