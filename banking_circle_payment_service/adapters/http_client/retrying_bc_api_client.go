package http_client

import (
	"context"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	ports2 "github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type RetryingBCClientOptions struct {
	Sleep            func()
	NumberOfAttempts int
}

type RetryingBankingCircleClient struct {
	delegate ports.BankingCircleAPI
	metrics  ports2.MetricsClient
	options  RetryingBCClientOptions
}

const (
	retryMetricName = "app_http_client_request_retry"
)

var (
	defaultRetryOptions = RetryingBCClientOptions{
		Sleep: func() {
			time.Sleep(30 * time.Second)
		},
		NumberOfAttempts: 5,
	}
	makePaymentMetricsTags = []string{"banking_circle", "make_payment"}
)

func NewRetryingBankingCircleClient(
	delegate ports.BankingCircleAPI,
	options *RetryingBCClientOptions,
	metrics ports2.MetricsClient,
) *RetryingBankingCircleClient {
	if options == nil {
		options = &defaultRetryOptions
	}
	return &RetryingBankingCircleClient{
		delegate: delegate,
		metrics:  metrics,
		options:  *options,
	}
}

func (r RetryingBankingCircleClient) RequestPayment(ctx context.Context, request single_payment_endpoint.RequestDto, slice *[]string) (single_payment_endpoint.ResponseDto, error) {
	var (
		err     error
		payment single_payment_endpoint.ResponseDto
	)

	for i := 0; i < r.options.NumberOfAttempts; i++ {
		if payment, err = r.delegate.RequestPayment(ctx, request, slice); err == nil {
			return payment, nil
		}
		// todo: use retryable errors
		r.observeRetry(ctx, request, i, err)
		r.options.Sleep()
	}
	return payment, err
}

func (r RetryingBankingCircleClient) CheckPaymentStatus(providerPaymentID models.ProviderPaymentID) (ports.PaymentStatus, error) {
	return r.delegate.CheckPaymentStatus(providerPaymentID)
}

func (r RetryingBankingCircleClient) GetRejectionReport(date string) (models2.RejectionReport, error) {
	return r.delegate.GetRejectionReport(date)
}

func (r RetryingBankingCircleClient) observeRetry(ctx context.Context, request single_payment_endpoint.RequestDto, i int, err error) {
	r.metrics.Count(ctx, retryMetricName, 1, makePaymentMetricsTags)
	zapctx.Info(ctx, "BC call failed, will try again",
		zap.Int("attempt", i+1),
		zap.Any("request", request),
		zap.Error(err),
	)
}

func (r RetryingBankingCircleClient) CheckAccountBalance(_ string) (models2.AccountBalance, error) {
	return models2.AccountBalance{}, nil
}
