// Package prometheus provides the required functions to publish service metrics.
// Metrics and tags should follow naming guidelines from prometheus, reference available here:
// https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
package prometheus

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"
)

type MetricsClient struct {
	// globalTags is a list of tag values that must be added to all metrics
	globalTags []string

	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
}

// New inits the metrics and automatically registers the metrics with the default prometheus default register.
// When adding tags keep in mind tag cardinality, prometheus client keeps tags in memory, so you should not use
// any unbounded value in the tag value. tags argument defines global tag values for all metrics.
func New(ctx context.Context, tags []string) (MetricsClient, error) {
	client := MetricsClient{
		// TODO: check what kind of tags are used here and these make sense for prometheus
		globalTags: tags,

		counters: map[string]*prometheus.CounterVec{
			"app_http_client_request": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_http_client_request",
				Help: "Counter for the number of http requests done",
			}, []string{"callee", "operation", "status_code"}),
			"app_http_client_connection": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_http_client_connection",
				Help: "Counter for the number of connections",
			}, []string{"reused"}),
			"app_http_client_request_retry": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_http_client_request_retry",
				Help: "Counter for the number retries",
			}, []string{"callee", "operation"}),
			"app_settlements_provider_check_payment_status": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_check_payment_status",
				Help: "Counter for the number of requests done to the payment status endpoint",
			}, []string{"result", "payment_provider", "currency"}),
			"app_settlements_provider_missing_funds": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_missing_funds",
				Help: "Counter for the number of times have missing funds while trying to execute a payment",
			}, []string{"payment_provider", "currency"}),
			"app_settlements_provider_request_payment": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_request_payment",
				Help: "Counter number of payments requested",
			}, []string{"result", "payment_provider", "currency"}),
			"app_queue_messages_received": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_queue_messages_received",
				Help: "Counter for the number of message received",
			}, []string{"queue_name"}),
			"app_payment_instructions_received": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_payment_instructions_received",
				Help: "Counter for the number of message received",
			}, []string{"source"}),
			"app_payment_instruction_store": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_payment_instruction_store",
				Help: "Counter for the number of gets from the datastore",
			}, []string{"callee", "operation", "result"}),
			"app_settlements_provider_missing_account_funds": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_missing_account_funds",
				Help: "Counter for the number of gets from the datastore",
			}, []string{"currency"}),
			"app_settlements_provider_data_cleaned": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_data_cleaned",
				Help: "Counter for the number of times the data sanitization runs",
			}, []string{"payment_provider"}),
			"app_settlements_provider_payments_processing_disabled": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_payments_processing_disabled",
				Help: "Counter for the number of times the service tries to process payments but feature flag is off",
			}, []string{"payment_provider"}),
			"app_settlements_payments_count": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_payments_count",
				Help: "Counter for the number of payments per currency and payment provider",
			}, []string{"payment_provider", "currency"}),
			"app_settlements_payments_amount": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_payments_amount",
				Help: "Counter for the total value of payments per currency and payment provider",
			}, []string{"payment_provider", "currency"}),
			"app_settlements_provider_missing_account_funds_using_loan": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_missing_account_funds_using_loan",
				Help: "Counter for the total number of loans per currency",
			}, []string{"currency"}),
			"app_settlements_provider_success_payment": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_provider_success_payment",
				Help: "Counter for the number of successful payments",
			}, []string{"payment_provider"}),
			"app_settlements_payments_incorrect_account_number": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_settlements_payments_incorrect_account_number",
				Help: "Counter for the number of times an account number is in an incorrect format",
			}, []string{"currency"}),
			"app_payment_instruction_update": promauto.NewCounterVec(prometheus.CounterOpts{
				Name: "app_payment_instruction_update",
				Help: "Counter for when we update the state of a payment",
			}, []string{"currency"}),
		},
		histograms: map[string]*prometheus.HistogramVec{
			"app_http_client_resp_time_ms": promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "app_http_client_resp_time_ms",
				Help: "HTTP request duration in milliseconds",
			}, []string{"callee", "operation", "status_code"}),
			"app_payment_processing_time_sec": promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "app_payment_processing_time_sec",
				Help: "Payment processing duration in seconds",
			}, []string{"result", "payment_provider"}),
			"app_storage_client_resp_time_ms": promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "app_storage_client_resp_time_ms",
				Help: "S3 file retrieval duration in milliseconds",
			}, []string{"callee", "operation"}),
			"app_use_case_execution_time_ms": promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "app_use_case_execution_time_ms",
				Help: "Operation duration in milliseconds",
			}, []string{"operation"}),
			"app_payment_instruction_store_duration": promauto.NewHistogramVec(prometheus.HistogramOpts{
				Name: "app_payment_instruction_store_duration",
				Help: "Histogram for the duration of loading from the datastore",
			}, []string{"callee", "operation", "result"}),
		},
	}

	client.initMetrics(ctx)

	return client, nil
}

func (m *MetricsClient) Count(ctx context.Context, name string, value int64, tags []string) {
	if counterVec, ok := m.counters[name]; ok {
		c, err := counterVec.GetMetricWithLabelValues(tags...)
		if err != nil {
			zapctx.Info(ctx, "metric not found",
				zap.String("type", "count"),
				zap.String("metric", name),
				zap.Strings("tags", tags),
				zap.Error(err),
			)

			return
		}
		c.Add(float64(value))

		return
	}

	zapctx.Info(ctx, "metric not found",
		zap.String("type", "count"),
		zap.String("metric", name),
	)
}

func (m *MetricsClient) Histogram(ctx context.Context, name string, value float64, tags []string) {
	if histogramVec, ok := m.histograms[name]; ok {
		h, err := histogramVec.GetMetricWithLabelValues(tags...)
		if err != nil {
			zapctx.Info(ctx, "metric not found",
				zap.String("type", "histogram"),
				zap.String("metric", name),
				zap.Strings("tags", tags),
				zap.Error(err),
			)
			return
		}
		h.Observe(value)

		return
	}

	zapctx.Info(ctx, "metric not found",
		zap.String("type", "histogram"),
		zap.String("metric", name),
	)
}
