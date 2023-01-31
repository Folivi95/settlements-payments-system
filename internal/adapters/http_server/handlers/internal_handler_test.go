//go:build unit
// +build unit

package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	mocks2 "github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

const (
	presignedURL = "https://localhost/s3/bucket/dummy"
)

func TestDeadLetterQueueHandler(t *testing.T) {
	t.Run("view a payment instruction inside a dead letter queue", func(t *testing.T) {
		is := is.New(t)
		message, err := testhelpers.NewPaymentInstructionBuilder().Build().MustToJSON()
		is.NoErr(err)
		testPaymentMessage := string(message)

		expectedDLQInfo := sqs.DLQInformation{Count: 1, Messages: []string{testPaymentMessage}}

		spyQueueMock := mocks2.QueueMock{PeekAllMessagesFunc: func(context.Context) (sqs.DLQInformation, error) {
			return expectedDLQInfo, nil
		}}

		dlqMapping := sqs.QueueClientMapping{sqs.BcUnprocessedPayments: &spyQueueMock}

		handler := handlers.NewInternalHandler(dlqMapping, false, ufx_downloader.UfxDownloader{})

		r := mux.NewRouter()
		r.HandleFunc("/internal/dead-letter-queues/{name}", handler.GetDlqInformation).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/internal/dead-letter-queues/bc-unprocessed", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		dlqInfo, err := sqs.NewDLQInformationFromJSON(res.Body)

		is.NoErr(err)
		is.Equal(dlqInfo.Count, expectedDLQInfo.Count)
		is.Equal(dlqInfo.Messages, expectedDLQInfo.Messages)
	})

	t.Run("return a not found if queue name doesn't exist", func(t *testing.T) {
		is := is.New(t)

		dlqMapping := sqs.QueueClientMapping{}

		handler := handlers.NewInternalHandler(dlqMapping, false, ufx_downloader.UfxDownloader{})

		r := mux.NewRouter()
		r.HandleFunc("/internal/dead-letter-queues/{name}", handler.GetDlqInformation).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/internal/dead-letter-queues/DOESNT-EXIST", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusNotFound)
	})

	t.Run("When a dead letter queue fails return status 500", func(t *testing.T) {
		is := is.New(t)

		message, err := testhelpers.NewPaymentInstructionBuilder().Build().MustToJSON()
		is.NoErr(err)
		testPaymentMessage := string(message)

		spyQueueMock := mocks2.QueueMock{PeekAllMessagesFunc: func(context.Context) (sqs.DLQInformation, error) {
			return sqs.DLQInformation{Count: 1, Messages: []string{testPaymentMessage}}, testhelpers2.RandomError()
		}}

		dlqMapping := sqs.QueueClientMapping{sqs.BcUnprocessedPayments: &spyQueueMock}

		handler := handlers.NewInternalHandler(dlqMapping, false, ufx_downloader.UfxDownloader{})

		r := mux.NewRouter()
		r.HandleFunc("/internal/dead-letter-queues/{name}", handler.GetDlqInformation).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, "/internal/dead-letter-queues/bc-unprocessed", nil)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusInternalServerError)
	})
}

func TestPaymentHandler_GetDLQURLs(t *testing.T) {
	const (
		baseURL   = "/internal/dead-letter-queues"
		localHost = "http://localhost:8080"
	)
	h := handlers.NewInternalHandler(
		sqs.QueueClientMapping{sqs.ProcessedPaymentsDlq: nil, sqs.BcUnprocessedPaymentsDlq: nil},
		false,
		ufx_downloader.UfxDownloader{},
	)

	t.Run("Get all DLQ URLs", func(t *testing.T) {
		is := is.New(t)
		processed := fmt.Sprintf(`<a href="%v/%v" target="_blank"> %v</a> <br>`, localHost+baseURL, sqs.ProcessedPaymentsDlq, sqs.ProcessedPaymentsDlq)
		unprocessed := fmt.Sprintf(`<a href="%v/%v" target="_blank"> %v</a> <br>`, localHost+baseURL, sqs.BcUnprocessedPaymentsDlq, sqs.BcUnprocessedPaymentsDlq)

		r := mux.NewRouter()
		r.HandleFunc(baseURL, h.GetDlqUrls).Methods(http.MethodGet)

		req := httptest.NewRequest(http.MethodGet, localHost+baseURL, http.NoBody)
		res := httptest.NewRecorder()

		r.ServeHTTP(res, req)

		is.Equal(res.Code, http.StatusOK)

		body, readAllErr := io.ReadAll(res.Body)
		is.NoErr(readAllErr)
		is.True(strings.Contains(string(body), processed))
		is.True(strings.Contains(string(body), unprocessed))
	})
}

func TestUfxDownloader(t *testing.T) {
	is := is.New(t)

	const (
		baseURL   = "/internal/ufx-file/"
		localHost = "http://localhost:8080"
	)

	todayFileScenarios := []struct {
		testTitle  string
		filetype   string
		statusCode int
		listReturn []*s3.Object
		serveCalls int
	}{
		{
			testTitle: "Download 'todays' main file",
			filetype:  "main",
			listReturn: []*s3.Object{
				{
					Key: aws.String(fmt.Sprint("OIC_Documents_SAXO_BORGUN_", time.Now().Format("20060102"), "_1")),
				},
			},
			statusCode: http.StatusOK,
			serveCalls: 1,
		},
		{
			testTitle: "Download 'todays' high-risk file",
			filetype:  "high-risk",
			listReturn: []*s3.Object{
				{
					Key: aws.String(fmt.Sprint("OIC_Documents_SAXO_HR_BORGUN_", time.Now().Format("20060102"), "_1")),
				},
			},
			statusCode: http.StatusOK,
			serveCalls: 1,
		},
		{
			testTitle:  "No files matching 'todays' high-risk file prefix",
			filetype:   "high-risk",
			listReturn: []*s3.Object{},
			statusCode: http.StatusNotFound,
			serveCalls: 0,
		},
		{
			testTitle:  "No files matching 'todays' main file prefix",
			filetype:   "main",
			listReturn: []*s3.Object{},
			statusCode: http.StatusNotFound,
			serveCalls: 0,
		},
	}

	for _, scenario := range todayFileScenarios {
		t.Run(scenario.testTitle, func(t *testing.T) {
			var (
				mockedS3 = generateS3Mock(scenario.listReturn, nil)
				h        = handlers.NewInternalHandler(
					nil,
					false,
					ufx_downloader.New(
						mockedS3,
					),
				)
			)

			r := mux.NewRouter()
			r.HandleFunc(fmt.Sprint(baseURL, "{filetype}"), h.DownloadUfxFile).Methods(http.MethodGet)

			reqUrl := fmt.Sprint(localHost, baseURL, scenario.filetype)
			req := httptest.NewRequest(http.MethodGet, reqUrl, http.NoBody)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)
			is.Equal(res.Code, scenario.statusCode)

			spyServeFile := mockedS3.GetPresignedURLCalls()
			is.True(len(spyServeFile) == scenario.serveCalls)
			if scenario.serveCalls > 0 {
				is.Equal(spyServeFile[0].Filename, aws.StringValue(scenario.listReturn[0].Key))
			}
		})
	}

	filenameSpecifiedScenarios := []struct {
		testTitle  string
		filename   string
		statusCode int
		expError   error
	}{
		{
			testTitle:  "Successfully download specified file",
			filename:   "OIC_Documents_SAXO_BORGUN_20211124_1.xml",
			statusCode: http.StatusOK,
			expError:   nil,
		},
		{
			testTitle:  "Return 404 error if specified file is not found",
			filename:   "OIC_Documents_SAXO_BORGUN_20211124_1.xml",
			statusCode: http.StatusNotFound,
			expError:   awserr.New(s3.ErrCodeNoSuchKey, "The specified key does not exist", fmt.Errorf("")),
		},
		{
			testTitle:  "Return 500 error if bucket is not found",
			filename:   "OIC_Documents_SAXO_BORGUN_20211124_1.xml",
			statusCode: http.StatusInternalServerError,
			expError:   awserr.New(s3.ErrCodeNoSuchBucket, "The specified bucket does not exist", fmt.Errorf("")),
		},
	}

	for _, scenario := range filenameSpecifiedScenarios {
		t.Run(scenario.testTitle, func(t *testing.T) {
			var (
				mockedS3 = generateS3Mock([]*s3.Object{}, scenario.expError)
				h        = handlers.NewInternalHandler(
					nil,
					false,
					ufx_downloader.New(
						mockedS3,
					),
				)
			)

			r := mux.NewRouter()
			r.HandleFunc(fmt.Sprint(baseURL, "{filetype}"), h.DownloadUfxFile).Methods(http.MethodGet)

			reqUrl := fmt.Sprint(localHost, baseURL, scenario.filename)
			req := httptest.NewRequest(http.MethodGet, reqUrl, http.NoBody)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)
			is.Equal(res.Code, scenario.statusCode)

			spyServeFile := mockedS3.GetPresignedURLCalls()
			is.True(len(spyServeFile) == 1)
			is.Equal(spyServeFile[0].Filename, scenario.filename)
		})
	}

	badRequestScenarios := []struct {
		testTitle string
		filetype  string
		date      string
	}{
		{
			"Invalid date separator passed",
			"main",
			"2021_11_24",
		},
		{
			"Invalid date structure passed",
			"main",
			"2021",
		},
		{
			"Invalid date order passed",
			"main",
			"24-11-2021",
		},
		{
			"Invalid filetype passed",
			"main-risk",
			"2021-11-24",
		},
	}

	for _, scenario := range badRequestScenarios {
		t.Run(scenario.testTitle, func(t *testing.T) {
			var (
				mockedS3 = generateS3Mock([]*s3.Object{}, nil)
				h        = handlers.NewInternalHandler(
					nil,
					false,
					ufx_downloader.New(
						mockedS3,
					),
				)
			)

			r := mux.NewRouter()
			r.HandleFunc(fmt.Sprint(baseURL, "{filetype}/{date}"), h.DownloadUfxFile).Methods(http.MethodGet)

			reqUrl := fmt.Sprint(localHost, baseURL, scenario.filetype, "/", scenario.date)
			req := httptest.NewRequest(http.MethodGet, reqUrl, http.NoBody)
			res := httptest.NewRecorder()

			r.ServeHTTP(res, req)
			is.Equal(res.Code, http.StatusBadRequest)

			spyServeFile := mockedS3.GetPresignedURLCalls()
			is.True(len(spyServeFile) == 0)
		})
	}
}

func generateS3Mock(listObjects []*s3.Object, serveError error) *mocks.S3ClientMock {
	return &mocks.S3ClientMock{
		ListObjectsV2Func: func(ctx context.Context, prefix string) ([]*s3.Object, error) {
			return listObjects, nil
		},
		GetPresignedURLFunc: func(ctx context.Context, filename string) (string, error) {
			if serveError == nil {
				return presignedURL, nil
			}
			return "", serveError
		},
	}
}
