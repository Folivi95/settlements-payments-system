//go:build unit
// +build unit

package http_client_test

import (
	"context"
	"errors"
	"testing"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestRetryingBankingCircleClient_CheckPaymentStatus(t *testing.T) {
	t.Run("return back whatever the delegate gives", func(t *testing.T) {
		is := is.New(t)

		expectedStatus := ports.PaymentStatus("whatever")
		paymentInstructionID := models.ProviderPaymentID("some-id")

		spyDelegate := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(paymentInstructionId models.ProviderPaymentID) (ports.PaymentStatus, error) {
			return expectedStatus, nil
		}}

		client := http_client.NewRetryingBankingCircleClient(spyDelegate, nil, testdoubles.DummyMetricsClient{})
		paymentStatus, err := client.CheckPaymentStatus(paymentInstructionID)

		is.NoErr(err)
		is.Equal(paymentStatus, expectedStatus)
		is.Equal(len(spyDelegate.CheckPaymentStatusCalls()), 1)
		is.Equal(spyDelegate.CheckPaymentStatusCalls()[0].PaymentID, paymentInstructionID)
	})
}

func TestRetryingBankingCircleClient_RequestPayment(t *testing.T) {
	t.Run("return back whatever the delegate gives", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)

		request := single_payment_endpoint.RequestDto{}
		expectedResponse := single_payment_endpoint.ResponseDto{Status: "whatever", PaymentID: "bleh"}

		spyDelegate := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(ctx context.Context, request single_payment_endpoint.RequestDto, slice *[]string) (single_payment_endpoint.ResponseDto, error) {
			return expectedResponse, nil
		}}

		client := http_client.NewRetryingBankingCircleClient(spyDelegate, nil, testdoubles.DummyMetricsClient{})
		response, err := client.RequestPayment(ctx, request, nil)

		is.NoErr(err)
		is.Equal(response, expectedResponse)
		is.Equal(len(spyDelegate.RequestPaymentCalls()), 1)
		is.Equal(spyDelegate.RequestPaymentCalls()[0].Request, request)
	})

	t.Run("delegate failing", func(t *testing.T) {
		t.Run("retries after the first fail, sleeping in between", func(t *testing.T) {
			var (
				ctx = context.Background()
				is  = is.New(t)
			)

			request := single_payment_endpoint.RequestDto{}
			expectedResponse := single_payment_endpoint.ResponseDto{Status: "whatever", PaymentID: "bleh"}

			spyDelegate := &mocks.BankingCircleAPIMock{}
			spyDelegate.RequestPaymentFunc = func(ctx context.Context, request single_payment_endpoint.RequestDto, slice *[]string) (single_payment_endpoint.ResponseDto, error) {
				if len(spyDelegate.RequestPaymentCalls()) == 1 {
					return single_payment_endpoint.ResponseDto{}, errors.New("oh no")
				}
				return expectedResponse, nil
			}

			spySleeper := &spySleeper{}

			client := http_client.NewRetryingBankingCircleClient(
				spyDelegate,
				&http_client.RetryingBCClientOptions{
					Sleep:            spySleeper.DummySleep,
					NumberOfAttempts: 2,
				},
				testdoubles.DummyMetricsClient{},
			)

			response, err := client.RequestPayment(ctx, request, nil)

			is.NoErr(err)
			is.Equal(response, expectedResponse)
			is.Equal(spySleeper.Calls, 1)
		})

		t.Run("gives up if the error is not transient after max number of attempts", func(t *testing.T) {
			var (
				ctx = context.Background()
				is  = is.New(t)
			)

			expectedErr := errors.New("always failing")
			maxNumberOfAttempts := 2

			spyDelegate := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(ctx context.Context, request single_payment_endpoint.RequestDto, slice *[]string) (single_payment_endpoint.ResponseDto, error) {
				return single_payment_endpoint.ResponseDto{}, expectedErr
			}}
			spySleeper := &spySleeper{}

			client := http_client.NewRetryingBankingCircleClient(
				spyDelegate,
				&http_client.RetryingBCClientOptions{
					Sleep:            spySleeper.DummySleep,
					NumberOfAttempts: maxNumberOfAttempts,
				},
				testdoubles.DummyMetricsClient{},
			)

			_, err := client.RequestPayment(ctx, single_payment_endpoint.RequestDto{}, nil)

			is.Equal(err, expectedErr)
			is.Equal(spySleeper.Calls, maxNumberOfAttempts)
		})
	})
}

type spySleeper struct {
	Calls int
}

func (s *spySleeper) DummySleep() {
	s.Calls++
}
