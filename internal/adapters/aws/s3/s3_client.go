package s3

import (
	"bytes"
	"context"
	"io"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Client struct {
	awsSession *session.Session
	bucket     string
}

func NewClient(awsSession *session.Session, bucketName string) Client {
	return Client{awsSession, bucketName}
}

func (s3Client *Client) PutBucketFile(
	ctx context.Context,
	file *bytes.Reader,
	filename string,
) error {
	svc := s3.New(s3Client.awsSession)
	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s3Client.bucket),
		Key:                aws.String(filename),
		Body:               file,
		ContentDisposition: aws.String("attachment"),
		// ACL:                aws.String(S3_ACL),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s3Client *Client) ReadFile(
	ctx context.Context,
	filename string,
) (io.Reader, error) {
	buf := new(bytes.Buffer)

	err := s3Client.ServeFile(ctx, buf, filename)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// ServeFile Given a writer and filename, we write the contents of the file fetched from s3 to the writer directly without loading the file in memory.
// Inputs: w Writer we will write to.
// filename File to fetch from S3.
// Output: error Errors occurred during fetching/writing.

func (s3Client *Client) ServeFile(
	ctx context.Context,
	w io.Writer,
	filename string,
) error {
	svc := s3.New(s3Client.awsSession)

	zapctx.Debug(ctx, "serve file",
		zap.String("bucket", s3Client.bucket),
		zap.String("filename", filename),
	)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Client.bucket),
		Key:    aws.String(filename),
	}

	result, err := svc.GetObjectWithContext(ctx, input)
	if err != nil {
		return err
	}
	defer result.Body.Close()

	_, err = io.Copy(w, result.Body)
	if err != nil {
		return err
	}
	return nil
}

// ListObjectsV2 Given a file prefix, list all files matching the prefix in the bucket
// Inputs:
// prefix: Prefix of File.
// Output:
// []*s3.Object: Metadata about each file queried and ordered in ascending order of key.
func (s3Client *Client) ListObjectsV2(
	ctx context.Context,
	prefix string,
) ([]*s3.Object, error) {
	svc := s3.New(s3Client.awsSession)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Client.bucket),
		Prefix: aws.String(prefix),
	}

	result, err := svc.ListObjectsV2WithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	return result.Contents, nil
}

// GetPresignedURL generate a presigned url to access the requested file directly from AWS S3.
func (s3Client *Client) GetPresignedURL(ctx context.Context, filename string) (string, error) {
	svc := s3.New(s3Client.awsSession)

	zapctx.Debug(ctx, "generating presigned url",
		zap.String("bucket", s3Client.bucket),
		zap.String("filename", filename),
	)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s3Client.bucket),
		Key:    aws.String(filename),
	}

	req, _ := svc.GetObjectRequest(input)
	req.SetContext(ctx)

	return req.Presign(15 * time.Minute)
}
