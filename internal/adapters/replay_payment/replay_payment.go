//go:generate moq -out mocks/ufx_converter_mock.go -pkg=mocks . UfxConverter
//go:generate moq -out mocks/s3_moq.go -pkg=mocks . S3Client

package replay_payment

import (
	"bytes"
	"context"
	"fmt"
	"io"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type (
	S3Client interface {
		ReadFile(ctx context.Context, filename string) (io.Reader, error)
		PutBucketFile(ctx context.Context, file *bytes.Reader, filename string) error
	}

	UfxConverter interface {
		FilterCurrency(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error)
	}

	MakePaymentExecutionError struct {
		IncomingInstruction models.IncomingInstruction
		Error               error
	}
)

const (
	PayCurrencyFromFile string = "pay_currency_from_file"
)

type ReplayPayment struct {
	s3Client     S3Client
	ufxConverter UfxConverter
}

func New(s3Client S3Client, ufxConverter UfxConverter) ReplayPayment {
	return ReplayPayment{
		s3Client:     s3Client,
		ufxConverter: ufxConverter,
	}
}

func (rp ReplayPayment) Execute(ctx context.Context, currency string, filename string) error {
	currencyCode := models.CurrencyCode(currency)

	err := rp.validateCurrency(currencyCode)
	if err != nil {
		zapctx.Error(ctx, "error validating request parameters", zap.Error(err))
		return err
	}

	ufxFile, err := rp.s3Client.ReadFile(ctx, filename)
	if err != nil {
		zapctx.Error(ctx, "error reading file", zap.Error(err))
		return fmt.Errorf("error getting file from S3. Err: %v", err)
	}

	filteredUFX, err := rp.ufxConverter.FilterCurrency(ufxFile, currencyCode)
	if err != nil {
		zapctx.Error(ctx, "error filtering currency", zap.Error(err))
		return fmt.Errorf("error filtering the xml file. Err: %v", err)
	}

	err = rp.s3Client.PutBucketFile(ctx, bytes.NewReader(filteredUFX), fmt.Sprintf("REPLAY_%s_%s.xml", filename, currency))
	if err != nil {
		zapctx.Error(ctx, "error sending the file to S3", zap.Error(err))
		return fmt.Errorf("error sending file to S3. Err: %v", err)
	}

	return nil
}

func (rp ReplayPayment) validateCurrency(currency models.CurrencyCode) error {
	_, found := models.CurrenciesToIso[currency]
	if !found {
		return fmt.Errorf("incorrect currency code `%s`", currency)
	}
	return nil
}
