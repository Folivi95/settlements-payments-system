//go:build integration
// +build integration

package integration_testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	sqs2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/sqs"
	"github.com/saltpay/settlements-payments-system/cmd/config"
	aws2 "github.com/saltpay/settlements-payments-system/internal/adapters/aws"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
)

// TODO: Get rid of this and use the proper mock.
type MyMockUseCase struct {
	MsgReceivedNotificationChan              chan interface{}
	paymentInstructionThatWasPassedToExecute models.PaymentInstruction
}

func (m *MyMockUseCase) Execute(ctx context.Context, paymentInstruction models.PaymentInstruction) (models.ProviderPaymentID, error) {
	zapctx.Info(ctx, "======= [Execute] make paymentInstruction", zap.Any("payment_instruction", paymentInstruction))
	m.paymentInstructionThatWasPassedToExecute = paymentInstruction
	m.MsgReceivedNotificationChan <- true
	return "", nil
}

func TestMakePaymentListenerIntegrationTest(t *testing.T) {
	t.Run("Picking up messages from the unprocessed queue and invoking the MakeBankingCirclePayment use case", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)

		appConfig, err := config.LoadConfig(ctx)
		is.NoErr(err)
		awsSession, err := aws2.CreateAwsSession(true, appConfig.AwsRegion, appConfig.AwsEndpoint)
		is.NoErr(err)

		//--- arrange
		// create a mock use case that will let us know when it has been called
		useCaseReceivedMessageChan := make(chan interface{})
		mockMakePaymentUseCase := &MyMockUseCase{MsgReceivedNotificationChan: useCaseReceivedMessageChan}

		// sqs set up
		unprocessedQueueName := fmt.Sprintf("int-tests-banking-circle-unprocessed-reqs-%s", uuid.NewString())
		is.NoErr(sqs.CreateQueue(ctx, awsSession, unprocessedQueueName))

		sqsClient, err := sqs.NewQueueClient(ctx, sqs.NewQueueClientOptions{
			AwsSession:             awsSession,
			QueueName:              unprocessedQueueName,
			ReceiveWaitTimeSeconds: appConfig.SqsDefaultReceiveWaitTimeSeconds,
			VisibilityTimeout:      appConfig.SqsDefaultVisibilityTimeoutSeconds,
			ReceiveBatchSize:       appConfig.SqsDefaultReceiveBatchSize,
			Metrics:                testdoubles.DummyMetricsClient{},
		})
		is.NoErr(err)

		defer func() {
			if err := sqsClient.DeleteQueue(ctx); err != nil {
				zapctx.Info(ctx, "failed to cleanup after test. Could not delete temp queue",
					zap.String("queue_name", unprocessedQueueName),
					zap.Error(err),
				)
			}
		}()

		// object under test
		paymentInstructionReceiver := sqs2.NewPaymentInstructionEventListener(
			mockMakePaymentUseCase,
			&sqsClient,
			nil,
			testdoubles.FeatureFlagService{},
			testdoubles.DummyMetricsClient{},
			1,
			0,
			time.Millisecond,
			1,
			5,
		)

		//--- act

		// start listening for messages
		go paymentInstructionReceiver.Listen(ctx)

		// put a message on the queue
		paymentInstruction, paymentInstructionJSON, err := testhelpers.ValidPaymentInstruction()
		is.NoErr(err)
		zapctx.Info(ctx, "[test] about to call sqsClient.SendMessage()")
		err = sqsClient.SendMessage(ctx, paymentInstructionJSON)
		zapctx.Info(ctx, "[test] finished call to sqsClient.SendMessage()")
		is.NoErr(err)

		//--- assert
		// wait for the use case to tell us it's been called
		select {
		case <-useCaseReceivedMessageChan:
			is.Equal(mockMakePaymentUseCase.paymentInstructionThatWasPassedToExecute, paymentInstruction)
			is.NoErr(sqs.ExpectNoMessagesInQueueEventually(ctx, awsSession, unprocessedQueueName, 15*time.Second))

		case <-time.After(30 * time.Second):
			t.Fatal("did not receive message in the use case within 30 seconds")
		}
	})
}
