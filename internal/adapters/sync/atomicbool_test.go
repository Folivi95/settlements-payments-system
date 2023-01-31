package sync_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/sync"
)

func TestAtomicBool_IsSet(t *testing.T) {
	t.Run("should be true", func(t *testing.T) {
		b := sync.New()
		b.Set()
		assert.True(t, b.IsSet())
	})

	t.Run("should be false", func(t *testing.T) {
		b := sync.New()
		b.UnSet()
		assert.False(t, b.IsSet())
	})
}
