package black_box_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	models2 "github.com/saltpay/settlements-payments-system/black-box-tests/models"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type SettlementsAPIClient struct {
	client  *http.Client
	baseURL string
	token   string
}

const (
	timeout      = 1 * time.Minute
	tickDuration = 2 * time.Second
	waitTime     = 15 * time.Second
)

func NewClient(baseURL string, token string) *SettlementsAPIClient {
	client := &http.Client{Timeout: 15 * time.Second}
	return &SettlementsAPIClient{client, baseURL, token}
}

func (m *SettlementsAPIClient) CheckIfHealthy() error {
	url := m.baseURL + "/health_check"
	res, err := m.client.Get(url)
	if err != nil {
		return err
	}
	_ = res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 from %q but got %d", url, res.StatusCode)
	}

	return nil
}

func (m *SettlementsAPIClient) SendPaymentInstruction(input models.IncomingInstruction) (models.ProviderPaymentID, error) {
	incomingInstructionJSON, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("json marshal error when creating payment instruction input: %w", err)
	}
	body := bytes.NewReader(incomingInstructionJSON)

	req, err := http.NewRequest(http.MethodPost, m.baseURL+"/payments", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	res, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error posting payment instruction to /payments: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("expected %v, but got %v. Error: %s", http.StatusCreated, res.StatusCode, body)
	}

	response, err := handlers.NewPaymentResponseFromJSON(res.Body)
	if err != nil {
		return "", fmt.Errorf("could not get a payment response: %w", err)
	}

	return response.ID, nil
}

func (m *SettlementsAPIClient) CheckPaymentWasSuccessful(ctx context.Context, id models.ProviderPaymentID) error {
	return m.pollForStatus(ctx, id, models.Successful)
}

func (m *SettlementsAPIClient) CheckFailedPayment(ctx context.Context, id models.ProviderPaymentID) error {
	return m.pollForStatus(ctx, id, models.Failed)
}

func (m *SettlementsAPIClient) CheckRejectedPayment(ctx context.Context, id models.ProviderPaymentID) error {
	return m.pollForStatus(ctx, id, models.Rejected)
}

func (m *SettlementsAPIClient) pollForStatus(ctx context.Context, id models.ProviderPaymentID, desiredStatus models.PaymentInstructionStatus) error {
	var lastStatus models.PaymentInstructionStatus

	url := m.baseURL + "/payments/" + string(id)

	endTime := time.Now().Add(timeout)
	for time.Until(endTime) > 0 {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+m.token)

		res, err := m.client.Do(req)
		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK {
			errBody, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			zapctx.Info(ctx, "response status code",
				zap.Int("status_code", res.StatusCode),
				zap.String("err_message", string(errBody)),
				zap.String("url", url),
			)

			wait()
			continue
		}

		zapctx.Info(ctx, "response status code",
			zap.Int("status_code", res.StatusCode),
			zap.String("url", url),
		)

		body, err := io.ReadAll(res.Body)
		if err != nil {
			zapctx.Error(ctx, "could not read the body of the response", zap.Error(err))

			return err
		}

		_ = res.Body.Close()
		paymentInstruction, err := models.NewPaymentInstructionFromJSON(body)
		if err != nil {
			return err
		}

		lastStatus = paymentInstruction.GetStatus()
		if paymentInstruction.GetStatus() == desiredStatus {
			return nil
		}

		wait()
	}
	return fmt.Errorf("did not receive status %q within %v, last status: %q, from %s", desiredStatus, timeout, lastStatus, url)
}

func (m *SettlementsAPIClient) GetReportToday() error {
	url := m.baseURL + "/payments/report"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("got %d from %s, %s", res.StatusCode, url, string(errBody))
	}

	_, err = models.NewPaymentReportFromJSON(res.Body)
	if err != nil {
		return fmt.Errorf("failed to parse report body: %w", err)
	}

	return nil
}

func (m *SettlementsAPIClient) GetReport() error {
	dateString := time.Now().Format("2006-01-02")
	url := m.baseURL + "/payments/report/" + dateString
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("got %d from %s, %s", res.StatusCode, url, string(errBody))
	}

	_, err = models.NewPaymentReportFromJSON(res.Body)
	if err != nil {
		return fmt.Errorf("failed to parse report body: %w", err)
	}

	return nil
}

func (m *SettlementsAPIClient) CheckUnprocessedPayments() error {
	url := m.baseURL + "/internal/dead-letter-queues/bc-unprocessed-dlq"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("got %d from %s, %s", res.StatusCode, url, string(errBody))
	}

	return nil
}

type RejectionResponse struct {
	Rejections []interface{} `json:"rejections"`
}

func (m *SettlementsAPIClient) RejectionReport(ctx context.Context) error {
	var statusCode int
	dateString := time.Now().Format("2006-01-02")
	url := m.baseURL + "/bc-report/" + dateString

	endTime := time.Now().Add(timeout)
	for time.Until(endTime) > 0 {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+m.token)

		res, err := m.client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		statusCode = res.StatusCode

		if res.StatusCode != http.StatusOK {
			errBody, _ := io.ReadAll(res.Body)
			zapctx.Info(ctx, "response status code",
				zap.Int("status_code", res.StatusCode),
				zap.String("err_message", string(errBody)),
				zap.String("url", url),
			)

			continue
		} else {
			rs := new(RejectionResponse)
			if err := json.NewDecoder(res.Body).Decode(rs); err != nil {
				return err
			}

			if len(rs.Rejections) == 0 {
				return errors.New("no rejections")
			}
			return nil
		}
	}
	return fmt.Errorf("did not receive status %d within %v, last status: %q, from %s", http.StatusOK, timeout, statusCode, url)
}

func (m *SettlementsAPIClient) GetCurrencyReport(ctx context.Context) error {
	var statusCode int
	url := m.baseURL + "/payments/currencies-report"

	endTime := time.Now().Add(timeout)
	for time.Until(endTime) > 0 {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+m.token)

		res, err := m.client.Do(req)
		if err != nil {
			return err
		}
		statusCode = res.StatusCode

		if res.StatusCode != http.StatusOK {
			errBody, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			zapctx.Info(ctx, "response status code",
				zap.Int("status_code", res.StatusCode),
				zap.String("url", url),
				zap.String("err_message", string(errBody)),
			)

			continue
		} else {
			_ = res.Body.Close()
			return nil
		}
	}
	return fmt.Errorf("did not receive status %d within %v, last status: %q, from %s", http.StatusOK, timeout, statusCode, url)
}

func (m *SettlementsAPIClient) GetPaymentByMid(ctx context.Context, mid string) (models.PaymentInstruction, error) {
	dateString := time.Now().Format("2006-01-02")
	url := m.baseURL + "/mid/" + mid + "/" + dateString

	endTime := time.Now().Add(timeout)
	for time.Until(endTime) > 0 {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return models.PaymentInstruction{}, err
		}
		req.Header.Set("Authorization", "Bearer "+m.token)

		res, err := m.client.Do(req)
		if err != nil {
			return models.PaymentInstruction{}, err
		}

		if res.StatusCode != http.StatusOK {
			errBody, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			zapctx.Info(ctx, "response status code",
				zap.Int("status_code", res.StatusCode),
				zap.String("url", url),
				zap.String("err_message", string(errBody)),
			)

			wait()
			continue
		}

		zapctx.Info(ctx, "response status code",
			zap.Int("status_code", res.StatusCode),
			zap.String("url", url),
		)

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return models.PaymentInstruction{}, fmt.Errorf("could not read the body of the response %v: %w", body, err)
		}
		_ = res.Body.Close()
		pi, err := models.NewPaymentInstructionFromJSON(body)
		if err != nil {
			return models.PaymentInstruction{}, err
		}

		return pi, nil
	}
	return models.PaymentInstruction{}, nil
}

func (m SettlementsAPIClient) PutUfxFileIntoS3Bucket(ufx models2.Ufx) error {
	ufxXML, err := xml.Marshal(ufx)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, m.baseURL+"/test/upload/file", bytes.NewReader(ufxXML))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+m.token)

	res, err := m.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("expected a %d status code but got %d instead", http.StatusOK, res.StatusCode)
	}

	return nil
}

func (m *SettlementsAPIClient) WaitUntilPaymentIsInState(ctx context.Context, client SettlementsAPIClient, mid string, expectedState models.PaymentInstructionStatus) (models.PaymentInstruction, error) {
	zapctx.Info(ctx, "waiting for payment", zap.String("mid", mid))

	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-timeoutTimer.C:
			return models.PaymentInstruction{}, fmt.Errorf("did not get payment with given state (%s) within given timeout (%s)", expectedState, timeout.String())
		case <-ticker.C:
			payment, err := client.GetPaymentByMid(ctx, mid)
			if err != nil {
				zapctx.Info(ctx, "got an error when getting payment details from SPS, trying again later...", zap.Error(err), zap.String("mid", mid))

				continue
			}
			if payment.GetStatus() != expectedState {
				zapctx.Info(ctx, "got the payment, but not with expected state, trying again later...", zap.String("mid", mid), zap.String("paymentState", string(payment.GetStatus())), zap.String("expectedState", string(expectedState)))

				continue
			}
			zapctx.Info(ctx, "received payment with expected state", zap.String("mid", mid), zap.String("paymentState", string(payment.GetStatus())))

			return payment, nil
		}
	}
}

func wait() {
	time.Sleep(waitTime)
}
