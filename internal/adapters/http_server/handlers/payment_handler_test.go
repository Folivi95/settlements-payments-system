//go:build unit
// +build unit

package handlers_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"github.com/pkg/errors"

	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	mocks3 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store/postgresql"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

func TestPaymentHandler_PostPaymentInstructions(t *testing.T) {
	t.Run("receives a payment instruction, parses it and sends it to the payment port", func(t *testing.T) {
		is := is.New(t)

		id := models.PaymentInstructionID("someID")
		spyMakePaymentUseCase := mocks.MakePaymentMock{ExecuteFunc: func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return id, nil
		}}
		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{}
		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)

		incomingInstruction := testhelpers.NewIncomingInstructionBuilder().Build()

		req := createPaymentRequest(incomingInstruction)
		res := httptest.NewRecorder()

		handler.PostPaymentInstructions(res, req)

		is.Equal(res.Code, http.StatusCreated)
		is.Equal(len(spyMakePaymentUseCase.ExecuteCalls()), 1)

		paymentResponse, err := handlers.NewPaymentResponseFromJSON(res.Body)
		is.NoErr(err)
		is.Equal(string(paymentResponse.ID), string(id))
	})

	t.Run("returns an internal server error when the port fails to execute the instruction", func(t *testing.T) {
		is := is.New(t)

		spyMakePaymentUseCase := mocks.MakePaymentMock{ExecuteFunc: func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "", errors.New("oh no")
		}}
		incomingInstruction := testhelpers.NewIncomingInstructionBuilder().Build()

		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{}
		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)

		req := createPaymentRequest(incomingInstruction)
		res := httptest.NewRecorder()

		handler.PostPaymentInstructions(res, req)
		is.Equal(res.Code, http.StatusInternalServerError)
	})

	t.Run("returns a bad request, when bad JSON is sent, and doesnt call the port", func(t *testing.T) {
		is := is.New(t)

		spyMakePaymentUseCase := mocks.MakePaymentMock{ExecuteFunc: func(ctx context.Context, incomingInstruction models.IncomingInstruction) (models.PaymentInstructionID, error) {
			return "anything", nil
		}}

		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{}
		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader("garbage"))
		res := httptest.NewRecorder()

		handler.PostPaymentInstructions(res, req)

		is.Equal(res.Code, http.StatusBadRequest)
		is.Equal(len(spyMakePaymentUseCase.ExecuteCalls()), 0)
	})
}

func TestPaymentHandler_GetPaymentInstruction(t *testing.T) {
	t.Run("calls the get payment instruction use case and returning a payment instruction", func(t *testing.T) {
		is := is.New(t)
		expectedPaymentInstruction, _, err := testhelpers.ValidPaymentInstruction()
		is.NoErr(err)
		spyMakePaymentUseCase := mocks.MakePaymentMock{}

		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{
			ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
				return expectedPaymentInstruction, nil
			},
		}
		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/{id}", handler.GetPaymentInstruction).Methods(http.MethodGet)

		// when
		req := httptest.NewRequest(http.MethodGet, "/payments/339aec00-771c-467e-a8c0-9056c6d2580a", nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		// then

		is.Equal(res.Code, http.StatusOK)

		paymentInstruction, err := models.NewPaymentInstructionFromJSON(res.Body.Bytes())
		is.NoErr(err)

		is.Equal(paymentInstruction, expectedPaymentInstruction)
	})

	t.Run("responds with a 500 if the usecase fails", func(t *testing.T) {
		is := is.New(t)
		expectedError := errors.New("some error from the getPayment use case")

		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{
			ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
				return models.PaymentInstruction{}, expectedError
			},
		}
		spyMakePaymentUseCase := mocks.MakePaymentMock{}

		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/{id}", handler.GetPaymentInstruction).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/payments/someID", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
		is.Equal(res.Code, http.StatusInternalServerError)
	})

	t.Run("responds with a 404 if unknown ID provided", func(t *testing.T) {
		is := is.New(t)
		expectedError := postgresql.PaymentInstructionMissingError{}

		spyGetPaymentInstructionUseCase := mocks.GetPaymentInstructionMock{
			ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
				return models.PaymentInstruction{}, expectedError
			},
		}

		spyMakePaymentUseCase := mocks.MakePaymentMock{}

		handler := handlers.NewPaymentHandler(&spyMakePaymentUseCase, &spyGetPaymentInstructionUseCase, nil, nil)

		r := mux.NewRouter()
		r.HandleFunc("/payments/{id}", handler.GetPaymentInstruction).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/payments/blah", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
		is.Equal(res.Code, http.StatusNotFound)
	})
}

func TestPaymentHandler_GetReportToday(t *testing.T) {
	dummyGetPaymentUsecase := mocks.GetPaymentInstructionMock{
		ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, nil
		},
	}

	dummyMakePaymentUsecase := mocks.MakePaymentMock{}

	t.Run("it returns a report from the repo", func(t *testing.T) {
		is := is.New(t)

		expectedReport := models.PaymentReport{Stats: models.PaymentStats{
			Successful:             10,
			Failed:                 20,
			SubmittedForProcessing: 30,
		}}

		stubGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return expectedReport, nil
		}}

		req := httptest.NewRequest(http.MethodGet, "/payments/report", nil)
		res := httptest.NewRecorder()

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, stubGetPaymentReport, nil)
		handler.GetReport(res, req)

		is.Equal(res.Code, http.StatusOK)

		report, err := models.NewPaymentReportFromJSON(res.Body)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})

	t.Run("returns a 500 if get fails", func(t *testing.T) {
		is := is.New(t)

		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return models.PaymentReport{}, testhelpers2.RandomError()
		}}

		req := httptest.NewRequest(http.MethodGet, "/payments/report", nil)
		res := httptest.NewRecorder()

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		handler.GetReport(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)
	})
}

func TestPaymentHandler_GetReport(t *testing.T) {
	dummyGetPaymentUsecase := mocks.GetPaymentInstructionMock{
		ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, nil
		},
	}

	dummyMakePaymentUsecase := mocks.MakePaymentMock{}

	t.Run("it returns a report from the repo", func(t *testing.T) {
		is := is.New(t)

		expectedReport := models.PaymentReport{Stats: models.PaymentStats{
			Successful:             10,
			Failed:                 20,
			SubmittedForProcessing: 30,
		}}

		stubGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return expectedReport, nil
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, stubGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/report/{date}", handler.GetReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/payments/report/2021-10-06", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		report, err := models.NewPaymentReportFromJSON(res.Body)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})

	t.Run("returns a 500 if get fails", func(t *testing.T) {
		is := is.New(t)

		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return models.PaymentReport{}, testhelpers2.RandomError()
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/report/{date}", handler.GetReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/report/2021-10-06", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)
	})

	t.Run("returns 400 if not a date", func(t *testing.T) {
		is := is.New(t)

		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return models.PaymentReport{}, testhelpers2.RandomError()
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/report/{date}", handler.GetReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/report/foo", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusBadRequest)
	})

	t.Run("returns 404 if date not on database", func(t *testing.T) {
		is := is.New(t)

		expectedError := postgresql.ReportMissingError{}
		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetReportFunc: func(ctx context.Context, time time.Time) (models.PaymentReport, error) {
			return models.PaymentReport{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/report/{date}", handler.GetReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/report/2020-10-06", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusNotFound)
	})
}

func TestPaymentHandler_GetBCRejectionReport(t *testing.T) {
	dummyGetPaymentUsecase := mocks.GetPaymentInstructionMock{
		ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, nil
		},
	}

	dummyMakePaymentUsecase := mocks.MakePaymentMock{}

	t.Run("Get Rejection Report for a given date", func(t *testing.T) {
		is := is.New(t)
		expectedReport := models2.RejectionReport{Rejections: []models2.Rejection{
			{
				PIdChannelUser:         "",
				PTxndate:               time.Time{},
				ReportDate:             time.Time{},
				CustomerID:             "",
				Account:                "",
				AccountCurrency:        "EUR",
				ValueDate:              time.Time{},
				PaymentAmount:          1000,
				PaymentCurrency:        "EUR",
				TransferCurrency:       "",
				DestinationIban:        "",
				PaymentReferenceNumber: "payment_id",
				UserReferenceNumber:    "",
				FileReferenceNumber:    "",
				SourceType:             "",
				Status:                 "Rejected",
				StatusReason:           "Duplicate",
			},
		}}

		stubGetPaymentReport := mocks3.GetBankingCircleRejectionReportMock{ExecuteFunc: func(date string) (models2.RejectionReport, error) {
			return expectedReport, nil
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, nil, &stubGetPaymentReport)
		r := mux.NewRouter()
		r.HandleFunc("/bc-report/{date}", handler.GetBCReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/bc-report/2021-10-06", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		body, readAllErr := io.ReadAll(res.Body)
		is.NoErr(readAllErr)

		report, err := models2.NewRejectionReportFromJSON(body)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})
	t.Run("Get Rejection Report for today if date not specified", func(t *testing.T) {
		is := is.New(t)

		today := time.Now().UTC()
		expectedReport := models2.RejectionReport{Rejections: []models2.Rejection{
			{
				PIdChannelUser:         "",
				PTxndate:               today,
				ReportDate:             today,
				CustomerID:             "",
				Account:                "",
				AccountCurrency:        "EUR",
				ValueDate:              today,
				PaymentAmount:          1000,
				PaymentCurrency:        "EUR",
				TransferCurrency:       "",
				DestinationIban:        "",
				PaymentReferenceNumber: "payment_id",
				UserReferenceNumber:    "",
				FileReferenceNumber:    "",
				SourceType:             "",
				Status:                 "Rejected",
				StatusReason:           "Duplicate",
			},
		}}

		stubGetPaymentReport := mocks3.GetBankingCircleRejectionReportMock{ExecuteFunc: func(date string) (models2.RejectionReport, error) {
			return expectedReport, nil
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, nil, &stubGetPaymentReport)
		r := mux.NewRouter()
		r.HandleFunc("/bc-report", handler.GetBCReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/bc-report", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		body, readAllErr := io.ReadAll(res.Body)
		is.NoErr(readAllErr)

		report, err := models2.NewRejectionReportFromJSON(body)
		is.NoErr(err)
		is.Equal(report.Rejections[0].PTxndate, today)
	})
	t.Run("returns 500 if execute fails", func(t *testing.T) {
		is := is.New(t)
		expectedError := fmt.Errorf("banking Circle failed")
		stubGetPaymentReport := mocks3.GetBankingCircleRejectionReportMock{ExecuteFunc: func(date string) (models2.RejectionReport, error) {
			return models2.RejectionReport{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, nil, &stubGetPaymentReport)
		r := mux.NewRouter()
		r.HandleFunc("/bc-report/{date}", handler.GetBCReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/bc-report/2021-10-06", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)

		body, readAllErr := io.ReadAll(res.Body)
		is.NoErr(readAllErr)
		is.True(strings.Contains(string(body), expectedError.Error()))
	})
}

func TestPaymentHandler_GetCurrencyReport(t *testing.T) {
	dummyGetPaymentUsecase := mocks.GetPaymentInstructionMock{
		ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, nil
		},
	}

	dummyMakePaymentUsecase := mocks.MakePaymentMock{}

	t.Run("it returns a report from the repo for today", func(t *testing.T) {
		is := is.New(t)

		expectedReport := models.PaymentCurrencyReport{}

		stubGetPaymentReport := &mocks.GetPaymentReportMock{GetCurrencyReportFunc: func(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error) {
			return expectedReport, nil
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, stubGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/currencies-report", handler.GetCurrencyReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/payments/currencies-report", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		report, err := models.NewPaymentCurrencyReportFromJSON(res.Body)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})
	t.Run("It returns a report for a valid day", func(t *testing.T) {
		is := is.New(t)

		expectedReport := models.PaymentCurrencyReport{}

		stubGetPaymentReport := &mocks.GetPaymentReportMock{GetCurrencyReportFunc: func(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error) {
			return expectedReport, nil
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, stubGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/currencies-report/{date}", handler.GetCurrencyReport).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/payments/currencies-report/2021-11-15", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		report, err := models.NewPaymentCurrencyReportFromJSON(res.Body)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})
	t.Run("return 500 if errors for today", func(t *testing.T) {
		is := is.New(t)

		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetCurrencyReportFunc: func(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error) {
			return models.PaymentCurrencyReport{}, testhelpers2.RandomError()
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/currencies-report", handler.GetCurrencyReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/currencies-report", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)
	})
	t.Run("return 400 for errors for an invalid day", func(t *testing.T) {
		is := is.New(t)

		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetCurrencyReportFunc: func(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error) {
			return models.PaymentCurrencyReport{}, testhelpers2.RandomError()
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/currencies-report/{date}", handler.GetCurrencyReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/currencies-report/202-14-12", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusBadRequest)
	})
	t.Run("return 404 if day is not in the database", func(t *testing.T) {
		is := is.New(t)

		expectedError := postgresql.ReportMissingError{}
		failingGetPaymentReport := &mocks.GetPaymentReportMock{GetCurrencyReportFunc: func(ctx context.Context, time time.Time) (models.PaymentCurrencyReport, error) {
			return models.PaymentCurrencyReport{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUsecase, &dummyGetPaymentUsecase, failingGetPaymentReport, nil)
		r := mux.NewRouter()
		r.HandleFunc("/payments/currencies-report/{date}", handler.GetCurrencyReport).Methods(http.MethodGet)
		req := httptest.NewRequest(http.MethodGet, "/payments/currencies-report/2030-11-12", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusNotFound)
	})
}

func TestPaymentHandler_GetPaymentByMid(t *testing.T) {
	dummyGetPaymentUsecase := mocks.GetPaymentInstructionMock{
		ExecuteFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, nil
		},
	}
	dummyMakePaymentUseCase := mocks.MakePaymentMock{}

	t.Run("calls the get instruction by mid use case and returns a payment instruction", func(t *testing.T) {
		is := is.New(t)
		expectedPaymentInstruction, _, err := testhelpers.ValidPaymentInstruction()
		is.NoErr(err)

		spyGetInstructionByMid := mocks.GetPaymentReportMock{GetPaymentByMidFunc: func(ctx context.Context, mid string, time time.Time) (models.PaymentInstruction, error) {
			return expectedPaymentInstruction, nil
		}}
		handler := handlers.NewPaymentHandler(&dummyMakePaymentUseCase, &dummyGetPaymentUsecase, &spyGetInstructionByMid, nil)
		r := mux.NewRouter()
		r.HandleFunc("/mid/{mid}/{date}", handler.GetInstructionByMid).Methods(http.MethodGet)

		// when
		req := httptest.NewRequest(http.MethodGet, "/mid/9000000/2021-12-28", nil)
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)

		// then

		is.Equal(res.Code, http.StatusOK)

		paymentInstruction, err := models.NewPaymentInstructionFromJSON(res.Body.Bytes())
		is.NoErr(err)

		is.Equal(paymentInstruction, expectedPaymentInstruction)
	})

	t.Run("responds with a 500 if the usecase fails", func(t *testing.T) {
		is := is.New(t)
		expectedError := errors.New("some error from the report use case")

		spyGetInstructionByMid := mocks.GetPaymentReportMock{GetPaymentByMidFunc: func(ctx context.Context, mid string, time time.Time) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUseCase, &dummyGetPaymentUsecase, &spyGetInstructionByMid, nil)
		r := mux.NewRouter()
		r.HandleFunc("/mid/{mid}/{date}", handler.GetInstructionByMid).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/mid/9000000/2021-12-28", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
		is.Equal(res.Code, http.StatusInternalServerError)
	})

	t.Run("responds with a 400 if request is bad formatted", func(t *testing.T) {
		is := is.New(t)
		expectedError := errors.New("some error")

		spyGetInstructionByMid := mocks.GetPaymentReportMock{GetPaymentByMidFunc: func(ctx context.Context, mid string, time time.Time) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUseCase, &dummyGetPaymentUsecase, &spyGetInstructionByMid, nil)
		r := mux.NewRouter()
		r.HandleFunc("/mid/{mid}/{date}", handler.GetInstructionByMid).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/mid/9000000/2021-12-", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
		is.Equal(res.Code, http.StatusBadRequest)
	})

	t.Run("responds with a 404 if unknown ID provided", func(t *testing.T) {
		is := is.New(t)
		expectedError := postgresql.MidMissingError{}

		spyGetInstructionByMid := mocks.GetPaymentReportMock{GetPaymentByMidFunc: func(ctx context.Context, mid string, time time.Time) (models.PaymentInstruction, error) {
			return models.PaymentInstruction{}, expectedError
		}}

		handler := handlers.NewPaymentHandler(&dummyMakePaymentUseCase, &dummyGetPaymentUsecase, &spyGetInstructionByMid, nil)

		r := mux.NewRouter()
		r.HandleFunc("/mid/{mid}/{date}", handler.GetInstructionByMid).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/mid/foo/2021-12-28", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)
		is.Equal(res.Code, http.StatusNotFound)
	})
}

func createPaymentRequest(payment models.IncomingInstruction) *http.Request {
	paymentInstructionInputJSON, _ := payment.ToJSON()
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(paymentInstructionInputJSON))
	req.SetBasicAuth("abc", "123")
	return req
}
