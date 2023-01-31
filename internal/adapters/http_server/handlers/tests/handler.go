package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws"
)

type Handler struct {
	ufxFileUploader aws.UfxFileUploader
}

func NewHandler(ufxFileUploader aws.UfxFileUploader) *Handler {
	return &Handler{
		ufxFileUploader: ufxFileUploader,
	}
}

func (h *Handler) AddFileToS3Bucket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zapctx.Error(ctx, "error reading body", zap.Error(err))
		return
	}

	err = h.ufxFileUploader.AddFileToS3Bucket(ctx, bytes.NewReader(file), fmt.Sprintf("ufx_file_%s", uuid.New().String()))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		zapctx.Error(ctx, "error putting the file into s3", zap.Error(err))
		return
	}
}
