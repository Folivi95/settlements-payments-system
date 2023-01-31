package payment_provider_event_listener

import (
	"context"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type loggingAndMetricsObservability struct {
	metricsClient ports.MetricsClient
}

func (l *loggingAndMetricsObservability) QueueProblem(ctx context.Context, action string, err error) {
	zapctx.Error(ctx, "could not perform action on queue", zap.String("action", action), zap.Error(err))
}

func (l *loggingAndMetricsObservability) GotBadMessage(ctx context.Context, message string, err error) {
	zapctx.Error(ctx, "could not parse status update from queue", zap.String("message", message), zap.Error(err))
}

func (l *loggingAndMetricsObservability) ExecutedUseCase(models.PaymentProviderEvent) {
	// todo: metric here?
}

func (l *loggingAndMetricsObservability) ProcessedMessage() {
	// todo: metric here?
}

func (l *loggingAndMetricsObservability) DebugLog(string) {
	// todo: metric here?
}

func (l *loggingAndMetricsObservability) ReceivedMessage() {
	// todo: metric here?
}
