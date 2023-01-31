package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
	"github.com/pkg/errors"

	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
)

func TestReplayPaymentHandler_ReplayMissingFundsPayments(t *testing.T) {
	is := is.New(t)

	t.Run("receives an action, currency and file and call the replay payment execute", func(t *testing.T) {
		// given an action "pay_currency_from_file", a currency code "RON" and a file "OIC_SAXO_HR_BORGUN_20211101_1.xml"
		var (
			action   = "pay_currency_from_file"
			currency = "RON"
			file     = "OIC_SAXO_HR_BORGUN_20211101_1.xml"
		)

		replayPayment := &mocks.ReplayPaymentMock{ExecuteFunc: func(ctx context.Context, currency string, file string) error {
			return nil
		}}

		// when we call the endpoint POST /payments?action=pay_currency_from_file&currency=GBP&file=OIC_SAXO_HR_BORGUN_20211101_1.xml
		handler := handlers.NewReplayPaymentHandler(replayPayment)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments?action=%s&currency=%s&file=%s", action, currency, file), nil)
		res := httptest.NewRecorder()

		handler.ReplayMissingFundsPayments(res, req)

		is.Equal(res.Code, http.StatusOK)
	})

	t.Run("only calls the replay payment when the action is pay_currency_from_file", func(t *testing.T) {
		var (
			action   = "something_else"
			currency = "RON"
			file     = "OIC_SAXO_HR_BORGUN_20211101_1.xml"
		)

		replayPayment := &mocks.ReplayPaymentMock{ExecuteFunc: func(ctx context.Context, currency string, file string) error {
			return nil
		}}

		handler := handlers.NewReplayPaymentHandler(replayPayment)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments?action=%s&currency=%s&file=%s", action, currency, file), nil)
		res := httptest.NewRecorder()

		handler.ReplayMissingFundsPayments(res, req)

		spyExecute := replayPayment.ExecuteCalls()
		callsToExecute := len(spyExecute)
		is.Equal(callsToExecute, 0)
	})

	t.Run("returns 500 if the replay payment fails", func(t *testing.T) {
		var (
			action        = "pay_currency_from_file"
			currency      = "RON"
			file          = "OIC_SAXO_HR_BORGUN_20211101_1.xml"
			expectedError = errors.New("something went wrong")
		)

		replayPayment := &mocks.ReplayPaymentMock{ExecuteFunc: func(ctx context.Context, currency string, file string) error {
			return expectedError
		}}

		handler := handlers.NewReplayPaymentHandler(replayPayment)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments?action=%s&currency=%s&file=%s", action, currency, file), nil)
		res := httptest.NewRecorder()

		handler.ReplayMissingFundsPayments(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)
	})

	t.Run("returns 400", func(t *testing.T) {
		testData := []struct {
			action    string
			currency  string
			file      string
			testTitle string
		}{
			{"", "RON", "OIC_SAXO_HR_BORGUN_20211101_1.xml", "when the action is empty"},
			{"pay_currency_from_file", "", "OIC_SAXO_HR_BORGUN_20211101_1.xml", "when the currency is empty"},
			{"pay_currency_from_file", "RON", "", "when the file name is empty"},
		}

		replayPayment := &mocks.ReplayPaymentMock{ExecuteFunc: func(ctx context.Context, currency string, file string) error {
			return nil
		}}

		handler := handlers.NewReplayPaymentHandler(replayPayment)

		for _, test := range testData {
			t.Run(test.testTitle, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/payments?action=%s&currency=%s&file=%s", test.action, test.currency, test.file), nil)
				res := httptest.NewRecorder()

				handler.ReplayMissingFundsPayments(res, req)

				is.Equal(res.Code, http.StatusBadRequest)
			})
		}
	})
}
