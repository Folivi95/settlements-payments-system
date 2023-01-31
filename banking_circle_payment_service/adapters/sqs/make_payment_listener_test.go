//go:build unit
// +build unit

package sqs_test

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/matryer/is"
	"github.com/pkg/errors"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/sqs"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"

	awsSqsAdapter "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	awsSqsAdapterMock "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	awsTestHelper "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	domainModelsTestHelpers "github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	domainPorts "github.com/saltpay/settlements-payments-system/internal/domain/ports"
	domainPortMocks "github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

const (
	sleepyTime                     = 1 * time.Second
	numberOfAttempts               = 5
	workerPoolSize                 = 1
	delayBetweenProcessingMessages = 0
)

func TestPaymentInstructionEventListener_Listen(t *testing.T) {
	var (
		dummyDLQ = &awsSqsAdapterMock.QueueMock{SendMessageFunc: func(context.Context, string) error { return nil }}
		is       = is.New(t)
	)

	t.Run("dlq when execute fails", func(t *testing.T) {
		ctx := context.Background()
		deleteCalled := make(chan struct{})

		pi, err := models.PaymentInstruction{}.MustToJSON()
		is.NoErr(err)

		message := awsTestHelper.NewSQSMessage(string(pi))

		failingUseCaseMakePayment := &mocks.MakeBankingCirclePaymentMock{
			ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (providerPaymentID models.ProviderPaymentID, err error) {
				return "", testhelpers.RandomError()
			},
		}

		spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			}, GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
				return &awsSqs.ReceiveMessageOutput{Messages: []*awsSqs.Message{message}}, nil
			},
		}

		spyDLQ := &awsSqsAdapterMock.QueueMock{SendMessageFunc: func(context.Context, string) error { return nil }}

		listener := newListener(failingUseCaseMakePayment, spyIncomingQueue, spyDLQ, testdoubles.FeatureFlagService{}, testdoubles.DummyMetricsClient{})
		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			awsTestHelper.AssertMessageWasDLQd(t, spyDLQ, spyIncomingQueue, *message)
		case <-time.After(sleepyTime):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})

	t.Run("disable toggle when execute fails n times", func(t *testing.T) {
		ctx := context.Background()
		featureToggleOffCalled := make(chan bool)

		pi, err := models.PaymentInstruction{}.MustToJSON()
		is.NoErr(err)

		message := awsTestHelper.NewSQSMessage(string(pi))

		failingUseCaseMakePayment := &mocks.MakeBankingCirclePaymentMock{
			ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (providerPaymentID models.ProviderPaymentID, err error) {
				return "", testhelpers.RandomError()
			},
		}

		spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				return nil
			}, GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
				return &awsSqs.ReceiveMessageOutput{Messages: []*awsSqs.Message{message, message, message, message, message}}, nil
			},
		}

		spyDLQ := &awsSqsAdapterMock.QueueMock{SendMessageFunc: func(context.Context, string) error { return nil }}
		mockedFeatureFlagService := &domainPortMocks.FeatureFlagServiceMock{
			IsIngestionEnabledFromBankingCircleUncheckedQueueFunc: func() bool {
				return true
			},
			IsIngestionEnabledFromBankingCircleUnprocessedQueueFunc: func() bool {
				return true
			},
			ToggleOffIngestionFromBankingCirclePaymentsFunc: func(context.Context) error {
				featureToggleOffCalled <- true
				return nil
			},
		}

		listener := newListener(
			failingUseCaseMakePayment,
			spyIncomingQueue,
			spyDLQ,
			mockedFeatureFlagService,
			testdoubles.DummyMetricsClient{},
		)
		go listener.Listen(ctx)

		select {
		case val := <-featureToggleOffCalled:
			is.True(val)
		case <-time.After(sleepyTime):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})

	t.Run("sleeps when get message fails", func(t *testing.T) {
		ctx := context.Background()
		deleteCalled := make(chan struct{})

		spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			}, GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
				return &awsSqs.ReceiveMessageOutput{}, errors.New("message failed")
			},
		}

		failingUseCaseMakePayment := &mocks.MakeBankingCirclePaymentMock{
			ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (providerPaymentID models.ProviderPaymentID, err error) {
				return "", testhelpers.RandomError()
			},
		}

		spyDLQ := &awsSqsAdapterMock.QueueMock{SendMessageFunc: func(context.Context, string) error { return nil }}

		listener := newListener(failingUseCaseMakePayment, spyIncomingQueue, spyDLQ, testdoubles.FeatureFlagService{}, testdoubles.DummyMetricsClient{})
		go listener.Listen(ctx)

		time.Sleep(sleepyTime)
		is.Equal(len(spyIncomingQueue.GetMessagesCalls()), 1)
	})

	t.Run("it DLQs invalid payment instructions and does not call the use case", func(t *testing.T) {
		ctx := context.Background()
		message := awsTestHelper.NewSQSMessage(testhelpers.RandomString())
		deleteCalled := make(chan struct{})

		spyUseCaseMakePayment := &mocks.MakeBankingCirclePaymentMock{
			ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (providerPaymentID models.ProviderPaymentID, err error) {
				return "", nil
			},
		}

		spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
			GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
				return &awsSqs.ReceiveMessageOutput{Messages: []*awsSqs.Message{message}}, nil
			},
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			},
		}

		spyDLQ := &awsSqsAdapterMock.QueueMock{SendMessageFunc: func(context.Context, string) error { return nil }}

		listener := newListener(spyUseCaseMakePayment, spyIncomingQueue, spyDLQ, testdoubles.FeatureFlagService{}, testdoubles.DummyMetricsClient{})
		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			awsTestHelper.AssertMessageWasDLQd(t, spyDLQ, spyIncomingQueue, *message)
			is.True(len(spyUseCaseMakePayment.ExecuteCalls()) == 0) // execute is never called
		case <-time.After(sleepyTime):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})

	t.Run("DeleteMessage retries error 500", func(t *testing.T) {
		ctx := context.Background()
		numberOfMessages := 1
		deleteWg := &sync.WaitGroup{}
		allDeletesCalled := make(chan bool)

		// create a valid instruction to be executed
		_, paymentInstructionJSON, err := domainModelsTestHelpers.ValidPaymentInstruction()
		is.NoErr(err)
		message := awsTestHelper.NewSQSMessage(paymentInstructionJSON)

		// we want to know when DeleteMessage finished retrying N times
		deleteWg.Add(numberOfAttempts)
		go func() {
			deleteWg.Wait()
			allDeletesCalled <- true
		}()
		spyUseCaseMakePayment := &mocks.MakeBankingCirclePaymentMock{
			ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (providerPaymentID models.ProviderPaymentID, err error) {
				return "", nil
			},
		}

		spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
			GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
				// we want GetMessage to return only once
				if numberOfMessages == 1 {
					numberOfMessages = 0
					return &awsSqs.ReceiveMessageOutput{Messages: []*awsSqs.Message{message}}, nil
				}
				return &awsSqs.ReceiveMessageOutput{}, nil
			},
			DeleteMessageFunc: func(context.Context, string) error {
				deleteWg.Done()
				return awsSqsAdapter.RetryableError{Message: "retry!"}
			},
		}

		spyDLQ := &awsSqsAdapterMock.QueueMock{
			SendMessageFunc: func(context.Context, string) error {
				return nil
			},
		}

		listener := newListener(spyUseCaseMakePayment, spyIncomingQueue, spyDLQ, testdoubles.FeatureFlagService{}, testdoubles.DummyMetricsClient{})
		go listener.Listen(ctx)

		// allDeletesCalled will receive a boolean when all workers are done
		select {
		case <-allDeletesCalled:
			is.Equal(len(spyIncomingQueue.DeleteMessageCalls()), numberOfAttempts)
		case <-time.After(sleepyTime):
			t.Fatal("time out waiting for DeleteMessage retries")
		}
	})

	t.Run("Using a feature flag to decide whether to ingest the unprocessed queue or not", func(t *testing.T) {
		var (
			dummyUseCaseMakePayment = &mocks.MakeBankingCirclePaymentMock{
				ExecuteFunc: func(ctx context.Context, request models.PaymentInstruction) (paymentId models.ProviderPaymentID, err error) {
					return "", nil
				},
			}

			someMessages = &awsSqs.ReceiveMessageOutput{Messages: []*awsSqs.Message{}}
		)

		t.Run("When the feature flag says INGEST the queue", func(t *testing.T) {
			t.Run("attempts to get messages from the queue", func(t *testing.T) {
				ctx := context.Background()
				getMessagesCalled := make(chan struct{})

				// --- given the feature flag says INGEST
				featureFlagSvc := &domainPortMocks.FeatureFlagServiceMock{
					IsIngestionEnabledFromBankingCircleUnprocessedQueueFunc: func() bool {
						return true
					},
				}

				spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
					GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
						getMessagesCalled <- struct{}{}
						return someMessages, nil
					},
				}

				// --- when the payment instruction receiver is activated
				listener := newListener(dummyUseCaseMakePayment, spyIncomingQueue, dummyDLQ, featureFlagSvc, testdoubles.DummyMetricsClient{})
				go listener.Listen(ctx)

				select {
				case <-getMessagesCalled:
					is.True(len(spyIncomingQueue.GetMessagesCalls()) > 0) // get messages was called
				case <-time.After(sleepyTime):
					t.Fatal("timed out waiting for get messages to be called")
				}
			})
		})
		t.Run("When the feature flag says DO NOT INGEST the queue", func(t *testing.T) {
			t.Run("does not attempt to get messages from the queue", func(t *testing.T) {
				ctx := context.Background()
				// --- given the feature flag says DO NOT INGEST
				featureFlagSvc := &domainPortMocks.FeatureFlagServiceMock{
					IsIngestionEnabledFromBankingCircleUnprocessedQueueFunc: func() bool {
						return false
					},
				}

				spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
					GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
						return someMessages, nil
					},
				}

				countCalls := 0
				metricsClient := &domainPortMocks.MetricsClientMock{
					CountFunc: func(_ context.Context, name string, value int64, tags []string) {
						if name == sqs.PaymentsProcessingDisabled && tags[0] == sqs.PaymentProviderTag {
							countCalls++
						}
					},
				}

				// --- when the payment instruction receiver is activated
				listener := newListener(dummyUseCaseMakePayment, spyIncomingQueue, dummyDLQ, featureFlagSvc, metricsClient)
				go listener.Listen(ctx)
				time.Sleep(sleepyTime) // todo: change this plz

				is.True(len(spyIncomingQueue.GetMessagesCalls()) == 0) // didnt get messages
				is.True(countCalls > 0)
			})
		})
		t.Run("When the system receives SIGTERM", func(t *testing.T) {
			t.Run("we stop listening from the queue", func(t *testing.T) {
				// given a queue
				ctx := context.Background()
				spyIncomingQueue := &awsSqsAdapterMock.QueueMock{
					GetMessagesFunc: func(context.Context) (*awsSqs.ReceiveMessageOutput, error) {
						return someMessages, nil
					},
				}

				// await for sigterm signal ---> how to do this better?
				signalCh := make(chan os.Signal, 1)
				signal.Notify(signalCh, syscall.SIGTERM)

				// start listening
				listener := newListener(dummyUseCaseMakePayment, spyIncomingQueue, dummyDLQ, testdoubles.FeatureFlagService{}, testdoubles.DummyMetricsClient{})
				go listener.Listen(ctx)

				// when receives sigterm signal, stops listening
				go func() {
					<-signalCh
					listener.StopListening()
				}()

				signalCh <- syscall.SIGTERM
				time.Sleep(sleepyTime)
				calls := len(spyIncomingQueue.GetMessagesCalls())

				time.Sleep(sleepyTime)
				is.True(len(spyIncomingQueue.GetMessagesCalls()) == calls)
			})
		})
	})
}

func newListener(
	useCaseMakePayment ports.MakeBankingCirclePayment,
	incomingQueueClient awsSqsAdapter.Queue,
	dlq awsSqsAdapter.Queue,
	featureFlagSvc domainPorts.FeatureFlagService,
	metrics domainPorts.MetricsClient,
) sqs.PaymentInstructionEventListener {
	return sqs.NewPaymentInstructionEventListener(
		useCaseMakePayment,
		incomingQueueClient,
		dlq,
		featureFlagSvc,
		metrics,
		workerPoolSize,
		delayBetweenProcessingMessages,
		time.Millisecond*1,
		numberOfAttempts,
		5,
	)
}
