package producers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/saltpay/go-kafka-driver"
	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/producers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/producers/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	portMocks "github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
)

func TestReportPaymentStatus(t *testing.T) {
	t.Run("should successfully send message to acquiring host kafka topic", func(t *testing.T) {
		var (
			ctx             = context.Background()
			ppEvent         = newPaymentProviderEvent()
			producerMock    = &mocks.ProducerMock{}
			featureFlagMock = &portMocks.FeatureFlagServiceMock{}
		)
		payload, err := json.Marshal(ppEvent)
		assert.NoError(t, err)

		featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsFunc = func() bool {
			return true
		}

		producerMock.WriteMessageFunc = func(ctx context.Context, message kafka.Message) error {
			assert.Equal(t, payload, message.Value)
			return nil
		}

		acquiringHostProducer := producers.NewPaymentStatusExporter(featureFlagMock, producerMock)
		err = acquiringHostProducer.ReportPaymentStatus(ctx, ppEvent)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsCalls()))
		assert.Equal(t, 1, len(producerMock.WriteMessageCalls()))
	})
	t.Run("should not try to send message if feature flag is disabled", func(t *testing.T) {
		var (
			ctx             = context.Background()
			ppEvent         = newPaymentProviderEvent()
			producerMock    = &mocks.ProducerMock{}
			featureFlagMock = &portMocks.FeatureFlagServiceMock{}
		)
		featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsFunc = func() bool {
			return false
		}

		acquiringHostProducer := producers.NewPaymentStatusExporter(featureFlagMock, producerMock)
		err := acquiringHostProducer.ReportPaymentStatus(ctx, ppEvent)
		assert.NoError(t, err)

		assert.Equal(t, 1, len(featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsCalls()))
	})

	t.Run("should return error when writing message to kafka returns an error", func(t *testing.T) {
		var (
			ctx             = context.Background()
			ppEvent         = newPaymentProviderEvent()
			writingMsgError = fmt.Errorf("error while writing msg")
			producerMock    = &mocks.ProducerMock{}
			featureFlagMock = &portMocks.FeatureFlagServiceMock{}
		)
		payload, err := json.Marshal(ppEvent)
		assert.NoError(t, err)

		featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsFunc = func() bool {
			return true
		}

		producerMock.WriteMessageFunc = func(ctx context.Context, message kafka.Message) error {
			assert.Equal(t, payload, message.Value)
			return writingMsgError
		}

		acquiringHostProducer := producers.NewPaymentStatusExporter(featureFlagMock, producerMock)
		err = acquiringHostProducer.ReportPaymentStatus(ctx, ppEvent)

		assert.Equal(t, 1, len(featureFlagMock.IsKafkaPublishingEnableForAcquiringHostTransactionsCalls()))
		assert.Equal(t, 1, len(producerMock.WriteMessageCalls()))
		assert.Same(t, writingMsgError, err)
	})
}

func newPaymentProviderEvent() models.PaymentProviderEvent {
	instruction, _, _ := testhelpers.ValidPaymentInstruction()
	(&instruction).SubmitForProcessing()
	ppEvent := models.PaymentProviderEvent{
		CreatedOn:                time.Now(),
		Type:                     models.Processed,
		PaymentInstruction:       instruction,
		PaymentProviderName:      models.BC,
		PaymentProviderPaymentID: "banking_circle_1234567890",
		FailureReason:            models.FailureReason{},
	}
	return ppEvent
}
