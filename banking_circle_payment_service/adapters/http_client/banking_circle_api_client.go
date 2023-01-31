//go:generate moq -out mocks/http_doer_moq.go -pkg=mocks . HTTPDoer

package http_client

// todo: We have req.Header.Add("X-Request-ID", uuid) for all methods that can be moved into a middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/google/uuid"

	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	ports2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"

	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type BankingCircleAPIClient struct {
	baseURL    string
	httpClient HTTPDoer
	observer   *bcClientObserver
}

type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

type AuthResponseDto struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func NewAPIClient(client HTTPDoer, baseURL string, metricsClient ports.MetricsClient) (*BankingCircleAPIClient, error) {
	return &BankingCircleAPIClient{
		httpClient: client,
		baseURL:    baseURL,
		observer: &bcClientObserver{
			metricsClient: metricsClient,
		},
	}, nil
}

func (b *BankingCircleAPIClient) RequestPayment(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error) {
	url := b.baseURL + "/payments/singles"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(request.ToJSON()))
	if err != nil {
		return spe.ResponseDto{}, err
	}

	uniqueID := uuid.NewString()
	*slice = append(*slice, uniqueID)
	req.Header.Add("X-Request-ID", uniqueID)

	body, status, err := b.getAndObserveResponse(req, "make_payment_instruction")
	if err != nil {
		return spe.ResponseDto{}, err
	}

	switch status {
	case http.StatusUnauthorized:
		return spe.ResponseDto{}, UnauthorisedWithBankingCircleError{URL: url, ResponseBody: string(body), UniqueID: UniqueID(uniqueID)}
	case http.StatusBadRequest:
		return spe.ResponseDto{}, InvalidPaymentRequestError{
			Request:      request,
			ErrorMessage: string(body),
			UniqueID:     UniqueID(uniqueID),
		}
	case http.StatusCreated:
		return spe.NewResponseDTOFromJSON(body, models.BankingReference(request.DebtorReference))
	default:
		return spe.ResponseDto{}, UnrecognisedBankingCircleError{
			URL:      url,
			Action:   RequestPayment,
			Status:   status,
			Body:     string(body),
			UniqueID: UniqueID(uniqueID),
		}
	}
}

func (b *BankingCircleAPIClient) CheckPaymentStatus(paymentRequestID models.ProviderPaymentID) (ports2.PaymentStatus, error) {
	url := fmt.Sprintf("%s/payments/singles/%s/status", b.baseURL, paymentRequestID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	uniqueID := uuid.NewString()
	req.Header.Add("X-Request-ID", uniqueID)

	body, status, err := b.getAndObserveResponse(req, "check_payment_status")
	if err != nil {
		return "", err
	}

	switch status {
	case http.StatusUnauthorized:
		return "", UnauthorisedWithBankingCircleError{
			URL:          url,
			ResponseBody: string(body),
			UniqueID:     UniqueID(uniqueID),
		}
	case http.StatusNotFound:
		return "", PaymentNotFoundError{
			URL:       url,
			PaymentID: string(paymentRequestID),
			UniqueID:  UniqueID(uniqueID),
		}
	case http.StatusOK:
		statusResponse, err := NewBankingCirclePaymentStatusResponseFromJSON(body)
		if err != nil {
			return "", err
		}
		return statusResponse.Status, err
	default:
		return "", UnrecognisedBankingCircleError{
			URL:      url,
			Action:   CheckingPayment,
			Status:   status,
			Body:     string(body),
			UniqueID: UniqueID(uniqueID),
		}
	}
}

func (b *BankingCircleAPIClient) GetRejectionReport(date string) (models2.RejectionReport, error) {
	parameters := "&IncludeReceived=false&IncludeMissingFunds=false"
	url := fmt.Sprintf("%s/reports/rejection-report/?TransactionDate=%s", b.baseURL, date+parameters)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("accept", "application/json")
	if err != nil {
		return models2.RejectionReport{}, err
	}

	uniqueID := uuid.NewString()
	req.Header.Add("X-Request-ID", uniqueID)

	body, status, err := b.getAndObserveResponse(req, "rejection_report")
	if err != nil {
		return models2.RejectionReport{}, err
	}

	switch status {
	case http.StatusOK:
		rejectionReport, err := models2.NewRejectionReportFromJSON(body)
		if err != nil {
			return models2.RejectionReport{}, err
		}
		return rejectionReport, nil
	default:
		return models2.RejectionReport{}, UnrecognisedBankingCircleError{
			URL:      url,
			Action:   RejectionReport,
			Status:   status,
			Body:     string(body),
			UniqueID: UniqueID(uniqueID),
		}
	}
}

func (b *BankingCircleAPIClient) CheckAccountBalance(accountID string) (models2.AccountBalance, error) {
	url := b.baseURL + "/accounts/" + accountID + "/balances"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return models2.AccountBalance{}, err
	}

	uniqueID := uuid.NewString()
	req.Header.Add("X-Request-ID", uniqueID)

	body, status, err := b.getAndObserveResponse(req, "account_balance")
	if err != nil {
		return models2.AccountBalance{}, err
	}

	switch status {
	case http.StatusOK:
		accountBalance, err := models2.NewAccountBalanceFromJSON(body)
		if err != nil {
			return models2.AccountBalance{}, err
		}
		return accountBalance, nil
	default:
		return models2.AccountBalance{}, fmt.Errorf("failed to GET account balance with status %d", status)
	}
}

func (b *BankingCircleAPIClient) getAndObserveResponse(req *http.Request, operation string) ([]byte, int, error) {
	startTime := time.Now()

	resp, err := b.httpClient.Do(b.addHTTPTracing(req))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, readAllErr := io.ReadAll(resp.Body)
	if readAllErr != nil {
		return nil, 0, fmt.Errorf("problem reading response body, status code from server: %d", resp.StatusCode)
	}

	b.observer.reportResponse(
		req.Context(),
		time.Since(startTime).Milliseconds(),
		resp.StatusCode,
		operation,
	)

	return body, resp.StatusCode, nil
}

func (b *BankingCircleAPIClient) addHTTPTracing(req *http.Request) *http.Request {
	traceCtx := httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			// TODO: fix the context used here, using the request context seems tricky because I'm not sure when this
			// function runs
			b.observer.ReusedConnection(context.Background(), info.Reused)
		},
	})

	return req.WithContext(traceCtx)
}
