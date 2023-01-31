//go:generate moq -out mocks/s3_moq.go -pkg=mocks . S3Client

package ufx_downloader

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"
)

type S3Client interface {
	ListObjectsV2(ctx context.Context, prefix string) ([]*s3.Object, error)
	GetPresignedURL(ctx context.Context, filename string) (string, error)
}

type UfxDownloader struct {
	s3Client S3Client
}

var filetypeToPrefix = map[string]string{
	"main":      "OIC_Documents_SAXO_BORGUN_",
	"high-risk": "OIC_Documents_SAXO_HR_BORGUN_",
}

func New(client S3Client) UfxDownloader {
	return UfxDownloader{
		s3Client: client,
	}
}

func (ud UfxDownloader) GetPresignedURL(ctx context.Context, date string, filetype string) (string, error) {
	prefix, err := ud.buildFileName(date, filetype)
	if err != nil {
		return "", err
	}

	filenames, err := ud.s3Client.ListObjectsV2(ctx, prefix)
	if err != nil {
		zapctx.Error(ctx, "error listing ufx files with prefix", zap.String("prefix", prefix), zap.Error(err))
		return "", err
	}
	if len(filenames) == 0 {
		return "", UfxDownloaderError{
			fmt.Errorf("no matching files found with prefix: %s", prefix),
			NoFileWithPrefix,
		}
	}

	filename := *filenames[0].Key

	return ud.GetPresignedURLWithFilename(ctx, filename)
}

func (ud UfxDownloader) GetPresignedURLWithFilename(ctx context.Context, filename string) (string, error) {
	url, err := ud.s3Client.GetPresignedURL(ctx, filename)
	if err != nil {
		zapctx.Error(ctx, "error generating presigned url", zap.String("file_name", filename), zap.Error(err))
		return "", err
	}

	return url, nil
}

func (ud UfxDownloader) ValidFileType(filetype string) bool {
	_, found := filetypeToPrefix[filetype]
	return found
}

func (ud UfxDownloader) buildFileName(date string, filetype string) (string, error) {
	// reformat and remove separator in date to match date in filename
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", UfxDownloaderError{err, InvalidDate}
	}
	reformattedDate := d.Format("20060102")
	filePrefix, found := filetypeToPrefix[filetype]
	if !found {
		return "", UfxDownloaderError{fmt.Errorf("incorrect filetype"), FileTypeNotFound}
	}
	// generate filename prefix
	prefix := fmt.Sprint(filePrefix, reformattedDate)
	return prefix, nil
}
