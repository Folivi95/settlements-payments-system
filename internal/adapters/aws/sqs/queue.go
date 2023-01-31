//go:generate moq -out mocks/queue_moq.go -pkg=mocks . Queue

package sqs

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/service/sqs"
)

type Queue interface {
	DeleteMessage(ctx context.Context, messageHandle string) error
	GetMessages(context.Context) (*sqs.ReceiveMessageOutput, error)
	SendMessage(context.Context, string) error
	PeekAllMessages(context.Context) (DLQInformation, error)
	Purge(context.Context) error
	Attributes(context.Context) (QueueAttributes, error)
}

type QueueName string

const (
	BcUnprocessedPayments     QueueName = "bc-unprocessed"
	BcUnprocessedPaymentsDlq  QueueName = "bc-unprocessed-dlq"
	IsbUnprocessedPayments    QueueName = "isb-unprocessed"
	IsbUnprocessedPaymentsDlq QueueName = "isb-unprocessed-dlq"
	ProcessedPayments         QueueName = "processed"
	ProcessedPaymentsDlq      QueueName = "processed-dlq"
	UfxFileEvents             QueueName = "ufx"
	UfxFileEventsDlq          QueueName = "ufx-dlq"
	UncheckedPayments         QueueName = "unchecked"
	UncheckedPaymentsDlq      QueueName = "unchecked-dlq"
)

func (q QueueName) IsDlq() bool {
	return strings.Contains(string(q), "-dlq")
}

type QueueClientMapping map[QueueName]Queue

type DLQInformation struct {
	Count    int
	Messages []string
}

type QueueAttributes map[string]*string

func NewDLQInformationFromJSON(in io.Reader) (DLQInformation, error) {
	var out DLQInformation
	err := json.NewDecoder(in).Decode(&out)
	return out, err
}
