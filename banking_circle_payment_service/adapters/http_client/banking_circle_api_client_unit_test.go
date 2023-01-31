//go:build unit
// +build unit

package http_client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client/mocks"
	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestBankingCircleApiClient_RequestPayment(t *testing.T) {
	var (
		emptyRequest  = single_payment_endpoint.RequestDto{}
		emptyResponse = single_payment_endpoint.ResponseDto{}
		dummyMetrics  = testdoubles.DummyMetricsClient{}
		dummySlice    = make([]string, 0)

		makeClientConfiguredTo = func(baseURL string) *http_client.BankingCircleAPIClient {
			bankingCircleAPIClient, _ := http_client.NewAPIClient(&http.Client{}, baseURL, dummyMetrics)
			return bankingCircleAPIClient
		}
	)

	t.Run("successful request to api", func(t *testing.T) {
		var (
			ctx              = context.Background()
			is               = is.New(t)
			expectedResponse = single_payment_endpoint.ResponseDto{
				PaymentID: "123",
				Status:    "Processed",
			}
			requestedPath   string
			requestedMethod string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			resAsJSON, _ := expectedResponse.ToJSON()
			_, _ = w.Write(resAsJSON)
			requestedPath = r.URL.Path
			requestedMethod = r.Method
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.RequestPayment(ctx, emptyRequest, &dummySlice)

		// Then
		is.NoErr(err)
		is.Equal(actualResponse, expectedResponse)
		is.Equal(requestedPath, "/payments/singles")
		is.Equal(requestedMethod, http.MethodPost)
	})

	t.Run("requests should have an unique header", func(t *testing.T) {
		var (
			ctx              = context.Background()
			is               = is.New(t)
			expectedResponse = single_payment_endpoint.ResponseDto{
				PaymentID: "123",
				Status:    "Processed",
			}
			requestedPath   string
			requestedMethod string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			resAsJSON, _ := expectedResponse.ToJSON()
			_, _ = w.Write(resAsJSON)
			requestedPath = r.URL.Path
			requestedMethod = r.Method
			id := r.Header.Get("X-Request-ID")
			is.True(id != "")
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.RequestPayment(ctx, emptyRequest, &dummySlice)

		// Then
		is.NoErr(err)
		is.Equal(actualResponse, expectedResponse)
		is.Equal(requestedPath, "/payments/singles")
		is.Equal(requestedMethod, http.MethodPost)
	})

	t.Run("when bc returns unauthorised an error is returned", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.RequestPayment(ctx, emptyRequest, &dummySlice)

		_, isUnauthorisedErr := err.(http_client.UnauthorisedWithBankingCircleError)
		is.True(isUnauthorisedErr)
		is.Equal(actualResponse, emptyResponse)
	})

	t.Run("when bc returns 400 return invalid payment error", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
			uid string
		)

		validationError := "this is bad"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(validationError))
			uid = r.Header.Get("X-Request-ID")
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.RequestPayment(ctx, emptyRequest, &dummySlice)

		invalidPaymentErr, isInvalidPaymentErr := err.(http_client.InvalidPaymentRequestError)
		is.True(isInvalidPaymentErr)
		is.Equal(invalidPaymentErr, http_client.InvalidPaymentRequestError{
			Request:      emptyRequest,
			ErrorMessage: validationError,
			UniqueID:     http_client.UniqueID(uid),
		})
		is.Equal(actualResponse, emptyResponse)
	})

	t.Run("fails for an unexpected reason", func(t *testing.T) {
		var (
			ctx       = context.Background()
			is        = is.New(t)
			errorBody = "bad times"
			errorCode = http.StatusTeapot
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(errorCode)
			_, _ = w.Write([]byte(errorBody))
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		_, err := bankingCircleAPIClient.RequestPayment(ctx, emptyRequest, &dummySlice)

		is.True(err != nil)

		unrecognisedErr, isUnrecognisedErr := err.(http_client.UnrecognisedBankingCircleError)
		is.True(isUnrecognisedErr)
		is.Equal(unrecognisedErr.Status, errorCode)
		is.Equal(unrecognisedErr.Body, errorBody)
		is.Equal(unrecognisedErr.Action, http_client.RequestPayment)
	})

	t.Run("when http fails, return an err (rather than a panic)", func(t *testing.T) {
		var (
			ctx               = context.Background()
			is                = is.New(t)
			failingHTTPClient = &mocks.HTTPDoerMock{}
		)
		failingHTTPClient.DoFunc = func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("oh no i fail")
		}

		bankingCircleAPIClient, _ := http_client.NewAPIClient(failingHTTPClient, "", dummyMetrics)
		_, err := bankingCircleAPIClient.RequestPayment(ctx, single_payment_endpoint.RequestDto{}, &dummySlice)

		is.True(err != nil)
	})
}

func TestBankingCircleApiClient_CheckPaymentStatus(t *testing.T) {
	var (
		dummyMetrics  = testdoubles.DummyMetricsClient{}
		somePaymentID = models.ProviderPaymentID("123")

		makeClientConfiguredTo = func(baseURL string) *http_client.BankingCircleAPIClient {
			bankingCircleAPIClient, _ := http_client.NewAPIClient(&http.Client{}, baseURL, dummyMetrics)
			return bankingCircleAPIClient
		}
	)

	t.Run("successful check from api", func(t *testing.T) {
		var (
			is              = is.New(t)
			response        = http_client.BankingCirclePaymentStatusResponse{Status: "Processed"}
			requestedPath   string
			requestedMethod string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestedPath = r.URL.Path
			requestedMethod = r.Method
			w.WriteHeader(http.StatusOK)
			json, _ := response.ToJSON()
			_, _ = w.Write(json)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		paymentStatus, err := bankingCircleAPIClient.CheckPaymentStatus(somePaymentID)

		is.NoErr(err)
		is.Equal(paymentStatus, response.Status)
		is.Equal(requestedMethod, http.MethodGet)
		is.Equal(requestedPath, "/payments/singles/123/status")
	})

	t.Run("return an error if bc says unauthorised", func(t *testing.T) {
		is := is.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.CheckPaymentStatus(somePaymentID)
		_, isUnauthorisedErr := err.(http_client.UnauthorisedWithBankingCircleError)

		is.True(isUnauthorisedErr)
		is.Equal(actualResponse, ports.PaymentStatus(""))
	})

	t.Run("return an error if payment not found", func(t *testing.T) {
		is := is.New(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		actualResponse, err := bankingCircleAPIClient.CheckPaymentStatus(somePaymentID)
		_, isNotFoundErr := err.(http_client.PaymentNotFoundError)

		is.True(isNotFoundErr)
		is.Equal(actualResponse, ports.PaymentStatus(""))
	})

	t.Run("return unexpected error, if its unexpected", func(t *testing.T) {
		var (
			is        = is.New(t)
			errorBody = "bad times"
			errorCode = http.StatusBadRequest
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(errorCode)
			_, _ = w.Write([]byte(errorBody))
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		_, err := bankingCircleAPIClient.CheckPaymentStatus(somePaymentID)

		is.True(err != nil)

		unrecognisedErr, isUnrecognisedErr := err.(http_client.UnrecognisedBankingCircleError)
		is.True(isUnrecognisedErr)
		is.Equal(unrecognisedErr.Status, errorCode)
		is.Equal(unrecognisedErr.Body, errorBody)
		is.Equal(unrecognisedErr.Action, http_client.CheckingPayment)
	})
}

func TestBankingCircleApiClient_GetRejectionReport(t *testing.T) {
	var (
		dummyMetrics = testdoubles.DummyMetricsClient{}

		makeClientConfiguredTo = func(baseURL string) *http_client.BankingCircleAPIClient {
			bankingCircleAPIClient, _ := http_client.NewAPIClient(&http.Client{}, baseURL, dummyMetrics)
			return bankingCircleAPIClient
		}
	)

	t.Run("successful check from api", func(t *testing.T) {
		var (
			is              = is.New(t)
			response        = models2.RejectionReport{Rejections: []models2.Rejection{}}
			requestedMethod string
			date            string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			keys := r.URL.Query()["TransactionDate"]
			date = keys[0]
			requestedMethod = r.Method
			w.WriteHeader(http.StatusOK)
			json, _ := response.ToJSON()
			_, _ = w.Write(json)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		expectedDate := "2021-10-07"
		report, err := bankingCircleAPIClient.GetRejectionReport(expectedDate)

		is.NoErr(err)
		is.Equal(report, response)
		is.Equal(requestedMethod, http.MethodGet)
		is.Equal(date, expectedDate)
	})
	t.Run("if BC fails, we should receive an empty report", func(t *testing.T) {
		is := is.New(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

		expectedDate := "2021-10-07"
		report, err := bankingCircleAPIClient.GetRejectionReport(expectedDate)
		is.Equal(report, models2.RejectionReport{})
		unrecognisedErr, isUnrecognisedErr := err.(http_client.UnrecognisedBankingCircleError)
		is.True(isUnrecognisedErr)
		is.Equal(unrecognisedErr.Status, http.StatusInternalServerError)
		is.Equal(unrecognisedErr.Body, "")
		is.Equal(unrecognisedErr.Action, http_client.RejectionReport)
	})
}

func TestBankingCircleApiClient_CheckAccountBalance(t *testing.T) {
	var (
		dummyMetrics = testdoubles.DummyMetricsClient{}
		accountID    = "accountId"

		makeClientConfiguredTo = func(baseURL string) *http_client.BankingCircleAPIClient {
			bankingCircleAPIClient, _ := http_client.NewAPIClient(&http.Client{}, baseURL, dummyMetrics)
			return bankingCircleAPIClient
		}
	)

	t.Run("happy path", func(t *testing.T) {
		t.Run("successful check from api", func(t *testing.T) {
			var (
				is              = is.New(t)
				response        = models2.AccountBalance{Result: []models2.Balance{}}
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.Path
				requestedMethod = r.Method
				w.WriteHeader(http.StatusOK)
				json, _ := response.ToJSON()
				_, _ = w.Write(json)
			}))
			defer server.Close()

			bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

			actualAccountBalance, err := bankingCircleAPIClient.CheckAccountBalance(accountID)

			is.NoErr(err)
			is.Equal(response, actualAccountBalance)
			is.Equal(requestedMethod, http.MethodGet)
			is.Equal(requestedPath, "/accounts/"+accountID+"/balances")
		})

		t.Run("requests should have an unique header", func(t *testing.T) {
			var (
				is              = is.New(t)
				response        = models2.AccountBalance{Result: []models2.Balance{}}
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.Path
				requestedMethod = r.Method
				w.WriteHeader(http.StatusOK)
				json, _ := response.ToJSON()
				_, _ = w.Write(json)
			}))
			defer server.Close()

			bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

			actualAccountBalance, err := bankingCircleAPIClient.CheckAccountBalance(accountID)

			// Then
			is.NoErr(err)
			is.Equal(actualAccountBalance, response)
			is.Equal(requestedPath, "/accounts/"+accountID+"/balances")
			is.Equal(requestedMethod, http.MethodGet)
		})
	})

	t.Run("unhappy path", func(t *testing.T) {
		t.Run("unable to unmarshall the response", func(t *testing.T) {
			var (
				is              = is.New(t)
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.Path
				requestedMethod = r.Method
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test"))
			}))
			defer server.Close()

			bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

			_, err := bankingCircleAPIClient.CheckAccountBalance(accountID)
			isError := err != nil

			is.True(isError)
			is.Equal(requestedMethod, http.MethodGet)
			is.Equal(requestedPath, "/accounts/"+accountID+"/balances")
		})

		t.Run("receive different status response then 200 we get an error and empty response", func(t *testing.T) {
			var (
				is              = is.New(t)
				response        = models2.AccountBalance{}
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestedPath = r.URL.Path
				requestedMethod = r.Method
				w.WriteHeader(http.StatusBadRequest)
				json, _ := response.ToJSON()
				_, _ = w.Write(json)
			}))
			defer server.Close()

			bankingCircleAPIClient := makeClientConfiguredTo(server.URL)

			actualAccountBalance, err := bankingCircleAPIClient.CheckAccountBalance(accountID)

			// Then
			isError := err != nil

			is.True(isError)
			is.Equal(actualAccountBalance, response)
			is.Equal(requestedPath, "/accounts/"+accountID+"/balances")
			is.Equal(requestedMethod, http.MethodGet)
		})
	})
}
