package sqs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	awssqs "github.com/aws/aws-sdk-go/service/sqs"
)

func ExpectNoMessagesInQueueEventually(ctx context.Context, awsSession *session.Session, queueName string, maxDuration time.Duration) error {
	svc := awssqs.New(awsSession)
	queueURL, err := GetSqsURL(ctx, svc, queueName)
	if err != nil {
		return errors.Wrap(err, "Get sqs queue url failed")
	}

	aout, err := svc.GetQueueAttributes(&awssqs.GetQueueAttributesInput{QueueUrl: &queueURL})
	if err != nil {
		return errors.Wrap(err, "failed to get queue attributes")
	}

	start := time.Now()

	for time.Until(start.Add(maxDuration)) > 0 {
		count := aout.Attributes["ApproximateNumberOfMessages"]
		if count == nil || *count == "0" {
			return nil
		}
	}

	return errors.New("messages still present in queue after max duration of %d")
}

func WaitForQueueToBeEmpty(ctx context.Context, awsSession *session.Session, queueName string) error {
	svc := awssqs.New(awsSession)
	out, err := svc.GetQueueUrlWithContext(ctx, &awssqs.GetQueueUrlInput{QueueName: &queueName})
	if err != nil {
		return err
	}
	queueURL := out.QueueUrl
	delayBetweenAttempts := 500 * time.Millisecond
	maxNumberOfAttempts := 30

	var i int
	for i = 0; i < maxNumberOfAttempts; i++ {
		output, err := svc.GetQueueAttributesWithContext(ctx, &awssqs.GetQueueAttributesInput{QueueUrl: queueURL})
		if err != nil {
			return err
		}

		numberOfMessages := 0
		if output.Attributes[("ApproximateNumberOfMessages")] != nil {
			numberOfMessages, err = strconv.Atoi(*output.Attributes[("ApproximateNumberOfMessages")])
			if err != nil {
				return err
			}
		}

		numberOfMessagesNotVisible := 0
		if output.Attributes[("ApproximateNumberOfMessagesNotVisible")] != nil {
			numberOfMessagesNotVisible, err = strconv.Atoi(*output.Attributes[("ApproximateNumberOfMessagesNotVisible")])
			if err != nil {
				return err
			}
		}

		if numberOfMessages == 0 && numberOfMessagesNotVisible == 0 {
			break
		}

		time.Sleep(delayBetweenAttempts)
	}
	fmt.Printf("[waitForQueueToBeEmpty] Exited the for loop after %d attempts\n", i+1)
	return nil
}

func CreateQueue(ctx context.Context, awsSession *session.Session, queueName string) error {
	sqsSvc := awssqs.New(awsSession)
	_, err := sqsSvc.CreateQueueWithContext(ctx, &awssqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}
	return WaitForQueueToBeEmpty(ctx, awsSession, queueName)
}
