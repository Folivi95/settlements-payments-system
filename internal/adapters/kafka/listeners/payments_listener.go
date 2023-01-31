//go:generate moq -out internal/mocks/kafka_consumer_mock.go -pkg=mocks . Consumer

package listeners

import (
	"context"

	"github.com/saltpay/go-kafka-driver"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners/internal/dto"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type Consumer interface {
	Listen(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy)
}

type Listener interface {
	Listen(ctx context.Context)
}

type PaymentsListener struct {
	consumer        Consumer
	makePayment     ports.MakePayment
	featureFlagServ ports.FeatureFlagService
}

// NewPaymentsListener creates a new kafka payment listener.
func NewPaymentsListener(
	consumer Consumer,
	makePayment ports.MakePayment,
	featureFlagServ ports.FeatureFlagService,
) *PaymentsListener {
	return &PaymentsListener{
		consumer:        consumer,
		makePayment:     makePayment,
		featureFlagServ: featureFlagServ,
	}
}

// Listen start listening for new incoming messages for the specified consumer. Since its a blocking long-lived task,
// it should live in a separate goroutine.
func (p *PaymentsListener) Listen(ctx context.Context) {
	p.consumer.Listen(ctx, p.processor, kafka.AlwaysCommitWithoutError, p.pauseProcessing)
}

func (p *PaymentsListener) processor(ctx context.Context, message kafka.Message) error {
	zapctx.Debug(ctx, "[PaymentsKafkaListener] processing new message")

	incInstDTO, err := dto.NewIncomingInstructionKafkaDTOFromBytes(message.Value)
	if err != nil {
		zapctx.Error(ctx, "error converting kafka message", zap.Error(err))
		return nil
	}

	// todo map it to service models
	// todo: add a metrics client and count the number payments received with solar source
	incomingInstruction := incInstDTO.MapFromIncomingInstructionKafkaDTO()

	_, err = p.makePayment.Execute(ctx, incomingInstruction)

	// todo: Error report should be sent back to caller
	if err != nil {
		zapctx.Error(ctx, "error making payment",
			zap.String("id", incomingInstruction.PaymentCorrelationId),
			zap.String("account_number", incomingInstruction.AccountNumber()),
			zap.Error(err),
		)
		return nil
	}

	zapctx.Debug(ctx, "[PaymentsKafkaListener] Successfully executed make payment")

	return nil
}

func (p *PaymentsListener) pauseProcessing(ctx context.Context) bool {
	if !p.featureFlagServ.IsKafkaIngestionEnabledForPaymentTransactions() {
		zapctx.Info(ctx, "[PaymentsKafkaListener] kafka ingestion for payment transactions is disabled")

		return true
	}

	return false
}
