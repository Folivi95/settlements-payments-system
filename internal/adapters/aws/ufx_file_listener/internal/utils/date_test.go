//go:build unit
// +build unit

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertToDate(t *testing.T) {
	t.Run("withInvalidDate", func(tt *testing.T) {
		// Act
		actualDate, actualErr := ConvertToDate("invalid date")

		// Assert
		assert.Zero(tt, actualDate)
		assert.EqualError(tt, actualErr, "parsing time \"invalid date\" as \"2006-01-02\": cannot parse \"invalid date\" as \"2006\"")
	})
	t.Run("withValidDate", func(tt *testing.T) {
		// Act
		actualDate, actualErr := ConvertToDate("2020-12-03")

		// Assert
		assert.NoError(tt, actualErr)
		assert.Equal(tt, time.Date(2020, time.December, 3, 0, 0, 0, 0, time.UTC), actualDate)
	})
}
