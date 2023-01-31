//go:generate moq -out mocks/payment_repo_obs_moq.go -pkg=mocks . PaymentRepoObservability

package payment_store

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentRepoObservability interface {
	StoreSuccessful(ctx context.Context, id models.PaymentInstructionID, responseTime int64)
	ReceivedStoreInstruction(ctx context.Context, id models.PaymentInstructionID)
	ReceivedGetInstruction(ctx context.Context, id models.PaymentInstructionID)
	FailedStore(ctx context.Context, id models.PaymentInstructionID, contractNumber string, err error)
	FailedUpdate(ctx context.Context, id models.PaymentInstructionID, err error)
	UpdateSuccessful(ctx context.Context, _ models.PaymentInstructionID, responseTime int64)
	FailedGet(ctx context.Context, id models.PaymentInstructionID, err error)
	GetSuccessful(ctx context.Context, id models.PaymentInstructionID, responseTime int64)
	PaymentInstructionNotFound(ctx context.Context, id models.PaymentInstructionID)
	GotReport(ctx context.Context, duration int64)
}
