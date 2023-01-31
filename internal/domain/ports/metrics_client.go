//go:generate moq -out mocks/metrics_client_moq.go -pkg=mocks . MetricsClient

package ports

import "context"

// MetricsClient sends various kinds of metrics to monitoring tools.
type MetricsClient interface {
	Histogram(ctx context.Context, name string, value float64, tags []string)
	Count(ctx context.Context, name string, value int64, tags []string)
}

const (
	MetricPaymentProcessingTimeInSeconds string = "app_payment_processing_time_sec"
)
