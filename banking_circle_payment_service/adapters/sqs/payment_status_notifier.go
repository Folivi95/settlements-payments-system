package sqs

import (
	"context"
	"encoding/json"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	ppEvents "github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentStatusNotifier struct {
	sqsClient sqs.QueueClient
}

func NewPaymentStatusNotifier(sqsClient sqs.QueueClient) PaymentStatusNotifier {
	return PaymentStatusNotifier{
		sqsClient: sqsClient,
	}
}

func (psn PaymentStatusNotifier) SendPaymentStatus(ctx context.Context, event ppEvents.PaymentProviderEvent) error {
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return psn.sqsClient.SendMessage(ctx, string(jsonBytes))
}
