//go:generate moq -out mocks/kafka_producer_moq.go -pkg=mocks . KafkaProducer

package producers

import (
	"context"

	"github.com/saltpay/go-kafka-driver"
)

type KafkaProducer interface {
	WriteMessage(ctx context.Context, msg kafka.Message) error
	Close()
}
