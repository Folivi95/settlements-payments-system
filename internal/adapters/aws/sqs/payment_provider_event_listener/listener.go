package payment_provider_event_listener

import (
	"context"

	awssqs "github.com/aws/aws-sdk-go/service/sqs"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/sync"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type PaymentProviderEventListener struct {
	incomingSQSClient   sqs.Queue
	dlqClient           sqs.Queue
	trackPaymentOutcome ports.TrackPaymentOutcome
	observer            PaymentUpdateObservability
	shouldListen        *sync.AtomicBool
}

type PaymentUpdateObservability interface {
	ReceivedMessage()
	ExecutedUseCase(event models.PaymentProviderEvent)
	ProcessedMessage()
	GotBadMessage(ctx context.Context, goString string, err error)
	DebugLog(message string)
	QueueProblem(ctx context.Context, action string, err error)
}

func New(
	sqsClient sqs.Queue,
	dlqClient sqs.Queue,
	trackPaymentOutcome ports.TrackPaymentOutcome,
	metricsClient ports.MetricsClient,
) *PaymentProviderEventListener {
	shouldListen := sync.New()
	shouldListen.Set()

	return &PaymentProviderEventListener{
		incomingSQSClient:   sqsClient,
		dlqClient:           dlqClient,
		trackPaymentOutcome: trackPaymentOutcome,
		observer: &loggingAndMetricsObservability{
			metricsClient: metricsClient,
		},
		shouldListen: shouldListen,
	}
}

func (p *PaymentProviderEventListener) Listen(ctx context.Context) {
	for {
		if !p.shouldListen.IsSet() {
			break
		}

		rmo, err := p.incomingSQSClient.GetMessages(ctx)
		if err != nil {
			zapctx.Debug(ctx, "error getting messages from the ufx queue", zap.Error(err))

			continue
		}

		// Introduce concurrency to process more than one payment
		p.processMessages(ctx, rmo)
	}
}

func (p *PaymentProviderEventListener) processMessages(ctx context.Context, rmo *awssqs.ReceiveMessageOutput) {
	for _, message := range rmo.Messages {
		p.observer.ReceivedMessage()
		paymentProviderEvent, err := models.NewPaymentProviderEventFromJSON([]byte(*message.Body))
		if err != nil {
			p.observer.GotBadMessage(ctx, message.GoString(), err)
			p.dlq(ctx, message)
			continue
		}

		if err := p.trackPaymentOutcome.Execute(ctx, paymentProviderEvent); err != nil {
			p.dlq(ctx, message)
			continue
		}
		p.observer.ExecutedUseCase(paymentProviderEvent)

		if err := p.incomingSQSClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
			p.observer.QueueProblem(ctx, "deleting message", err)
			continue
		}

		p.observer.ProcessedMessage()
	}
}

func (p *PaymentProviderEventListener) dlq(ctx context.Context, message *awssqs.Message) {
	if err := p.dlqClient.SendMessage(ctx, *message.Body); err != nil {
		p.observer.QueueProblem(ctx, "dlq-ing message", err)
	}
	if err := p.incomingSQSClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
		p.observer.QueueProblem(ctx, "deleting message", err)
	}
}

func (p *PaymentProviderEventListener) StopListening() {
	p.shouldListen.UnSet()
}
