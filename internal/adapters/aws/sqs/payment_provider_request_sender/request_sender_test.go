//go:build unit
// +build unit

package payment_provider_request_sender

import (
	"context"
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"

	"github.com/saltpay/go-kafka-driver"
	"github.com/stretchr/testify/assert"
)

func TestPaymentProviderRequestSender_SendPaymentInstruction(t *testing.T) {
	t.Run("given a payment instruction with ISB as a payment provider, then send payment instruction to ISB kafka topic", func(t *testing.T) {
		// Given a payment instruction with payment provider 'ISL'
		var (
			ctx                = context.Background()
			paymentInstruction = testhelpers.NewPaymentInstructionBuilder().WithPaymentProvider(models.Islandsbanki).Build()
			bcQueueClientMock  = &mocks.QueueMock{SendMessageFunc: func(context.Context, string) error {
				return nil
			}}
			mockISBProducer = &mocks.KafkaProducerMock{
				WriteMessageFunc: func(ctx context.Context, msg kafka.Message) error {
					return nil
				},
				CloseFunc: func() {
				},
			}
		)
		paymentProviderRequestSender := NewPaymentProviderRequestSender(bcQueueClientMock, mockISBProducer)
		err := paymentProviderRequestSender.SendPaymentInstruction(ctx, paymentInstruction)
		assert.NoError(t, err)

		// then
		assert.Equal(t, 1, len(mockISBProducer.WriteMessageCalls()))
		assert.Equal(t, 0, len(bcQueueClientMock.SendMessageCalls()))
	})

	t.Run("given a payment instruction with BC as a payment provider, then send payment instruction to BC sqs queue", func(t *testing.T) {
		// Given a payment instruction with payment provider 'ISL'
		var (
			ctx                = context.TODO()
			paymentInstruction = testhelpers.NewPaymentInstructionBuilder().WithPaymentProvider(models.BankingCircle).Build()
			bcQueueClientMock  = &mocks.QueueMock{SendMessageFunc: func(context.Context, string) error {
				return nil
			}}
			mockISBProducer = &mocks.KafkaProducerMock{
				WriteMessageFunc: func(ctx context.Context, msg kafka.Message) error {
					return nil
				},
				CloseFunc: func() {
				},
			}
		)
		paymentProviderRequestSender := NewPaymentProviderRequestSender(bcQueueClientMock, mockISBProducer)
		err := paymentProviderRequestSender.SendPaymentInstruction(ctx, paymentInstruction)
		assert.NoError(t, err)

		// then
		assert.Equal(t, 1, len(bcQueueClientMock.SendMessageCalls()))
		assert.Equal(t, 0, len(mockISBProducer.WriteMessageCalls()))
	})
}
