package ufx_downloader

import "github.com/aws/aws-sdk-go/service/s3"

type Error interface {
	error
	Code() string
}

type UfxDownloaderError struct {
	error
	ErrorCode string
}

const (
	BucketNotFound   = s3.ErrCodeNoSuchBucket
	FileNotFound     = s3.ErrCodeNoSuchKey
	FileTypeNotFound = "NoFileType"
	InvalidDate      = "NoValidDate"
	NoFileWithPrefix = "NoFileFoundWithPrefix"
)

func (ude UfxDownloaderError) Code() string {
	return ude.ErrorCode
}

func (ude UfxDownloaderError) Error() string {
	return ude.error.Error()
}
