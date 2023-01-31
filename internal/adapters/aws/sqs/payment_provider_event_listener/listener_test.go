//go:build unit
// +build unit

package payment_provider_event_listener_test

import (
	"context"
	"testing"
	"time"

	awssqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/matryer/is"

	mocks2 "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	payment_provider_event_listener2 "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/payment_provider_event_listener"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

const timeout = 20 * time.Millisecond

func TestPaymentProviderEventListener_Listen(t *testing.T) {
	t.Run("executes the trackPaymentOutcome and deletes the message", func(t *testing.T) {
		var (
			ctx     = context.Background()
			is      = is.New(t)
			ppEvent = models.PaymentProviderEvent{}

			deleteCalled = make(chan struct{})
		)
		sqsMessage, err := paymentProviderEventToSQSMessage(ppEvent)
		is.NoErr(err)
		spyIncomingQueue := &mocks2.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			},
			GetMessagesFunc: func(context.Context) (*awssqs.ReceiveMessageOutput, error) {
				return &awssqs.ReceiveMessageOutput{Messages: []*awssqs.Message{sqsMessage}}, nil
			},
		}

		spyUseCase := &mocks.TrackPaymentOutcomeMock{ExecuteFunc: func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}}

		listener := payment_provider_event_listener2.New(
			spyIncomingQueue,
			&mocks2.QueueMock{},
			spyUseCase,
			testdoubles.DummyMetricsClient{},
		)

		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			testhelpers.AssertMessageWasDeleted(t, spyIncomingQueue, *sqsMessage)
			is.True(len(spyUseCase.ExecuteCalls()) > 0)
			is.Equal(spyUseCase.ExecuteCalls()[0].PpEvent, ppEvent)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})

	t.Run("DLQs when the use case fails", func(t *testing.T) {
		var (
			ctx          = context.Background()
			is           = is.New(t)
			deleteCalled = make(chan struct{})
		)
		msg, err := paymentProviderEventToSQSMessage(models.PaymentProviderEvent{})
		is.NoErr(err)
		spyIncomingQueue := &mocks2.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			},
			GetMessagesFunc: func(context.Context) (*awssqs.ReceiveMessageOutput, error) {
				return &awssqs.ReceiveMessageOutput{Messages: []*awssqs.Message{msg}}, nil
			},
		}

		failingUseCase := &mocks.TrackPaymentOutcomeMock{ExecuteFunc: func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return testhelpers2.RandomError()
		}}

		spyDLQ := &mocks2.QueueMock{SendMessageFunc: func(context.Context, string) error {
			return nil
		}}

		listener := payment_provider_event_listener2.New(
			spyIncomingQueue,
			spyDLQ,
			failingUseCase,
			testdoubles.DummyMetricsClient{},
		)

		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			testhelpers.AssertMessageWasDLQd(t, spyDLQ, spyIncomingQueue, *msg)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})

	t.Run("DLQ when the message is not valid from the incoming queue", func(t *testing.T) {
		var (
			ctx          = context.Background()
			msg          = testhelpers.NewSQSMessage("not valid")
			deleteCalled = make(chan struct{})
		)

		spyIncomingQueue := &mocks2.QueueMock{
			DeleteMessageFunc: func(context.Context, string) error {
				deleteCalled <- struct{}{}
				return nil
			},
			GetMessagesFunc: func(context.Context) (*awssqs.ReceiveMessageOutput, error) {
				return &awssqs.ReceiveMessageOutput{Messages: []*awssqs.Message{msg}}, nil
			},
		}

		spyDLQ := &mocks2.QueueMock{SendMessageFunc: func(context.Context, string) error {
			return nil
		}}

		listener := payment_provider_event_listener2.New(
			spyIncomingQueue,
			spyDLQ,
			&mocks.TrackPaymentOutcomeMock{},
			testdoubles.DummyMetricsClient{},
		)

		go listener.Listen(ctx)

		select {
		case <-deleteCalled:
			testhelpers.AssertMessageWasDLQd(t, spyDLQ, spyIncomingQueue, *msg)
		case <-time.After(timeout):
			t.Fatal("timed out waiting for delete message to be called")
		}
	})
}

func paymentProviderEventToSQSMessage(paymentProviderEvent models.PaymentProviderEvent) (*awssqs.Message, error) {
	paymentProviderEventJSON, err := paymentProviderEvent.ToJSON()
	if err != nil {
		return &awssqs.Message{}, err
	}
	return testhelpers.NewSQSMessage(string(paymentProviderEventJSON)), nil
}
