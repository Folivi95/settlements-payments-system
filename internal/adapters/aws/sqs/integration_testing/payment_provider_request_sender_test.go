//go:build integration
// +build integration

package integration_testing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	. "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/payment_provider_request_sender"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"

	"github.com/google/uuid"
	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/cmd/config"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"

	. "github.com/saltpay/settlements-payments-system/internal/adapters/aws"
	. "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
)

func TestPaymentProviderRequestSenderIntegration(t *testing.T) {
	var (
		ctx = context.Background()
		is  = is.New(t)
	)

	appConfig, err := config.LoadConfig(ctx)
	is.NoErr(err)

	awsSession, err := CreateAwsSession(true, appConfig.AwsRegion, appConfig.AwsEndpoint)
	is.NoErr(err)

	t.Run("the payment request is converted to JSON and added to the unprocessed BC payments SQS queue", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)
		//--- arrange
		paymentReq, _, err := testhelpers.ValidPaymentInstruction()
		is.NoErr(err)

		// set up clients and payment provider request sender
		bcUnprocessedQueueName := fmt.Sprintf("int-tests-banking-circle-unprocessed-reqs-%s", uuid.NewString())
		is.NoErr(CreateQueue(ctx, awsSession, bcUnprocessedQueueName))

		bcSqsUnprocessedReqsClient, err := NewQueueClient(ctx, NewQueueClientOptions{
			AwsSession:             awsSession,
			QueueName:              bcUnprocessedQueueName,
			ReceiveWaitTimeSeconds: appConfig.SqsDefaultReceiveWaitTimeSeconds,
			VisibilityTimeout:      appConfig.SqsDefaultVisibilityTimeoutSeconds,
			ReceiveBatchSize:       appConfig.SqsDefaultReceiveBatchSize,
			Metrics:                testdoubles.DummyMetricsClient{},
		})
		is.NoErr(err)

		isbUnprocessedQueueName := fmt.Sprintf("int-tests-islandsbanki-unprocessed-reqs-%s", uuid.NewString())
		is.NoErr(CreateQueue(ctx, awsSession, isbUnprocessedQueueName))

		isbSqsUnprocessedReqsClient, err := NewQueueClient(ctx, NewQueueClientOptions{
			AwsSession:             awsSession,
			QueueName:              isbUnprocessedQueueName,
			ReceiveWaitTimeSeconds: appConfig.SqsDefaultReceiveWaitTimeSeconds,
			VisibilityTimeout:      appConfig.SqsDefaultVisibilityTimeoutSeconds,
			ReceiveBatchSize:       appConfig.SqsDefaultReceiveBatchSize,
			Metrics:                testdoubles.DummyMetricsClient{},
		})
		is.NoErr(err)

		defer func() {
			err := bcSqsUnprocessedReqsClient.DeleteQueue(ctx)
			if err != nil {
				log := fmt.Sprintf("Failed to cleanup after test. Could not delete temp queue %s. Error: %v", bcUnprocessedQueueName, err)
				fmt.Println(log)
			}
			err = isbSqsUnprocessedReqsClient.DeleteQueue(ctx)
			if err != nil {
				log := fmt.Sprintf("Failed to cleanup after test. Could not delete temp queue %s. Error: %v", isbUnprocessedQueueName, err)
				fmt.Println(log)
			}
		}()

		paymentProviderRequestSender := NewPaymentProviderRequestSender(&bcSqsUnprocessedReqsClient, nil)

		// Delete all messages in the queue
		is.NoErr(bcSqsUnprocessedReqsClient.Purge(ctx))
		is.NoErr(isbSqsUnprocessedReqsClient.Purge(ctx))
		is.NoErr(WaitForQueueToBeEmpty(ctx, awsSession, bcUnprocessedQueueName))
		is.NoErr(WaitForQueueToBeEmpty(ctx, awsSession, isbUnprocessedQueueName))

		//--- act
		is.NoErr(paymentProviderRequestSender.SendPaymentInstruction(ctx, paymentReq))

		//--- assert
		// Poll client to confirm that message was sent to corresponding queue
		bcMsgs, err := bcSqsUnprocessedReqsClient.GetMessages(ctx)
		is.NoErr(err)
		is.Equal(len(bcMsgs.Messages), 1)

		isbMsgs, err := isbSqsUnprocessedReqsClient.GetMessages(ctx)
		is.NoErr(err)
		is.Equal(len(isbMsgs.Messages), 0)

		paymentReqJSON, err := json.Marshal(paymentReq)
		is.NoErr(err)

		is.Equal(*bcMsgs.Messages[0].Body, string(paymentReqJSON))
	})

	t.Run("Peek messages in the DLQ", func(t *testing.T) {
		//--- arrange (add more than 10 payments)
		const (
			amountOfPayments = 50
		)

		var (
			is                              = is.New(t)
			ctx                             = context.Background()
			expectedPaymentInstructionSlice = make([]models.PaymentInstruction, 0, amountOfPayments)
		)

		for i := 0; i < amountOfPayments; i++ {
			expectedPaymentInstructionSlice = append(expectedPaymentInstructionSlice, testhelpers.NewPaymentInstructionBuilder().Build())
		}

		// set up clients and payment provider request sender
		bcDlqName := "bc-unprocessed-payments-deadletter"
		is.NoErr(CreateQueue(ctx, awsSession, bcDlqName))

		bcQueueClient, err := NewQueueClient(ctx, NewQueueClientOptions{
			AwsSession:             awsSession,
			QueueName:              bcDlqName,
			ReceiveWaitTimeSeconds: appConfig.SqsDefaultReceiveWaitTimeSeconds,
			VisibilityTimeout:      3,
			ReceiveBatchSize:       appConfig.SqsDefaultReceiveBatchSize,
			Metrics:                testdoubles.DummyMetricsClient{},
		})
		is.NoErr(err)

		isbDlqName := "isb-unprocessed-payments-deadletter"
		is.NoErr(CreateQueue(ctx, awsSession, isbDlqName))

		defer func() {
			err := bcQueueClient.DeleteQueue(ctx)
			t.Logf("Failed to cleanup after test. Could not delete temp queue %s. Error: %s\n", bcDlqName, err)
			is.NoErr(err)
			if err != nil {
				t.Logf("Failed to cleanup after test. Could not delete temp queue %s. Error: %v", bcDlqName, err)
			}
		}()

		paymentProviderRequestSender := NewPaymentProviderRequestSender(&bcQueueClient, nil)

		is.NoErr(bcQueueClient.Purge(ctx))
		is.NoErr(WaitForQueueToBeEmpty(ctx, awsSession, bcDlqName))

		//--- act
		for i := 0; i < amountOfPayments; i++ {
			is.NoErr(paymentProviderRequestSender.SendPaymentInstruction(ctx, expectedPaymentInstructionSlice[i]))
		}

		dlqInformation, err := bcQueueClient.PeekAllMessages(ctx)
		is.NoErr(err)
		is.Equal(dlqInformation.Count, amountOfPayments)

		paymentInstructionJSON, err := json.Marshal(expectedPaymentInstructionSlice[0])
		is.NoErr(err)
		is.Equal(dlqInformation.Messages[0], string(paymentInstructionJSON))

		time.Sleep(4 * time.Second) // TODO: is there a better way to do this?
		messages, err := bcQueueClient.PeekAllMessages(ctx)
		is.NoErr(err)
		is.Equal(messages.Count, amountOfPayments)
		is.Equal(messages.Count, len(messages.Messages))
	})
}
