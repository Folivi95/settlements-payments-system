package ufx_file_listener

import (
	"context"
	"io"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type UfxConverter interface {
	ConvertUfx(ctx context.Context, ufxFileContents io.Reader, ufxFileName string) (models.IncomingInstructions, error)
}
