package sqs

import (
	"context"
	"sync/atomic"
	"time"

	awssqs "github.com/aws/aws-sdk-go/service/sqs"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/sync"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	domainPorts "github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type PaymentInstructionEventListener struct {
	useCasePaymentRequest            ports.MakeBankingCirclePayment
	incomingQueueClient              sqs.Queue
	dlqClient                        sqs.Queue
	featureFlagSvc                   domainPorts.FeatureFlagService
	metrics                          domainPorts.MetricsClient
	workerPoolSize                   int64
	waitTimeBetweenFeatureFlagChecks time.Duration
	delayBetweenProcessingMessages   int64
	shouldListen                     *sync.AtomicBool
	sleepTime                        time.Duration
	numberOfDeleteRetries            int
	numberOfFailedPaymentsThreshold  int
	counter                          int32
}

const (
	waitTimeBetweenFeatureFlagChecks = 20 * time.Second
	paymentsProcessingDisabled       = "app_settlements_provider_payments_processing_disabled"
	paymentProviderTag               = "banking_circle"
)

func NewPaymentInstructionEventListener(
	useCasePaymentRequest ports.MakeBankingCirclePayment,
	incomingQueueClient sqs.Queue,
	dlqClient sqs.Queue,
	featureFlagSvc domainPorts.FeatureFlagService,
	metrics domainPorts.MetricsClient,
	workerPoolSize int64,
	delayBetweenProcessingMessages int64,
	sleepTime time.Duration,
	numberOfDeleteRetries int,
	numberOfFailedPaymentsThreshold int,
) PaymentInstructionEventListener {
	shouldListen := sync.New()
	shouldListen.Set()
	return PaymentInstructionEventListener{
		useCasePaymentRequest:            useCasePaymentRequest,
		incomingQueueClient:              incomingQueueClient,
		dlqClient:                        dlqClient,
		featureFlagSvc:                   featureFlagSvc,
		metrics:                          metrics,
		workerPoolSize:                   workerPoolSize,
		waitTimeBetweenFeatureFlagChecks: waitTimeBetweenFeatureFlagChecks,
		shouldListen:                     shouldListen,
		delayBetweenProcessingMessages:   delayBetweenProcessingMessages,
		sleepTime:                        sleepTime,
		numberOfDeleteRetries:            numberOfDeleteRetries,
		numberOfFailedPaymentsThreshold:  numberOfFailedPaymentsThreshold,
	}
}

// Listen starts long-polling the SQS client, and executes the supplied use case for each message that comes through.
func (pir *PaymentInstructionEventListener) Listen(ctx context.Context) {
	jobs := make(chan *awssqs.Message)
	pir.createWorkerPool(jobs)

	for {
		if !pir.shouldListen.IsSet() {
			break
		}

		if !pir.featureFlagSvc.IsIngestionEnabledFromBankingCircleUnprocessedQueue() {
			zapctx.Error(ctx, "(Listen) payments processing disabled")
			pir.metrics.Count(ctx, paymentsProcessingDisabled, 1, []string{paymentProviderTag})
			time.Sleep(5 * time.Second)
			continue
		}
		batch, err := pir.incomingQueueClient.GetMessages(ctx)
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

func (pir *PaymentInstructionEventListener) processMessage(ctx context.Context, msg *awssqs.Message) {
	paymentInstruction, err := models.NewPaymentInstructionFromJSON([]byte(*msg.Body))
	if err != nil {
		zapctx.Error(ctx, "error unmarshalling sqs message body", zap.Error(err))
		pir.dlq(ctx, msg)
		return
	}
	zapctx.Debug(ctx, "flow_step #8: payment instruction received by the Banking Circle Payment Service",
		zap.String("id", string(paymentInstruction.ID())),
		zap.String("contract_number", paymentInstruction.ContractNumber()),
	)

	_, err = pir.useCasePaymentRequest.Execute(ctx, paymentInstruction)
	if err != nil {
		zapctx.Error(ctx, "error executing the Banking Circle make payment use case", zap.Error(err))
		pir.dlq(ctx, msg)
		if c := atomic.AddInt32(&pir.counter, 1); int(c) >= pir.numberOfFailedPaymentsThreshold {
			zapctx.Info(ctx, "disabling payment ingestion")
			if err = pir.featureFlagSvc.ToggleOffIngestionFromBankingCirclePayments(ctx); err != nil {
				zapctx.Error(ctx, "error disabling feature flag", zap.Error(err))
			}
		}
		return
	}

	for i := 0; i < pir.numberOfDeleteRetries; i++ {
		err := pir.incomingQueueClient.DeleteMessage(ctx, *msg.ReceiptHandle)
		if err == nil {
			return
		}
		zapctx.Error(ctx, "error deleting message from the Banking Circle unprocessed queue", zap.Error(err))
		if _, isRetryable := err.(sqs.RetryableError); !isRetryable {
			break
		}
		zapctx.Error(ctx, "error is retryable", zap.Int("attempt", i))
		time.Sleep(pir.sleepTime)
	}
	zapctx.Error(ctx, "maximum retries reached, could not delete message", zap.Int("max_retries", pir.numberOfDeleteRetries))
}

func (pir *PaymentInstructionEventListener) createWorkerPool(jobs <-chan *awssqs.Message) {
	var w int64
	for w = 1; w <= pir.workerPoolSize; w++ {
		go pir.worker(jobs)
	}
}

func (pir *PaymentInstructionEventListener) worker(messages <-chan *awssqs.Message) {
	for msg := range messages {
		// TODO: replace context.Background
		pir.processMessage(context.Background(), msg)
		time.Sleep(time.Duration(pir.delayBetweenProcessingMessages) * time.Millisecond)
	}
}

func (pir *PaymentInstructionEventListener) dlq(ctx context.Context, message *awssqs.Message) {
	if pir.dlqClient != nil {
		if err := pir.dlqClient.SendMessage(ctx, *message.Body); err != nil {
			zapctx.Error(ctx, "could not sent message to dlq",
				zap.Any("message", message),
				zap.Error(err),
			)
		}
	}
	if err := pir.incomingQueueClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
		zapctx.Error(ctx, "could not delete message",
			zap.Any("message", message),
			zap.Error(err),
		)
	}
}

func (pir *PaymentInstructionEventListener) StopListening() {
	pir.shouldListen.UnSet()
}
