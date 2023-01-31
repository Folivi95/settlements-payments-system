package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	zapctx "github.com/saltpay/go-zap-ctx"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	awsEndpoint := os.Getenv("AWS_ENDPOINT")
	if awsEndpoint == "" {
		log.Fatal("missing AWS_ENDPOINT environment variable")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("missing AWS_REGION environment variable")
	}

	err := WaitForConnectivity(ctx, []string{awsEndpoint}, 60, time.Second)
	if err != nil {
		zapctx.Fatal(ctx, "failed to wait for connectivity", zap.Error(err))
	}

	awsSession, err := CreateAwsSession(awsRegion, awsEndpoint)
	if err != nil {
		zapctx.Fatal(ctx, "failed to create AWS session", zap.Error(err))
	}

	connectedToLocalStack := time.Now()
	defer func() {
		zapctx.Info(ctx, "localstack components ready", zap.Duration("elapsed", time.Since(connectedToLocalStack)))
	}()

	var wg sync.WaitGroup
	wg.Add(12)

	queueEnvNames := []string{"SQS_BANKING_CIRCLE_UNPROCESSED_QUEUE_NAME", "SQS_BANKING_CIRCLE_UNPROCESSED_DLQ_NAME", "SQS_BANKING_CIRCLE_PROCESSED_QUEUE_NAME", "SQS_BANKING_CIRCLE_PROCESSED_DLQ_NAME", "SQS_BANKING_CIRCLE_UNCHECKED_QUEUE_NAME", "SQS_BANKING_CIRCLE_UNCHECKED_DLQ_NAME", "SQS_ISLANDSBANKI_UNPROCESSED_QUEUE_NAME", "SQS_ISLANDSBANKI_UNPROCESSED_DLQ_NAME", "SQS_UFX_FILE_NOTIFICATION_QUEUE_NAME", "SQS_UFX_FILE_NOTIFICATION_DLQ_NAME"}

	queues := []string{}
	for _, queueEnvName := range queueEnvNames {
		queueName := os.Getenv(queueEnvName)
		if queueName == "" {
			log.Fatal("missing " + queueEnvName + " environment variable")
		}
		queues = append(queues, queueName)
	}

	// create sqs queues
	for _, queueName := range queues {
		go createSqsQueue(ctx, &wg, awsSession, queueName)
	}

	s3UfxPaymentFilesBucketName := os.Getenv("S3_UFX_PAYMENT_FILES_BUCKET_NAME")
	if s3UfxPaymentFilesBucketName == "" {
		log.Fatal("missing S3_UFX_PAYMENT_FILES_BUCKET_NAME environment variable")
	}

	sqsUfxFileNotificationQueueName := os.Getenv("SQS_UFX_FILE_NOTIFICATION_QUEUE_NAME")
	if sqsUfxFileNotificationQueueName == "" {
		log.Fatal("missing SQS_UFX_FILE_NOTIFICATION_QUEUE_NAME environment variable")
	}

	// create s3 bucket
	go createS3Bucket(ctx, &wg, awsSession, s3UfxPaymentFilesBucketName)
	// set up bucket notification
	go setUpS3BucketNotifications(ctx, &wg, awsSession, s3UfxPaymentFilesBucketName, sqsUfxFileNotificationQueueName)

	wg.Wait()
}

func createSqsQueue(ctx context.Context, wg *sync.WaitGroup, awsSession *session.Session, queueName string) {
	defer wg.Done()
	zapctx.Info(ctx, "creating sqs queue", zap.String("queue_name", queueName))

	svc := sqs.New(awsSession)
	_, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]*string{
			"MessageRetentionPeriod": aws.String("86400"),
		},
	})
	if err != nil {
		zapctx.Fatal(ctx, "failed to create sqs queue", zap.String("queue_name", queueName), zap.Error(err))
	}
}

func createS3Bucket(ctx context.Context, wg *sync.WaitGroup, awsSession *session.Session, bucketName string) {
	defer wg.Done()
	zapctx.Info(ctx, "creating s3 bucket", zap.String("bucket_name", bucketName))

	svc := s3.New(awsSession)
	_, err := svc.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil && !strings.Contains(err.Error(), "BucketAlreadyExists") && !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
		zapctx.Info(ctx, "failed to create s3 bucket", zap.String("bucket_name", bucketName), zap.Error(err))
	}

	zapctx.Info(ctx, "s3 bucket created", zap.String("bucket_name", bucketName))
}

func setUpS3BucketNotifications(ctx context.Context, wg *sync.WaitGroup, awsSession *session.Session, bucketName string, queueName string) {
	defer wg.Done()
	zapctx.Info(ctx, "setting up notifications for s3 bucket", zap.String("bucket_name", bucketName))

	svc := s3.New(awsSession)
	_, err := svc.PutBucketNotificationConfigurationWithContext(ctx, &s3.PutBucketNotificationConfigurationInput{
		Bucket: aws.String(bucketName),
		NotificationConfiguration: &s3.NotificationConfiguration{
			QueueConfigurations: []*s3.QueueConfiguration{
				{
					Events: []*string{
						aws.String("s3:ObjectCreated:*"),
					},
					QueueArn: aws.String(queueName),
				},
			},
		},
	})
	if err != nil {
		zapctx.Fatal(ctx, "failed to set up bucket notifications",
			zap.String("bucket_name", bucketName),
			zap.String("sqs_name", queueName),
			zap.Error(err),
		)
	}
	zapctx.Info(ctx, "notifications for s3 bucket ans sqs queue created",
		zap.String("bucket_name", bucketName),
		zap.String("sqs_name", queueName),
	)
}

func CreateAwsSession(awsRegion string, awsEndpoint string) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(awsRegion),
		Endpoint:         aws.String(awsEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return sess, nil
}

// WaitForConnectivity will iterate the addresses supplied and retry / wait until network is reachable
// before returning ok.
func WaitForConnectivity(
	ctx context.Context,
	remoteAddrToWaitFor []string,
	retryCount int,
	retryDelay time.Duration,
) error {
	// Check each address one by one
	for _, address := range remoteAddrToWaitFor {
		start := time.Now()
		for i := 0; i <= retryCount; i++ {
			// if we tested the number of times we expected it, we give up
			if i == retryCount {
				return fmt.Errorf("could not connect to %s; tried %d times in %v", address, retryCount, time.Since(start))
			}

			zapctx.Info(ctx, fmt.Sprintf("attempt %d in connecting to %s", i, address))
			conn, err := net.Dial("tcp", address)
			if err == nil {
				// No error, managed to connect. We can skip this address
				_ = conn.Close()
				break
			}

			// delay retry and try again
			time.Sleep(retryDelay)
		}
	}

	return nil
}
