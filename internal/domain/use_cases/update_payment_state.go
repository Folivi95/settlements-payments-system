package use_cases

import (
	"context"

	zapctx "github.com/saltpay/go-zap-ctx"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"

	"go.uber.org/zap"
)

type UpdatePaymentState struct {
	paymentStore ports.StorePaymentInstructionToRepo
	metrics      ports.MetricsClient
}

func NewUpdatePaymentState(paymentStore ports.StorePaymentInstructionToRepo, metrics ports.MetricsClient) *UpdatePaymentState {
	return &UpdatePaymentState{
		paymentStore: paymentStore,
		metrics:      metrics,
	}
}

func (tps *UpdatePaymentState) Execute(ctx context.Context, paymentInstructionID string, state models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
	err := tps.paymentStore.UpdatePayment(ctx, models.PaymentInstructionID(paymentInstructionID), state, event)
	if err != nil {
		zapctx.Error(ctx, "Unable to execute update payment state", zap.Error(err), zap.String("paymentInstructionID", paymentInstructionID), zap.String("state", string(state)))
		return err
	}

	zapctx.Info(ctx, "flow_step #12a: Payment is successfully processed by payment provider", zap.String("id", paymentInstructionID))

	if state != models.Successful && state != models.StateSubmitted {
		zapctx.Error(ctx, "flow_step #12b: Payment is failed to be processed by the payment provider",
			zap.String("id", paymentInstructionID),
			zap.String("state", string(state)),
			zap.String("eventType", string(event.Type)),
			zap.Any("details", event.Details),
		)
	}

	if state == models.Successful {
		tps.metrics.Count(ctx, "app_settlements_provider_success_payment", 1, []string{"islandsbanki"})
	}

	return nil
}
