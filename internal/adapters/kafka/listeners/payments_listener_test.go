//go:build unit
// +build unit

package listeners_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/saltpay/go-kafka-driver"
	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners"
	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners/internal/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	portMocks "github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
)

func TestListen(t *testing.T) {
	t.Run("should successfully fetch one message from kafka, execute make payment and commit message", func(t *testing.T) {
		var (
			ctx             = context.Background()
			consumerMock    = &mocks.ConsumerMock{}
			makePaymentMock = &portMocks.MakePaymentMock{}
		)

		_, payload, err := testhelpers.ValidPaymentInstruction()
		assert.NoError(t, err)

		consumerMock.ListenFunc = func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
			err := processor(ctx, kafka.Message{Value: []byte(payload)})
			assert.NoError(t, err)
		}

		makePaymentMock.ExecuteFunc = func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "random payment instruction id", nil
		}

		paymentsListener := listeners.NewPaymentsListener(consumerMock, makePaymentMock, testdoubles.FeatureFlagService{})
		paymentsListener.Listen(ctx)

		assert.Equal(t, len(makePaymentMock.ExecuteCalls()), 1)
	})
	t.Run("should handle unmarshal error gracefully", func(t *testing.T) {
		var (
			ctx             = context.Background()
			consumerMock    = &mocks.ConsumerMock{}
			makePaymentMock = &portMocks.MakePaymentMock{}
		)

		consumerMock.ListenFunc = func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
			err := processor(ctx, kafka.Message{Value: []byte("invalid payload")})
			assert.NoError(t, err)
		}

		makePaymentMock.ExecuteFunc = func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "random payment instruction id", nil
		}

		paymentsListener := listeners.NewPaymentsListener(consumerMock, makePaymentMock, testdoubles.FeatureFlagService{})
		paymentsListener.Listen(ctx)

		assert.Equal(t, len(makePaymentMock.ExecuteCalls()), 0)
	})
	t.Run("should handle error when make payment returns an error", func(t *testing.T) {
		var (
			ctx              = context.Background()
			makePaymentError = fmt.Errorf("unexpected error while executing make payment instruction")
			consumerMock     = &mocks.ConsumerMock{}
			makePaymentMock  = &portMocks.MakePaymentMock{}
		)

		_, payload, err := testhelpers.ValidPaymentInstruction()
		assert.NoError(t, err)

		consumerMock.ListenFunc = func(ctx context.Context, processor kafka.Processor, toggle kafka.CommitStrategy, ps kafka.PauseStrategy) {
			err := processor(ctx, kafka.Message{Value: []byte(payload)})
			assert.NoError(t, err)
		}

		makePaymentMock.ExecuteFunc = func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "", makePaymentError
		}

		paymentsListener := listeners.NewPaymentsListener(consumerMock, makePaymentMock, testdoubles.FeatureFlagService{})
		paymentsListener.Listen(ctx)

		assert.Equal(t, len(makePaymentMock.ExecuteCalls()), 1)
	})
}
