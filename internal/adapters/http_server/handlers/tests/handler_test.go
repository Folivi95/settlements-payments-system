package tests_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/mocks"

	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers/tests"
)

func TestHandler_AddFileToS3Bucket(t *testing.T) {
	t.Run("returns 200 when the file has been added to the bucket", func(t *testing.T) {
		var (
			mockedS3 = &mocks.S3ClientMock{PutBucketFileFunc: func(ctx context.Context, file *bytes.Reader, filename string) error {
				return nil
			}}
			ufxFileUploader = aws.NewUfxFileUploader(mockedS3)
		)

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test/upload/file", bytes.NewReader([]byte("hello world")))
		req.Header.Add("Content-Type", "multipart/form-data; boundary=---WebKitFormBoundary7MA4YWxkTrZu0gW")

		handler := tests.NewHandler(ufxFileUploader)
		handler.AddFileToS3Bucket(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})
}
