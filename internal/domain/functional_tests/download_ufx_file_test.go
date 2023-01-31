package functional_tests

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader/mocks"
)

func TestUfxDownload_Serve(t *testing.T) {
	is := is.New(t)

	t.Run("Given we receive a request to download a main ufx file with a valid date and exists in the bucket, then return no errors", func(t *testing.T) {
		var (
			ctx      = context.Background()
			date     = "2021-11-22"
			filetype = "main"
			mockedS3 = generateS3Mock([]*s3.Object{
				{
					Key: aws.String("OIC_Documents_SAXO_BORGUN_20211222_1"),
				},
			})
			ud = ufx_downloader.New(mockedS3)
		)

		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.NoErr(err)
	})

	t.Run("Given we receive a request to download a high-risk ufx file with a valid date and exists in the bucket, then return no errors", func(t *testing.T) {
		var (
			ctx      = context.Background()
			date     = "2021-11-22"
			filetype = "high-risk"
			mockedS3 = generateS3Mock([]*s3.Object{
				{
					Key: aws.String("OIC_Documents_SAXO_HR_BORGUN_20211222_1"),
				},
			})
			ud = ufx_downloader.New(mockedS3)
		)

		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.NoErr(err)
	})

	t.Run("Given we receive a request to download a high-risk ufx file with a valid date and does not exist in the bucket, then return error", func(t *testing.T) {
		var (
			ctx      = context.Background()
			date     = "2021-11-22"
			filetype = "high-risk"
			mockedS3 = generateS3Mock([]*s3.Object{})
			ud       = ufx_downloader.New(mockedS3)
		)

		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.True(err != nil)
	})

	t.Run("Given we receive a request to download a main ufx file with a valid date and multiple files exists in the bucket with the same prefix, select the first one", func(t *testing.T) {
		var (
			ctx            = context.Background()
			date           = "2021-11-22"
			filetype       = "high-risk"
			expFileToFetch = "OIC_Documents_SAXO_BORGUN_20211222_1"
			mockedS3       = generateS3Mock([]*s3.Object{
				{
					Key: aws.String(expFileToFetch),
				},
				{
					Key: aws.String("OIC_Documents_SAXO_BORGUN_20211222_1.xml_348_payments.xml"),
				},
			})
			ud = ufx_downloader.New(mockedS3)
		)

		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.NoErr(err)

		spy := mockedS3.GetPresignedURLCalls()
		is.Equal(spy[0].Filename, expFileToFetch)
	})
}

func generateS3Mock(listObjects []*s3.Object) *mocks.S3ClientMock {
	return &mocks.S3ClientMock{
		ListObjectsV2Func: func(ctx context.Context, prefix string) ([]*s3.Object, error) {
			return listObjects, nil
		},
		GetPresignedURLFunc: func(ctx context.Context, filename string) (string, error) {
			return "", nil
		},
	}
}
