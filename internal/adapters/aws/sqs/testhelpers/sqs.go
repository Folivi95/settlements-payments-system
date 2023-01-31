package testhelpers

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

func AssertMessageWasDLQd(t *testing.T, DLQ *mocks.QueueMock, incomingQueue *mocks.QueueMock, message sqs.Message) {
	t.Helper()
	is := is.New(t)

	is.True(len(DLQ.SendMessageCalls()) > 0) // DLQ was called
	is.Equal(DLQ.SendMessageCalls()[0].S, *message.Body)

	AssertMessageWasDeleted(t, incomingQueue, message)
}

func AssertMessageWasDeleted(t *testing.T, incomingQueue *mocks.QueueMock, message sqs.Message) {
	t.Helper()
	is := is.New(t)
	is.True(len(incomingQueue.DeleteMessageCalls()) > 0) // Delete message from incoming queue called
	is.Equal(incomingQueue.DeleteMessageCalls()[0].MessageHandle, *message.ReceiptHandle)
}

func NewSQSMessage(body string) *sqs.Message {
	return &sqs.Message{
		ReceiptHandle: aws.String(testhelpers.RandomString()),
		Body:          &body,
	}
}
