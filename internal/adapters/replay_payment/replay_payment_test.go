//go:build unit
// +build unit

package replay_payment_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/matryer/is"
	"github.com/pkg/errors"

	"github.com/saltpay/settlements-payments-system/internal/adapters/replay_payment"
	"github.com/saltpay/settlements-payments-system/internal/adapters/replay_payment/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

func TestReplayPayment_Execute(t *testing.T) {
	is := is.New(t)

	t.Run("Currency validation", func(t *testing.T) {
		t.Run("Returns an error if the currency does not exist", func(t *testing.T) {
			var (
				ctx             = context.Background()
				invalidCurrency = "GPB"
				filename        = testhelpers.RandomString()
				expErrorMessage = "incorrect currency code `GPB`"
			)
			rp := replay_payment.New(nil, nil)
			err := rp.Execute(ctx, invalidCurrency, filename)
			is.True(err != nil)
			is.Equal(err.Error(), expErrorMessage)
		})
	})

	t.Run("Read file from S3", func(t *testing.T) {
		t.Run("Calls s3 to get file once with correct file name", func(t *testing.T) {
			var (
				ctx                = context.Background()
				file               = []byte{1}
				mockedS3           = generateS3Mock(nil, nil, nil)
				mockedUfxConverter = generateUfxConverterMock(file, nil)
				currency           = string(models.EUR)
				filename           = testhelpers.RandomString()
			)

			rp := replay_payment.New(mockedS3, mockedUfxConverter)
			err := rp.Execute(ctx, currency, filename)
			is.NoErr(err)

			spyReadFile := mockedS3.ReadFileCalls()

			readCalls := len(spyReadFile)
			is.Equal(readCalls, 1)

			is.Equal(spyReadFile[0].Filename, filename)
		})

		t.Run("Returns error if file cannot be retrieved from s3", func(t *testing.T) {
			var (
				ctx           = context.Background()
				mockedS3      = generateS3Mock(nil, errors.New("Cannot fetch file"), nil)
				currency      = string(models.EUR)
				filename      = testhelpers.RandomString()
				expectedError = "error getting file from S3. Err: Cannot fetch file"
			)

			rp := replay_payment.New(mockedS3, nil)
			err := rp.Execute(ctx, currency, filename)
			is.True(err != nil)
			is.Equal(err.Error(), expectedError)
		})
	})

	t.Run("FilterCurrency from file", func(t *testing.T) {
		t.Run("Calls the UFX converter to filter the file for the specified currency", func(t *testing.T) {
			var (
				ctx                = context.Background()
				ufxFile            = []byte{1}
				file               = strings.NewReader(testhelpers.RandomString())
				mockedS3           = generateS3Mock(file, nil, nil)
				mockedUfxConverter = generateUfxConverterMock(ufxFile, nil)
				currency           = models.EUR
				filename           = testhelpers.RandomString()
			)

			rp := replay_payment.New(mockedS3, mockedUfxConverter)
			err := rp.Execute(ctx, string(currency), filename)
			is.NoErr(err)

			spyFilterCurrency := mockedUfxConverter.FilterCurrencyCalls()

			is.Equal(len(spyFilterCurrency), 1)

			is.Equal(spyFilterCurrency[0].Currency, currency)
			is.Equal(spyFilterCurrency[0].UfxFileContents, file)
		})

		t.Run("Returns an error if the conversion fails", func(t *testing.T) {
			var (
				ctx                = context.Background()
				file               = strings.NewReader(testhelpers.RandomString())
				mockedS3           = generateS3Mock(file, nil, nil)
				mockedUfxConverter = generateUfxConverterMock(nil, errors.New("oh no"))
				currency           = models.EUR
				filename           = testhelpers.RandomString()
				expectedError      = "error filtering the xml file. Err: oh no"
			)

			rp := replay_payment.New(mockedS3, mockedUfxConverter)
			err := rp.Execute(ctx, string(currency), filename)
			is.True(err != nil)
			is.Equal(err.Error(), expectedError)
		})
	})

	t.Run("Puts file back into S3", func(t *testing.T) {
		t.Run("Calls Put from S3 with the xml and a file name", func(t *testing.T) {
			var (
				ctx                = context.Background()
				file               = strings.NewReader(testhelpers.RandomString())
				ufxFile            = []byte{1}
				mockedS3           = generateS3Mock(file, nil, nil)
				mockedUfxConverter = generateUfxConverterMock(ufxFile, nil)
				currency           = string(models.EUR)
				filename           = "some_ufx_file.xml"
			)

			rp := replay_payment.New(mockedS3, mockedUfxConverter)
			err := rp.Execute(ctx, currency, filename)
			is.NoErr(err)

			spyS3Put := mockedS3.PutBucketFileCalls()

			is.Equal(len(spyS3Put), 1)
			expectedFileName := "REPLAY_some_ufx_file.xml_EUR.xml"
			is.Equal(spyS3Put[0].Filename, expectedFileName)
		})

		t.Run("Returns an error when the call to S3 fails", func(t *testing.T) {
			var (
				ctx                = context.Background()
				file               = strings.NewReader(testhelpers.RandomString())
				ufxFile            = []byte{1}
				mockedS3           = generateS3Mock(file, nil, errors.New("Can't put the file back"))
				mockedUfxConverter = generateUfxConverterMock(ufxFile, nil)
				currency           = string(models.EUR)
				filename           = testhelpers.RandomString()
				expectedErr        = "error sending file to S3. Err: Can't put the file back"
			)

			rp := replay_payment.New(mockedS3, mockedUfxConverter)
			err := rp.Execute(ctx, currency, filename)
			is.True(err != nil)
			is.Equal(err.Error(), expectedErr)
		})
	})
}

func generateUfxConverterMock(ufxFile []byte, err error) *mocks.UfxConverterMock {
	return &mocks.UfxConverterMock{FilterCurrencyFunc: func(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error) {
		return ufxFile, err
	}}
}

func generateS3Mock(reader io.Reader, getError error, putError error) *mocks.S3ClientMock {
	return &mocks.S3ClientMock{
		ReadFileFunc: func(ctx context.Context, filename string) (io.Reader, error) {
			return reader, getError
		},
		PutBucketFileFunc: func(ctx context.Context, file *bytes.Reader, filename string) error {
			return putError
		},
	}
}
