//go:generate moq -out mocks/file_store_moq.go -pkg=mocks . FileStore

package ufx_file_listener

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	sqsInternal "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/replay_payment"
	"github.com/saltpay/settlements-payments-system/internal/adapters/sync"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type FileStore interface {
	ReadFile(ctx context.Context, filename string) (io.Reader, error)
}

// UfxFileListener listens to an SQS queue for an event that a Way4 UFX payment instruction file is available.
// Then it fetches that file from S3, creates PaymentInstruction objects from it and invokes the MakeBankingCirclePayment use case on each one.
// Once all invocations have succeeded, it deletes the message from the queue to mark it as processed.
type UfxFileListener struct {
	makePayment                          ports.MakePayment
	checkPaymentAccountFundsAvailability ports.CheckPaymentAccountFundsAvailability
	ufxConverter                         UfxConverter
	fileStoreClient                      FileStore
	ufxQueueClient                       sqsInternal.Queue
	dlqClient                            sqsInternal.Queue
	ufxProcessingMaxMessages             int64
	metricsClient                        ports.MetricsClient
	shouldListen                         *sync.AtomicBool
	featureFlagSvc                       ports.FeatureFlagService
}

func New(
	makePayment ports.MakePayment,
	checkPaymentAccountFundsAvailability ports.CheckPaymentAccountFundsAvailability,
	ufxConverter UfxConverter,
	s3Client FileStore,
	sqsClient sqsInternal.Queue,
	dlqClient sqsInternal.Queue,
	ufxProcessingMaxMessages int64,
	metricsClient ports.MetricsClient,
	featureFlagSvc ports.FeatureFlagService,
) UfxFileListener {
	shouldListen := sync.New()
	shouldListen.Set()
	return UfxFileListener{
		makePayment:                          makePayment,
		checkPaymentAccountFundsAvailability: checkPaymentAccountFundsAvailability,
		ufxConverter:                         ufxConverter,
		fileStoreClient:                      s3Client,
		ufxQueueClient:                       sqsClient,
		dlqClient:                            dlqClient,
		ufxProcessingMaxMessages:             ufxProcessingMaxMessages,
		metricsClient:                        metricsClient,
		shouldListen:                         shouldListen,
		featureFlagSvc:                       featureFlagSvc,
	}
}

const (
	clientStorageMetricName        = "app_storage_client_resp_time_ms"
	instructionsReceivedMetricName = "app_payment_instructions_received"
	useCaseExecutionMetricName     = "app_use_case_execution_time_ms"
)

func (ufl *UfxFileListener) Listen(ctx context.Context) {
	for {
		if !ufl.shouldListen.IsSet() {
			break
		}

		if !ufl.featureFlagSvc.IsIngestionEnabledFromUfxFileNotificationQueue() {
			zapctx.Error(ctx, "Ufx file notification queue ingestion is not enabled")
			time.Sleep(5 * time.Second)
			continue
		}

		receiveMessageOutput, err := ufl.ufxQueueClient.GetMessages(ctx)
		if err != nil {
			zapctx.Error(ctx, "error getting messages from the ufx queue", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		for _, newS3FileMessage := range receiveMessageOutput.Messages {
			filename, err := ufl.extractFileNameFromSqsMessage(ctx, newS3FileMessage)
			if err != nil {
				zapctx.Error(ctx, "error reading filename from message", zap.Error(err))
				continue
			}
			if filename == "" {
				ufl.dlq(ctx, newS3FileMessage)
				zapctx.Debug(ctx, "skipping message because filename picked from SQS queue is empty")
				continue
			}
			ufl.handleNewS3FileMessage(ctx, newS3FileMessage)
		}
	}
}

func (ufl *UfxFileListener) handleNewS3FileMessage(ctx context.Context, message *sqs.Message) {
	ufxContents, fileName, err := ufl.getFileFromS3(ctx, message)
	if err != nil {
		zapctx.Error(ctx, "", zap.Error(err))
		return
	}

	incomingInstructions, err := ufl.convertUfxToIncomingInstructions(ctx, ufxContents, fileName)
	if err != nil {
		ufl.dlq(ctx, message)
		zapctx.Error(ctx, "error converting ufx from file to PaymentInstructions",
			zap.String("file_name", fileName),
			zap.Error(err),
		)
		return
	}

	summary, err := incomingInstructions.SumByCurrency()
	if err != nil {
		zapctx.Error(ctx, "error summing payments by currency",
			zap.Error(err),
		)
		return
	}

	validPaymentInstructions := ufl.validateBalances(ctx, incomingInstructions, summary, fileName)

	executionErrors := ufl.makeAllPayments(ctx, validPaymentInstructions)
	ufl.handleUseCaseExecutionErrors(ctx, executionErrors)

	if err := ufl.ufxQueueClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
		zapctx.Error(ctx, "error deleting message for s3 file from ufx queue",
			zap.String("file_name", fileName),
			zap.Error(err),
		)
		return
	}

	zapctx.Info(ctx, "successfully deleted sqs message for s3 file", zap.String("file_name", fileName))
}

func (ufl *UfxFileListener) getFileFromS3(ctx context.Context, message *sqs.Message) (io.Reader, string, error) {
	fileName, err := ufl.extractFileNameFromSqsMessage(ctx, message)
	if err != nil {
		ufl.dlq(ctx, message)
		return nil, "", fmt.Errorf("[UfxFileListener] Error reading S3 file name from SQS message. Err: %s", err)
	}

	ufxFileContentsReader, err := ufl.fetchFileFromS3(ctx, fileName)
	if err != nil {
		ufl.dlq(ctx, message)
		return nil, "", fmt.Errorf("[UfxFileListener] Error fetching S3 file %s. Err: %s", fileName, err)
	}

	return ufxFileContentsReader, fileName, nil
}

func (ufl *UfxFileListener) dlq(ctx context.Context, message *sqs.Message) {
	if err := ufl.ufxQueueClient.DeleteMessage(ctx, *message.ReceiptHandle); err != nil {
		zapctx.Error(ctx, "error deleting message",
			zap.Error(err),
		)
	}
	if err := ufl.dlqClient.SendMessage(ctx, *message.Body); err != nil {
		zapctx.Error(ctx, "error sending message to DLQ",
			zap.Error(err),
		)
	}
}

func (ufl *UfxFileListener) extractFileNameFromSqsMessage(ctx context.Context, message *sqs.Message) (string, error) {
	type (
		messageBody struct {
			Records []struct {
				S3 struct {
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				} `json:"s3"`
			} `json:"Records"`
		}
	)

	var msgBody messageBody
	err := json.Unmarshal([]byte(*message.Body), &msgBody)
	if err != nil {
		return "", errors.Wrap(err, "[UfxFileListener] (processMessages) failed to parse message body")
	}
	if len(msgBody.Records) == 0 {
		return "", errors.Wrap(err, "[UfxFileListener] (processMessages) no records found in message")
	}
	fileName := msgBody.Records[0].S3.Object.Key
	zapctx.Info(ctx, "flow_step #3: received new UFX file event over SQS",
		zap.String("file_name", fileName),
	)
	return fileName, nil
}

func (ufl *UfxFileListener) fetchFileFromS3(ctx context.Context, fileName string) (io.Reader, error) {
	startTime1 := time.Now()
	ufxFileContentsReader, err := ufl.fileStoreClient.ReadFile(ctx, fileName)
	if err != nil {
		return nil, errors.Wrap(err, "[UfxFileListener] (processMessages) error reading file from bucket")
	}
	responseTime := time.Since(startTime1)
	responseTimeMs := float64(responseTime.Milliseconds())
	tags := []string{"s3-ufx-file-bucket", "fetch-ufx-file"}
	ufl.metricsClient.Histogram(ctx, clientStorageMetricName, responseTimeMs, tags)

	zapctx.Info(ctx, "flow_step #4: fetched UFX file from S3",
		zap.String("file_name", fileName),
		zap.Duration("duration", responseTime),
	)
	return ufxFileContentsReader, nil
}

func (ufl *UfxFileListener) convertUfxToIncomingInstructions(ctx context.Context, ufxContents io.Reader, ufxFileName string) (models.IncomingInstructions, error) {
	startTime := time.Now()
	incomingInstructions, err := ufl.ufxConverter.ConvertUfx(ctx, ufxContents, ufxFileName)
	if err != nil {
		return nil, errors.Wrapf(err, "[UfxFileListener] (processMessages) error converting UFX file %s", ufxFileName)
	}
	elapsed := time.Since(startTime)
	zapctx.Info(ctx, "flow_step #5: payment instructions created from UFX file",
		zap.Int("total", len(incomingInstructions)),
		zap.String("file_name", ufxFileName),
		zap.Duration("duration", elapsed),
	)
	ufl.metricsClient.Count(ctx, instructionsReceivedMetricName, int64(len(incomingInstructions)), []string{"way4_ufx"})
	return incomingInstructions, nil
}

func (ufl *UfxFileListener) makeAllPayments(ctx context.Context, incomingInstructions models.IncomingInstructions) []MakePaymentExecutionError {
	useCaseExecutionErrors := make([]MakePaymentExecutionError, 0, len(incomingInstructions))

	zapctx.Info(ctx, "starting loop to call makePayment usecase")
	batchStartTime := time.Now()
	for _, incomingInstruction := range incomingInstructions {
		piStartTime := time.Now()
		_, err := ufl.makePayment.Execute(ctx, incomingInstruction)
		ufl.metricsClient.Histogram(ctx, useCaseExecutionMetricName, float64(time.Since(piStartTime).Milliseconds()), []string{"make_payment"})
		if err != nil {
			useCaseExecutionErrors = append(useCaseExecutionErrors, MakePaymentExecutionError{
				IncomingInstruction: incomingInstruction,
				Error:               err,
			})
			continue
		}
	}
	elapsedLoopTime := time.Since(batchStartTime)
	zapctx.Info(ctx, "loop to call makePayment usecase for payment instructions ended", // took %v. %d use case executions failed")
		zap.Int("total", len(incomingInstructions)),
		zap.Duration("duration", elapsedLoopTime),
		zap.Int("failed", len(useCaseExecutionErrors)),
	)
	return useCaseExecutionErrors
}

type MakePaymentExecutionError struct {
	IncomingInstruction models.IncomingInstruction
	Error               error
}

func (ufl *UfxFileListener) handleUseCaseExecutionErrors(ctx context.Context, invocationErrors []MakePaymentExecutionError) {
	if len(invocationErrors) > 0 {
		zapctx.Error(ctx, "payment instructions failed when trying to call the make payment usecase",
			zap.Int("failed", len(invocationErrors)),
			zap.Any("errors", invocationErrors),
		)
	}
}

func (ufl *UfxFileListener) validateBalances(ctx context.Context, instructions models.IncomingInstructions, sumByCurrency []models.IncomingInstructionsSummary, filename string) models.IncomingInstructions {
	for _, sum := range sumByCurrency {
		if sum.CurrencyCode == models.ISK {
			continue
		}
		hasBalance, err := ufl.checkPaymentAccountFundsAvailability.Execute(ctx, sum.CurrencyCode, sum.Amount, sum.HighRisk)
		if err != nil {
			// Keep processing other currencies, if account for currency X is not found or has some other problem,
			// remove it from the instructions slice.
			zapctx.Error(ctx, "[UfxFileListener] Error checking balance for currency", zap.Error(err))
		}

		if !hasBalance {
			instructions = instructions.FilterOutCurrency(sum.CurrencyCode)
		}

		if !hasBalance && err == nil {
			// Not enough funds error, generate replayPayment endpoint
			zapctx.Error(ctx, "replay payment for missing funds",
				zap.String("method", "POST"),
				zap.String("path", fmt.Sprintf("/replay-payment?action=%s&currency=%s&file=%s\\", replay_payment.PayCurrencyFromFile, sum.CurrencyCode, filename)),
			)
		}
	}

	return instructions
}

func (ufl *UfxFileListener) StopListening() {
	ufl.shouldListen.UnSet()
}
