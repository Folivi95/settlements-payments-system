package payment_provider_request_sender

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/producers"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"

	"github.com/saltpay/go-kafka-driver"
)

type PaymentProviderRequestSender struct {
	bcSqsClient      sqs.Queue
	isbKafkaProducer producers.KafkaProducer
}

func NewPaymentProviderRequestSender(bcSqsClient sqs.Queue, isbKafkaProducer producers.KafkaProducer) PaymentProviderRequestSender {
	return PaymentProviderRequestSender{
		bcSqsClient:      bcSqsClient,
		isbKafkaProducer: isbKafkaProducer,
	}
}

// todo CB: unit test this and simplify integration test

func (b PaymentProviderRequestSender) SendPaymentInstruction(ctx context.Context, paymentInstruction models.PaymentInstruction) error {
	jsonBytes, err := paymentInstruction.MarshalJSON()
	if err != nil {
		return err
	}

	if paymentInstruction.PaymentProvider() == models.Islandsbanki {
		err = b.isbKafkaProducer.WriteMessage(ctx, kafka.Message{Value: jsonBytes})
	} else {
		err = b.bcSqsClient.SendMessage(ctx, string(jsonBytes))
	}
	return err
}
