//go:build unit
// +build unit

package functional_tests

import (
	"context"
	"testing"

	is2 "github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
)

func TestCheckPaymentInstruction_Execute(t *testing.T) {
	t.Run("Given a payment has been made successfully, it allows you to check the payment instruction", func(t *testing.T) {
		is := is2.New(t)
		ctx := context.Background()

		paymentInstruction, _, err := testhelpers.ValidPaymentInstruction()
		is.NoErr(err)
		mockPaymentRepo := &mocks.GetPaymentInstructionFromRepoMock{GetFunc: func(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
			return paymentInstruction, nil
		}}

		useCase := use_cases.NewCheckPaymentInstruction(mockPaymentRepo)
		paymentInstructionReceived, err := useCase.Execute(ctx, paymentInstruction.ID())

		is.NoErr(err)
		is.Equal(paymentInstruction, paymentInstructionReceived)
	})
}
