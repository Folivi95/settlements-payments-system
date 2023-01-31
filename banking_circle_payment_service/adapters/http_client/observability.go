package http_client

import (
	"context"
	"fmt"

	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type bcClientObserver struct {
	metricsClient ports.MetricsClient
}

const (
	clientRequestMetricName    = "app_http_client_request"
	clientResponseMetricName   = "app_http_client_resp_time_ms"
	clientConnectionMetricName = "app_http_client_connection"
)

func (b bcClientObserver) reportResponse(ctx context.Context, responseTime int64, status int, operation string) {
	statusCode := fmt.Sprintf("%dxx", status/100)
	tags := []string{"banking_circle", operation, statusCode}

	b.metricsClient.Histogram(ctx, clientResponseMetricName, float64(responseTime), tags)
	b.metricsClient.Count(ctx, clientRequestMetricName, 1, tags)
}

func (b bcClientObserver) ReusedConnection(ctx context.Context, reused bool) {
	b.metricsClient.Count(ctx, clientConnectionMetricName, 1, []string{fmt.Sprintf("%t", reused)})
}
