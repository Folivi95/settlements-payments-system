package sqs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type QueueClient struct {
	queueURL                 *string
	queueName                string
	receiveWaitTimeSeconds   *int64
	visibilityTimeoutSeconds *int64
	receiveBatchSize         *int64
	sqsSvc                   *sqs.SQS
	metrics                  ports.MetricsClient
}

type NewQueueClientOptions struct {
	AwsSession             *session.Session
	QueueName              string
	ReceiveWaitTimeSeconds int64
	VisibilityTimeout      int64
	ReceiveBatchSize       int64
	Metrics                ports.MetricsClient
}

func NewQueueClient(ctx context.Context, options NewQueueClientOptions) (QueueClient, error) {
	sqsSvc := sqs.New(options.AwsSession)
	queueURL, err := GetSqsURL(ctx, sqsSvc, options.QueueName)
	if err != nil {
		return QueueClient{}, errors.Wrap(err, "get sqs queue url failed")
	}
	sqsClient := QueueClient{
		queueURL:                 aws.String(queueURL),
		queueName:                options.QueueName,
		receiveWaitTimeSeconds:   aws.Int64(options.ReceiveWaitTimeSeconds),
		visibilityTimeoutSeconds: aws.Int64(options.VisibilityTimeout),
		receiveBatchSize:         aws.Int64(options.ReceiveBatchSize),
		sqsSvc:                   sqsSvc,
		metrics:                  options.Metrics,
	}
	return sqsClient, nil
}

const (
	queueCountMetricName = "app_queue_messages_received"
)

func GetSqsURL(ctx context.Context, sqsSvc *sqs.SQS, queueName string) (string, error) {
	req, err := sqsSvc.GetQueueUrlWithContext(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", err
	}
	if *req.QueueUrl == "" {
		return "", errors.New(fmt.Sprintf("empty queue URL returned by AWS for %s", queueName))
	}
	return *req.QueueUrl, nil
}

// GetMessages fetches messages from an AWS SQS queue.
func (sqsClient *QueueClient) GetMessages(ctx context.Context) (*sqs.ReceiveMessageOutput, error) {
	msgResult, err := sqsClient.sqsSvc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            sqsClient.queueURL,
		MaxNumberOfMessages: sqsClient.receiveBatchSize,
		VisibilityTimeout:   sqsClient.visibilityTimeoutSeconds, // avoid multiple consumers picking up the same message
		WaitTimeSeconds:     sqsClient.receiveWaitTimeSeconds,   // long-polling
	})
	if err != nil {
		return nil, fmt.Errorf("could not receive message from queue client: %v", err)
	}
	sqsClient.observeReceiveMessageCount(ctx, msgResult)

	return msgResult, nil
}

func (sqsClient *QueueClient) PeekAllMessages(ctx context.Context) (DLQInformation, error) {
	if !strings.Contains(sqsClient.queueName, "deadletter") {
		return DLQInformation{}, fmt.Errorf("invalid queue name or deadletter queue unavailable, queue name: %s", sqsClient.queueName)
	}

	dlqMessageCount, err := sqsClient.getDLQMessageCount(ctx)
	if err != nil {
		return DLQInformation{}, err
	}

	dlqInformation := DLQInformation{}

	wantMessageCount := dlqMessageCount
	timeout := time.Now().Add(time.Second * time.Duration(*sqsClient.visibilityTimeoutSeconds))

	for wantMessageCount > 0 && time.Until(timeout) > 0 {
		messages, err := sqsClient.GetMessages(ctx)
		if err != nil {
			return DLQInformation{}, fmt.Errorf("could not get messages, error: %v", err)
		}

		for _, message := range messages.Messages {
			dlqInformation.Messages = append(dlqInformation.Messages, *message.Body)
		}
		wantMessageCount -= len(messages.Messages)
	}

	dlqInformation.Count = len(dlqInformation.Messages)
	return dlqInformation, nil
}

func (sqsClient *QueueClient) SendMessage(ctx context.Context, messageBody string) error {
	_, err := sqsClient.sqsSvc.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    sqsClient.queueURL,
		MessageBody: aws.String(messageBody),
	})
	return err
}

func (sqsClient *QueueClient) DeleteMessage(ctx context.Context, messageHandle string) error {
	// ideally we would call sqsClient.sqsSvc.DeleteMessage(),
	// however, for the required retry capabilities, we need to take a look at the HTTP status code,
	// so we do it this way instead
	req, _ := sqsClient.sqsSvc.DeleteMessageRequest(&sqs.DeleteMessageInput{
		QueueUrl:      sqsClient.queueURL,
		ReceiptHandle: aws.String(messageHandle),
	})
	req.SetContext(ctx)

	err := req.Send()
	if err != nil && req.HTTPResponse != nil && req.HTTPResponse.StatusCode >= 500 {
		return RetryableError{Message: fmt.Sprintf("failed to delete message due to sqs error: %v", err)}
	}
	return err
}

func (sqsClient *QueueClient) Purge(ctx context.Context) error {
	if _, err := sqsClient.sqsSvc.PurgeQueueWithContext(ctx, &sqs.PurgeQueueInput{QueueUrl: sqsClient.queueURL}); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case sqs.ErrCodeQueueDoesNotExist:
				zapctx.Error(ctx, "error queue does not exists when trying to purge", zap.String("queue_name", sqsClient.queueName))
			case sqs.ErrCodePurgeQueueInProgress:
				zapctx.Error(ctx, "error queue %s already has a purge in progress", zap.String("queue_name", sqsClient.queueName))
			}
		}
		return err
	}

	zapctx.Error(ctx, "successfully purged queue", zap.String("queue_name", sqsClient.queueName))
	return nil
}

func (sqsClient *QueueClient) DeleteQueue(ctx context.Context) error {
	_, err := sqsClient.sqsSvc.DeleteQueueWithContext(ctx, &sqs.DeleteQueueInput{
		QueueUrl: sqsClient.queueURL,
	})
	return err
}

func (sqsClient *QueueClient) Attributes(ctx context.Context) (QueueAttributes, error) {
	allAttributes := "All"
	input := sqs.GetQueueAttributesInput{
		QueueUrl:       sqsClient.queueURL,
		AttributeNames: []*string{&allAttributes},
	}

	attr, err := sqsClient.sqsSvc.GetQueueAttributesWithContext(ctx, &input)
	if err != nil {
		zapctx.Error(ctx, "error getting attributes for queue", zap.String("queue_name", sqsClient.queueName))
		return nil, err
	}

	queueAttributes := make(QueueAttributes)
	for k, v := range attr.Attributes {
		queueAttributes[k] = v
	}

	return queueAttributes, nil
}

func (sqsClient *QueueClient) QueueName() string {
	return sqsClient.queueName
}

func (sqsClient *QueueClient) observeReceiveMessageCount(ctx context.Context, receiveMsgOutput *sqs.ReceiveMessageOutput) {
	if receiveMsgOutput == nil {
		return
	}
	msgCount := len(receiveMsgOutput.Messages)
	if msgCount > 0 {
		zapctx.Debug(ctx, "fetched messages from SQS", zap.Int("total", msgCount))
		queueNameLabel := sqsClient.queueName
		sqsClient.metrics.Count(ctx, queueCountMetricName, int64(msgCount), []string{queueNameLabel})
	}
}

func (sqsClient *QueueClient) getDLQMessageCount(ctx context.Context) (int, error) {
	awsAttribute := "ApproximateNumberOfMessages"
	resp, err := sqsClient.sqsSvc.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		AttributeNames: []*string{
			aws.String(awsAttribute),
		},
		QueueUrl: sqsClient.queueURL,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get queue attributes: %v", err)
	}

	messageCountString := resp.Attributes[awsAttribute]
	messageCount, err := strconv.Atoi(*messageCountString)
	if err != nil {
		return 0, fmt.Errorf("failed to convert message attribute: %v", err)
	}

	return messageCount, nil
}

type RetryableError struct {
	Message string
}

func (r RetryableError) Error() string {
	return fmt.Sprintf("error 5xx from DeleteMessage: %s", r.Message)
}
