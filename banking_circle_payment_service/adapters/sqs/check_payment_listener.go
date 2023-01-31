package sqs

import (
	"context"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	awssqs "github.com/aws/aws-sdk-go/service/sqs"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/sync"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	domainPorts "github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type CheckPaymentStatusListener struct {
	useCasePaymentStatus ports.CheckBankingCirclePaymentStatus
	uncheckedQueueClient sqs.Queue
	dlqClient            sqs.Queue
	featureFlagSvc       domainPorts.FeatureFlagService
	metrics              domainPorts.MetricsClient
	workerPoolSize       int64
	shouldListen         *sync.AtomicBool
}

func NewCheckPaymentStatusListener(
	useCasePaymentStatus ports.CheckBankingCirclePaymentStatus,
	uncheckedQueueClient sqs.Queue,
	dlqClient sqs.Queue,
	featureFlagSvc domainPorts.FeatureFlagService,
	metrics domainPorts.MetricsClient,
	workerPoolSize int64,
) CheckPaymentStatusListener {
	shouldListen := sync.New()
	shouldListen.Set()

	return CheckPaymentStatusListener{
		useCasePaymentStatus: useCasePaymentStatus,
		uncheckedQueueClient: uncheckedQueueClient,
		dlqClient:            dlqClient,
		featureFlagSvc:       featureFlagSvc,
		metrics:              metrics,
		workerPoolSize:       workerPoolSize,
		shouldListen:         shouldListen,
	}
}

// Listen starts long-polling the SQS client, and executes the supplied use case for each message that comes through.
func (cps *CheckPaymentStatusListener) Listen(ctx context.Context) {
	jobs := make(chan *awssqs.Message)
	cps.createWorkerPool(ctx, jobs)

	for {
		if !cps.shouldListen.IsSet() {
			break
		}

		if !cps.featureFlagSvc.IsIngestionEnabledFromBankingCircleUncheckedQueue() {
			continue
		}

		batch, err := cps.uncheckedQueueClient.GetMessages(ctx)
		if err != nil {
			zapctx.Error(ctx, "failed to fetch sqs message", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}

		for _, message := range batch.Messages {
			jobs <- message
		}
	}
}

func (cps *CheckPaymentStatusListener) processMessage(ctx context.Context, msg *awssqs.Message) {
	paymentProviderEvent, err := models.NewPaymentProviderEventFromJSON([]byte(*msg.Body))
	if err != nil {
		zapctx.Error(ctx, "error unmarshalling sqs message body", zap.Error(err))
		cps.dlq(ctx, msg)
		return
	}

	err = cps.useCasePaymentStatus.Execute(ctx, paymentProviderEvent.PaymentInstruction, paymentProviderEvent.PaymentProviderPaymentID, paymentProviderEvent.BankingReference, paymentProviderEvent.CreatedOn)

	if err != nil {
		zapctx.Error(ctx, "error executing the Banking Circle check payment use case for payment instruction",
			zap.String("id", string(paymentProviderEvent.PaymentProviderPaymentID)),
			zap.Error(err),
		)
		cps.dlq(ctx, msg)
		return
	}

	if err := cps.uncheckedQueueClient.DeleteMessage(ctx, *msg.ReceiptHandle); err != nil {
		zapctx.Error(ctx, "error deleting message after executing the Banking Circle check payment use case for payment instruction",
			zap.String("id", string(paymentProviderEvent.PaymentProviderPaymentID)),
			zap.Error(err),
		)
		return
	}
}

func (cps *CheckPaymentStatusListener) createWorkerPool(ctx context.Context, jobs <-chan *awssqs.Message) {
	var w int64
	for w = 1; w <= cps.workerPoolSize; w++ {
		go cps.worker(ctx, jobs)
	}
}

func (cps *CheckPaymentStatusListener) worker(ctx context.Context, messages <-chan *awssqs.Message) {
	for msg := range messages {
		cps.processMessage(ctx, msg)
	}
}

func (cps *CheckPaymentStatusListener) dlq(ctx context.Context, message *awssqs.Message) {
	if cps.dlqClient != nil {
		if err := cps.dlqClient.SendMessage(ctx, *message.Body); err != nil {
			zapctx.Error(ctx, "could not send message to dlq",
				zap.Any("message", *message),
				zap.Error(err),
			)
		}
	}
	if err := cps.uncheckedQueueClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
		zapctx.Error(ctx, "could not delete message",
			zap.Any("message", *message),
			zap.Error(err))
	}
}

func (cps *CheckPaymentStatusListener) StopListening() {
	cps.shouldListen.UnSet()
}
