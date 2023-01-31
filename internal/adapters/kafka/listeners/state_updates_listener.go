//go:generate moq -out internal/mocks/update_payment_state_mock.go -pkg=mocks ../../../domain/ports UpdatePaymentState

package listeners

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saltpay/go-kafka-driver"

	"github.com/saltpay/settlements-payments-system/internal/adapters/kafka/listeners/internal/dto"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type StateUpdatesListener struct {
	consumer           Consumer
	updatePaymentState ports.UpdatePaymentState
}

func NewStateUpdatesListener(consumer Consumer, updatePaymentState ports.UpdatePaymentState) *StateUpdatesListener {
	return &StateUpdatesListener{
		consumer:           consumer,
		updatePaymentState: updatePaymentState,
	}
}

func (l *StateUpdatesListener) Listen(ctx context.Context) {
	l.consumer.Listen(ctx, l.processor, kafka.AlwaysCommitWithoutError, kafka.NeverPause)
}

func (l *StateUpdatesListener) processor(ctx context.Context, message kafka.Message) error {
	var paymentStatus dto.PaymentStateUpdate

	err := json.Unmarshal(message.Value, &paymentStatus)
	if err != nil {
		return fmt.Errorf("error unmarshalling the message: %w", err)
	}

	err = l.updatePaymentState.Execute(ctx, paymentStatus.PaymentInstructionID, paymentStatus.PaymentInstructionStatus(), paymentStatus.Event())
	if err != nil {
		return fmt.Errorf("error updating state in the db: %w", err)
	}

	return nil
}
