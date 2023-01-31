package sqs

import (
	awsSqs "github.com/aws/aws-sdk-go/service/sqs"
)

type DummyQueueClient struct{}

func (d DummyQueueClient) DeleteMessage(messageHandle string) error {
	return nil
}

func (d DummyQueueClient) GetMessages() (*awsSqs.ReceiveMessageOutput, error) {
	return nil, nil
}

func (d DummyQueueClient) SendMessage(s string) error {
	return nil
}

func (d DummyQueueClient) PeekAllMessages() (DLQInformation, error) {
	return DLQInformation{}, nil
}

func (d DummyQueueClient) Purge() error {
	return nil
}

func (d DummyQueueClient) Attributes() (QueueAttributes, error) {
	return QueueAttributes{}, nil
}
