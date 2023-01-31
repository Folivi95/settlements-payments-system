//go:build integration
// +build integration

package integration_testing_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/cmd/config"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/s3"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
)

type MockMakePayment struct {
	MsgReceivedNotificationChan               chan interface{}
	incomingInstructionThatWasPassedToExecute models.IncomingInstruction
}

func (m *MockMakePayment) Execute(_ context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
	fmt.Printf("[Execute] request: %+v;", incomingInstruction)
	m.incomingInstructionThatWasPassedToExecute = incomingInstruction
	m.MsgReceivedNotificationChan <- true
	return "some_id", nil
}

func TestUFXListener(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping UFX File Listener integration tests because short mode was specified")
	}

	t.Run("given a ufx file listener", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)
		appConfig, err := config.LoadConfig(ctx)
		is.NoErr(err)
		awsSession, err := aws.CreateAwsSession(true, appConfig.AwsRegion, appConfig.AwsEndpoint)
		is.NoErr(err)
		ufxFileNotificationQueueName := appConfig.SqsUfxFileNotificationQueueName

		t.Run("a valid ufx file invokes the make payment request with the corresponding incoming instruction", func(t *testing.T) {
			ctx := context.Background()
			ufxAndIncomingInst := testhelpers.ValidUfxAndIncomingInstruction()

			// create a mock make payment that will let us know when it has been called
			useCaseMakePaymentReceivedMessageChan := make(chan interface{})
			mockMakePaymentUseCase := &MockMakePayment{
				MsgReceivedNotificationChan: useCaseMakePaymentReceivedMessageChan,
			}
			mockCheckPaymentAccountFundsAvailability := &mocks.CheckPaymentAccountFundsAvailabilityMock{
				ExecuteFunc: func(ctx context.Context, code models.CurrencyCode, amount float64, highRisk bool) (bool, error) {
					return true, nil
				},
			}

			metricsClient := testdoubles.DummyMetricsClient{}
			featureFlag := testdoubles.FeatureFlagService{}

			// set up clients and ufx file listener
			s3PaymentFilesClient := s3.NewClient(awsSession, appConfig.S3UfxPaymentFilesBucketName)
			sqsInboundQueueClient, err := sqs.NewQueueClient(ctx, sqs.NewQueueClientOptions{
				AwsSession:             awsSession,
				QueueName:              ufxFileNotificationQueueName,
				ReceiveWaitTimeSeconds: appConfig.SqsDefaultReceiveWaitTimeSeconds,
				VisibilityTimeout:      appConfig.SqsDefaultVisibilityTimeoutSeconds,
				ReceiveBatchSize:       appConfig.SqsDefaultReceiveBatchSize,
				Metrics:                metricsClient,
			})
			is.NoErr(err)

			var sqsDLQ sqs.QueueClient

			ufxFileHandler := ufx_file_listener.New(
				mockMakePaymentUseCase,
				mockCheckPaymentAccountFundsAvailability,
				&ufx_file_listener.UfxToPaymentInstructionsConverter{},
				&s3PaymentFilesClient,
				&sqsInboundQueueClient,
				&sqsDLQ,
				appConfig.UfxProcessingMaxMessages,
				metricsClient,
				featureFlag,
			)

			// Delete all messages in the queue
			err = sqsInboundQueueClient.Purge(ctx)
			is.NoErr(err)
			err = sqs.WaitForQueueToBeEmpty(ctx, awsSession, ufxFileNotificationQueueName)
			is.NoErr(err)

			// Start listening for messages
			go ufxFileHandler.Listen(ctx)

			// Add a file to the s3 bucket
			testFile := bytes.NewReader([]byte(ufxAndIncomingInst.UfxFileContents))
			err = s3PaymentFilesClient.PutBucketFile(ctx, testFile, ufxAndIncomingInst.UfxFileName)
			is.NoErr(err)

			// Then

			// wait for the use case to tell us it's been called
			select {
			case <-useCaseMakePaymentReceivedMessageChan:
				// Expect that the make payment request was called correctly
				is.Equal(mockMakePaymentUseCase.incomingInstructionThatWasPassedToExecute, ufxAndIncomingInst.IncomingInstruction)

				// Sqs queue should have had message deleted
				err := sqs.ExpectNoMessagesInQueueEventually(ctx, awsSession, ufxFileNotificationQueueName, 15*time.Second)
				is.NoErr(err)
			case <-time.After(30 * time.Second):
				t.Fatal("did not receive message in the use case within 30 seconds")
			}
		})
	})
}
