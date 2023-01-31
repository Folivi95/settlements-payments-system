//go:generate moq -out mocks/kafka_producer.go -pkg=mocks . Producer

package producers

import (
	"context"
	"encoding/json"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/go-kafka-driver"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type (
	Producer interface {
		WriteMessage(ctx context.Context, message kafka.Message) error
		Close()
	}

	PaymentStatusExporter struct {
		producer     Producer
		featureFlags ports.FeatureFlagService
	}
)

func NewPaymentStatusExporter(featureFlags ports.FeatureFlagService, producer Producer) *PaymentStatusExporter {
	return &PaymentStatusExporter{
		producer:     producer,
		featureFlags: featureFlags,
	}
}

func (p *PaymentStatusExporter) ReportPaymentStatus(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
	if !p.featureFlags.IsKafkaPublishingEnableForAcquiringHostTransactions() {
		zapctx.Warn(ctx, "kafka message publishing for acquiring-settlements-service disabled")
		return nil
	}

	msg, err := json.Marshal(ppEvent)
	if err != nil {
		zapctx.Error(ctx, "error while marshaling payment provider event to bytes", zap.Error(err))
		return err
	}

	err = p.producer.WriteMessage(ctx, kafka.Message{Value: msg})
	if err != nil {
		zapctx.Error(ctx, "error writing message to acquiring settlements service kafka", zap.Error(err))
		return err
	}

	zapctx.Info(ctx, "message to acquiring-settlements-service sent")

	return nil
}

func (p *PaymentStatusExporter) Close() {
	p.producer.Close()
}
