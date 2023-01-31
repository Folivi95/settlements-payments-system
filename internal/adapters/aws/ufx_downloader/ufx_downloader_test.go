package ufx_downloader_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader/mocks"
)

func TestUfxDownloader(t *testing.T) {
	is := is.New(t)

	t.Run("Return error when invalid date passed", func(t *testing.T) {
		var (
			ctx      = context.Background()
			date     = "2021-1-22"
			filetype = "main"
			mockedS3 = generateS3Mock([]*s3.Object{}, nil, nil)
		)
		ud := ufx_downloader.New(mockedS3)
		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.True(err != nil)
	})

	t.Run("Generate file name", func(t *testing.T) {
		var (
			ctx      = context.Background()
			mockedS3 = generateS3Mock([]*s3.Object{
				{
					Key: aws.String("OIC_Documents_SAXO_BORGUN_20211222_1"),
				},
			},
				nil,
				nil)
			ud = ufx_downloader.New(mockedS3)
		)

		scenarios := []struct {
			date             string
			filetype         string
			expectedFilename string
			err              bool
		}{
			{
				"2021-12-22",
				"main",
				"OIC_Documents_SAXO_BORGUN_20211222",
				false,
			},
			{
				"2021-12-22",
				"high-risk",
				"OIC_Documents_SAXO_HR_BORGUN_20211222",
				false,
			},
		}

		for i, scenario := range scenarios {
			_, err := ud.GetPresignedURL(ctx, scenario.date, scenario.filetype)
			is.Equal(err != nil, scenario.err)

			listSpy := mockedS3.ListObjectsV2Calls()
			is.Equal(listSpy[i].Prefix, scenario.expectedFilename)
		}
	})

	t.Run("Return error with invalid filetype passed", func(t *testing.T) {
		var (
			ctx      = context.Background()
			date     = "2021-12-22"
			filetype = "main-risk"
			mockedS3 = generateS3Mock([]*s3.Object{}, nil, nil)
			expError = fmt.Errorf("incorrect filetype")
		)
		ud := ufx_downloader.New(mockedS3)
		_, err := ud.GetPresignedURL(ctx, date, filetype)
		is.True(err != nil)
		is.Equal(err.Error(), expError.Error())
	})
}

func generateS3Mock(listObjects []*s3.Object, listError error, serveError error) *mocks.S3ClientMock {
	return &mocks.S3ClientMock{
		ListObjectsV2Func: func(ctx context.Context, prefix string) ([]*s3.Object, error) {
			return listObjects, listError
		},
		GetPresignedURLFunc: func(ctx context.Context, filename string) (string, error) {
			return "", serveError
		},
	}
}
