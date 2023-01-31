//go:generate moq -out mocks/s3_moq.go -pkg=mocks . S3Client

package aws

import (
	"bytes"
	"context"
)

type S3Client interface {
	PutBucketFile(ctx context.Context, file *bytes.Reader, filename string) error
}

type UfxFileUploader struct {
	s3Client S3Client
}

func NewUfxFileUploader(s3Client S3Client) UfxFileUploader {
	return UfxFileUploader{
		s3Client: s3Client,
	}
}

func (u UfxFileUploader) AddFileToS3Bucket(ctx context.Context, file *bytes.Reader, filename string) error {
	return u.s3Client.PutBucketFile(ctx, file, filename)
}
