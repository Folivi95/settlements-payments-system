//go:build unit
// +build unit

package ufx_file_listener_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/matryer/is"

	mocks3 "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener"
	mocks2 "github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
)

const timeout = 5 * time.Second

func TestUfxFileListener_Listen(t *testing.T) {
	t.Run("when a valid instruction is found in the file store, and we execute the usecase successfully, then we send the message on and delete the message", func(t *testing.T) {
		var (
			ctx          = context.Background()
			is           = is.New(t)
			msg          = testhelpers.NewSQSMessage(`{"Records": [{"s3": {"object": {"key": "test.ufx"}}}]}`)
			deleteCalled = make(chan bool)
		)

		ufxAndIncomingInst := testhelpers2.ValidUfxAndIncomingInstruction()

		var (
			spyIncomingQueue                         = &mocks3.QueueMock{}
			stubFileStore                            = &mocks2.FileStoreMock{}
			spyUseCase                               = &mocks.MakePaymentMock{}
			stubCheckPaymentAccountFundsAvailability = &mocks.CheckPaymentAccountFundsAvailabilityMock{
				ExecuteFunc: func(ctx context.Context, code models.CurrencyCode, amount float64, highRisk bool) (bool, error) {
					return true, nil
				},
			}
			metricsClient = &mocks.MetricsClientMock{
				CountFunc:     func(ctx context.Context, name string, value int64, _ []string) {},
				HistogramFunc: func(context.Context, string, float64, []string) {},
			}
			mockFeatureFlagService = &mocks.FeatureFlagServiceMock{IsIngestionEnabledFromUfxFileNotificationQueueFunc: func() bool {
				return true
			}}
		)

		listener := ufx_file_listener.New(
			spyUseCase,
			stubCheckPaymentAccountFundsAvailability,
			&ufx_file_listener.UfxToPaymentInstructionsConverter{},
			stubFileStore,
			spyIncomingQueue,
			&mocks3.QueueMock{},
			10,
			metricsClient,
			mockFeatureFlagService,
		)

		stubFileStore.ReadFileFunc = func(ctx context.Context, filename string) (io.Reader, error) {
			return strings.NewReader(ufxAndIncomingInst.UfxFileContents), nil
		}

		spyIncomingQueue.GetMessagesFunc = func(ctx context.Context) (*sqs.ReceiveMessageOutput, error) {
			return &sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{msg},
			}, nil
		}

		spyIncomingQueue.DeleteMessageFunc = func(ctx context.Context, messageHandle string) error {
			deleteCalled <- true
			return nil
		}

		spyUseCase.ExecuteFunc = func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "", nil
		}

		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			is.True(len(spyUseCase.ExecuteCalls()) > 0)                                                        // called the use case
			is.Equal(spyUseCase.ExecuteCalls()[0].IncomingInstruction, ufxAndIncomingInst.IncomingInstruction) // use case was called from the extracted instruction
			testhelpers.AssertMessageWasDeleted(t, spyIncomingQueue, *msg)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for message to be deleted")
		}
	})

	t.Run("send message to DLQ if the instruction is invalid", func(t *testing.T) {
		var (
			ctx                 = context.Background()
			invalidFileContents = `<?xml version="1.0" encoding="utf-8"?><DocFile><FileHeader></FileHeader><`
			msg                 = testhelpers.NewSQSMessage(`{"Records": [{"s3": {"object": {"key": "poop.xml"}}}]}`)
			dlqCalled           = make(chan bool)
		)

		stubMakePayment := &mocks.MakePaymentMock{}
		stubCheckPaymentAccountFundsAvailability := &mocks.CheckPaymentAccountFundsAvailabilityMock{}
		stubFileStore := &mocks2.FileStoreMock{}
		spyIncomingQueue := &mocks3.QueueMock{}
		spyDLQ := &mocks3.QueueMock{}
		mockFeatureFlagService := &mocks.FeatureFlagServiceMock{IsIngestionEnabledFromUfxFileNotificationQueueFunc: func() bool {
			return true
		}}

		listener := ufx_file_listener.New(
			stubMakePayment,
			stubCheckPaymentAccountFundsAvailability,
			&ufx_file_listener.UfxToPaymentInstructionsConverter{},
			stubFileStore,
			spyIncomingQueue,
			spyDLQ,
			10,
			testdoubles.DummyMetricsClient{},
			mockFeatureFlagService,
		)

		stubFileStore.ReadFileFunc = func(context.Context, string) (io.Reader, error) {
			return strings.NewReader(invalidFileContents), nil
		}

		spyDLQ.SendMessageFunc = func(context.Context, string) error {
			dlqCalled <- true
			return nil
		}

		spyIncomingQueue.GetMessagesFunc = func(context.Context) (*sqs.ReceiveMessageOutput, error) {
			return &sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{msg},
			}, nil
		}

		spyIncomingQueue.DeleteMessageFunc = func(context.Context, string) error {
			return nil
		}

		go listener.Listen(ctx)

		select {
		case <-dlqCalled:
			testhelpers.AssertMessageWasDLQd(t, spyDLQ, spyIncomingQueue, *msg)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for message to be sent to DLQ")
		}
	})

	t.Run("send message to DLQ if records are empty", func(t *testing.T) {
		var (
			ctx                 = context.Background()
			invalidFileContents = `<?xml version="1.0" encoding="utf-8"?>
<DocFile>
	<FileHeader></FileHeader><`
			msg       = testhelpers.NewSQSMessage(`{}`)
			dlqCalled = make(chan bool)
		)

		stubMakePayment := &mocks.MakePaymentMock{}
		stubCheckPaymentAccountFundsAvailability := &mocks.CheckPaymentAccountFundsAvailabilityMock{}
		stubFileStore := &mocks2.FileStoreMock{}
		stubIncomingQueue := &mocks3.QueueMock{}
		spyDLQ := &mocks3.QueueMock{}
		mockFeatureFlagService := &mocks.FeatureFlagServiceMock{IsIngestionEnabledFromUfxFileNotificationQueueFunc: func() bool {
			return true
		}}

		listener := ufx_file_listener.New(
			stubMakePayment,
			stubCheckPaymentAccountFundsAvailability,
			&ufx_file_listener.UfxToPaymentInstructionsConverter{},
			stubFileStore,
			stubIncomingQueue,
			spyDLQ,
			10,
			testdoubles.DummyMetricsClient{},
			mockFeatureFlagService,
		)

		stubFileStore.ReadFileFunc = func(context.Context, string) (io.Reader, error) {
			return strings.NewReader(invalidFileContents), nil
		}

		spyDLQ.SendMessageFunc = func(context.Context, string) error {
			dlqCalled <- true
			return nil
		}

		stubIncomingQueue.GetMessagesFunc = func(context.Context) (*sqs.ReceiveMessageOutput, error) {
			return &sqs.ReceiveMessageOutput{
				Messages: []*sqs.Message{msg},
			}, nil
		}

		stubIncomingQueue.DeleteMessageFunc = func(context.Context, string) error {
			return nil
		}

		go listener.Listen(ctx)

		select {
		case <-dlqCalled:
			testhelpers.AssertMessageWasDLQd(t, spyDLQ, stubIncomingQueue, *msg)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for message to be sent to DLQ")
		}
	})

	t.Run("feature flag", func(t *testing.T) {
		var (
			mockMakePayment = &mocks.MakePaymentMock{ExecuteFunc: func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
				return "", nil
			}}
			mockCheckPaymentAccountFundsAvailability = &mocks.CheckPaymentAccountFundsAvailabilityMock{ExecuteFunc: func(ctx context.Context, code models.CurrencyCode, amount float64, highRisk bool) (bool, error) {
				return true, nil
			}}
			ufxAndIncomingInst = testhelpers2.ValidUfxAndIncomingInstruction()

			mockFileStore = &mocks2.FileStoreMock{ReadFileFunc: func(ctx context.Context, filename string) (io.Reader, error) {
				return strings.NewReader(ufxAndIncomingInst.UfxFileContents), nil
			}}
			mockDLQ     = &mocks3.QueueMock{}
			msg         = testhelpers.NewSQSMessage(`{"Records": [{"s3": {"object": {"key": "test.ufx"}}}]}`)
			ctx, cancel = context.WithCancel(context.Background())
		)

		t.Run("when IsIngestionEnabledFromUfxFileNotificationQueueCalls flag is enabled, then GetMessages should be called", func(t *testing.T) {
			var (
				mockFeatureFlagService = &mocks.FeatureFlagServiceMock{IsIngestionEnabledFromUfxFileNotificationQueueFunc: func() bool {
					return true
				}}
				ctx               = context.Background()
				getMessagesCalled = make(chan bool)
			)

			mockIncomingQueue := &mocks3.QueueMock{GetMessagesFunc: func(contextMoqParam context.Context) (*sqs.ReceiveMessageOutput, error) {
				getMessagesCalled <- true
				return &sqs.ReceiveMessageOutput{
					Messages: []*sqs.Message{msg},
				}, nil
			}}

			mockIncomingQueue.DeleteMessageFunc = func(ctx context.Context, messageHandle string) error {
				return nil
			}

			listener := ufx_file_listener.New(mockMakePayment,
				mockCheckPaymentAccountFundsAvailability,
				&ufx_file_listener.UfxToPaymentInstructionsConverter{},
				mockFileStore,
				mockIncomingQueue,
				mockDLQ,
				10,
				testdoubles.DummyMetricsClient{},
				mockFeatureFlagService)

			go listener.Listen(ctx)
			defer cancel()

			select {
			case <-getMessagesCalled:
				assert.GreaterOrEqual(t, len(mockFeatureFlagService.IsIngestionEnabledFromUfxFileNotificationQueueCalls()), 1, "feature flag service should be called")
				assert.GreaterOrEqual(t, len(mockIncomingQueue.GetMessagesCalls()), 1, "GetMessages should be called")
			case <-time.After(timeout):
				t.Fatal("timed out waiting for feature flag to be checked")
			}
		})

		t.Run("when IsIngestionEnabledFromUfxFileNotificationQueueCalls flag is disabled, then GetMessages should not be called", func(t *testing.T) {
			isIngestionEnabled := make(chan bool)
			mockFeatureFlagService := &mocks.FeatureFlagServiceMock{IsIngestionEnabledFromUfxFileNotificationQueueFunc: func() bool {
				isIngestionEnabled <- true
				return false
			}}

			mockIncomingQueue := &mocks3.QueueMock{GetMessagesFunc: func(contextMoqParam context.Context) (*sqs.ReceiveMessageOutput, error) {
				return nil, nil
			}}

			listener := ufx_file_listener.New(mockMakePayment,
				mockCheckPaymentAccountFundsAvailability,
				&ufx_file_listener.UfxToPaymentInstructionsConverter{},
				mockFileStore,
				mockIncomingQueue,
				mockDLQ,
				10,
				testdoubles.DummyMetricsClient{},
				mockFeatureFlagService)

			go listener.Listen(ctx)
			defer cancel()

			select {
			case <-isIngestionEnabled:
				assert.GreaterOrEqual(t, len(mockFeatureFlagService.IsIngestionEnabledFromUfxFileNotificationQueueCalls()), 1, "feature flag service should be called")
				assert.Len(t, mockIncomingQueue.GetMessagesCalls(), 0, "GetMessages should not be called")
			case <-time.After(timeout):
				t.Fatal("timed out waiting for feature flag to be checked")
			}
		})
	})
}
