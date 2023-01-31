package zap

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap/zapcore"
)

// NewPrometheusZapHook returns a zap core hook that increments a metric everytime
// an error message is logged.
func NewPrometheusZapHook() func(entry zapcore.Entry) error {
	m := promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_error_log_entries",
		Help: "Counter for the number of error logs that we write.",
	})

	return func(entry zapcore.Entry) error {
		if entry.Level == zapcore.ErrorLevel {
			m.Inc()
		}

		return nil
	}
}
