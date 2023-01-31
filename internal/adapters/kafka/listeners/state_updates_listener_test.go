//go:build unit
// +build unit

package listeners_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"

	"github.com/saltpay/go-kafka-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners"
	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners/internal/dto"
	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners/internal/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestStateUpdatesListener_Listen(t *testing.T) {
	t.Run("should fetch one message from kafka, execute store payment state update", func(t *testing.T) {
		// Given two state update messages
		var (
			ctx       = context.Background()
			paymentID = "testPaymentID"
		)

		state := dto.PaymentStateUpdate{
			PaymentInstructionID: paymentID,
			UpdatedState:         dto.StateProcessed,
		}
		stateJson, err := json.Marshal(state)
		require.NoError(t, err)
		message := kafka.Message{Value: stateJson}

		// And a mock kafka consumer
		consumerMock := &mocks.ConsumerMock{
			ListenFunc: func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
				err := processor(ctx, message)
				require.NoError(t, err)
			},
		}

		// And a mock state tracker
		trackerMock := &mocks.UpdatePaymentStateMock{
			ExecuteFunc: func(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
				return nil
			},
		}

		// And a state update listener
		stateUpdatesListener := listeners.NewStateUpdatesListener(consumerMock, trackerMock)

		// When Listen is called on the state update listener
		stateUpdatesListener.Listen(ctx)

		// Then UpdatePaymentStatusCalls should be called once with proper payment id and status
		require.Len(t, trackerMock.ExecuteCalls(), 1, "Execute method should be called once")
		assert.Equal(t, paymentID, trackerMock.ExecuteCalls()[0].PaymentInstructionID, "Execute method should be called with proper payment id")
		assert.Equal(t, models.Successful, trackerMock.ExecuteCalls()[0].State, "Execute method should be called with Successful state")
		assert.Equal(t, models.DomainProcessingSucceeded, trackerMock.ExecuteCalls()[0].Event.Type, "An event with ProcessingSucceeded type should be passed to Execute method")
		assert.Empty(t, trackerMock.ExecuteCalls()[0].Event.Details)
	})
	t.Run("should process failure state update message from kafka", func(t *testing.T) {
		// Given one state update messages
		var (
			ctx                = context.Background()
			paymentID          = "testPaymentID"
			testFailureMessage = "testFailureMessage"
		)

		state := dto.PaymentStateUpdate{
			PaymentInstructionID: paymentID,
			UpdatedState:         dto.StateFailure,
			FailureReason: dto.FailureReason{
				Code:    dto.FailureCodeMissingFunding,
				Message: testFailureMessage,
			},
		}
		stateJson, err := json.Marshal(state)
		require.NoError(t, err)
		message := kafka.Message{Value: stateJson}

		// And a mock kafka consumer
		consumerMock := &mocks.ConsumerMock{
			ListenFunc: func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
				err := processor(ctx, message)
				require.NoError(t, err)
			},
		}

		// And a mock state tracker
		trackerMock := &mocks.UpdatePaymentStateMock{
			ExecuteFunc: func(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
				return nil
			},
		}

		// And a state update listener
		stateUpdatesListener := listeners.NewStateUpdatesListener(consumerMock, trackerMock)

		// When Listen is called on the state update listener
		stateUpdatesListener.Listen(ctx)

		// Then UpdatePaymentStatusCalls should be called once with proper payment id and status
		require.Len(t, trackerMock.ExecuteCalls(), 1, "Execute method should be called once")
		assert.Equal(t, paymentID, trackerMock.ExecuteCalls()[0].PaymentInstructionID, "Execute method should be called with proper payment id")
		assert.Equal(t, models.Failed, trackerMock.ExecuteCalls()[0].State, "Execute method should be called with Failed state")
		assert.Equal(t, models.DomainRejected, trackerMock.ExecuteCalls()[0].Event.Type, "An event with Rejected type should be passed to Execute method")
		assert.Equal(t, state.FailureReason, trackerMock.ExecuteCalls()[0].Event.Details)
	})
	t.Run("should handle unmarshal error gracefully", func(t *testing.T) {
		// Given a invalid state update message
		var (
			ctx     = context.Background()
			message = kafka.Message{Value: []byte("invalidJSON")}
		)

		// And a mock kafka consumer
		consumerMock := &mocks.ConsumerMock{
			ListenFunc: func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
				err := processor(ctx, message)
				require.Error(t, err, "should return an error")
				require.Contains(t, err.Error(), "error unmarshalling the message")
			},
		}

		// And a mock state tracker
		trackerMock := &mocks.UpdatePaymentStateMock{
			ExecuteFunc: func(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
				return nil
			},
		}

		// And a state update listener
		stateUpdatesListener := listeners.NewStateUpdatesListener(consumerMock, trackerMock)

		// When Listen is called on the state update listener
		stateUpdatesListener.Listen(ctx)

		// Then Execute should NOT be called
		assert.Len(t, trackerMock.ExecuteCalls(), 0, "Execute method should be called once")
	})
	t.Run("should handle error when storing state update returns an error", func(t *testing.T) {
		// Given a invalid state update message
		var (
			ctx          = context.Background()
			paymentID    = "testPaymentID"
			updatedState = dto.StateProcessed
		)

		state := dto.PaymentStateUpdate{
			PaymentInstructionID: paymentID,
			UpdatedState:         updatedState,
		}
		stateJson, err := json.Marshal(state)
		require.NoError(t, err)
		message := kafka.Message{Value: stateJson}

		// And a mock kafka consumer
		consumerMock := &mocks.ConsumerMock{
			ListenFunc: func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
				err := processor(ctx, message)
				require.Error(t, err, "should return an error")
				require.Contains(t, err.Error(), "error updating state in the db")
			},
		}

		// And a mock state tracker
		trackerMock := &mocks.UpdatePaymentStateMock{
			ExecuteFunc: func(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
				return errors.New("some error")
			},
		}

		// And a state update listener
		stateUpdatesListener := listeners.NewStateUpdatesListener(consumerMock, trackerMock)

		// When Listen is called on the state update listener
		stateUpdatesListener.Listen(ctx)

		// Then Execute should be called once
		assert.Len(t, trackerMock.ExecuteCalls(), 1, "Execute method should be called once")
	})
}
