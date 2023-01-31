//go:build unit
// +build unit

package prometheus_test

import (
	"context"
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/adapters/prometheus"
)

func TestPrometheus_New(t *testing.T) {
	t.Run("should create client without panics due to metric registration", func(t *testing.T) {
		var (
			ctx  = context.Background()
			tags []string
		)

		// due to the way we are registering metrics, prometheus might panic when launching the service
		// this ensures that if a panic happens we fail the build and log the reason.
		// most likely we are reusing a metric name by mistake
		_, _ = prometheus.New(ctx, tags)
		if r := recover(); r != nil {
			t.Fatal("got panic registering prometheus metrics: %w", r)
		}
	})
}
