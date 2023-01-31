//go:build unit
// +build unit

package use_cases_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
	mocks2 "github.com/saltpay/settlements-payments-system/internal/domain/validation/mocks"
)

func TestMakePayment_RoundHUFPaymentsToNextInteger(t *testing.T) {
	assert := assert.New(t)

	// Given
	mockMetricsClient := newEmptyMetricsClientMock()
	mockPaymentInstructionRepo := newPaymentInstructionRepoMock()
	mockValidator := alwaysValidValidatorMock()

	var sentPaymentInstruction models.PaymentInstruction
	mockPaymentSender := &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(ctx context.Context, paymentInstruction models.PaymentInstruction) error {
		sentPaymentInstruction = paymentInstruction
		return nil
	}}

	makePayment := use_cases.NewMakePayment(mockMetricsClient, mockPaymentSender, mockPaymentInstructionRepo, mockValidator)

	hufAmounts := []struct {
		originalAmount              string
		expectedAmountAfterRounding string
	}{
		{
			originalAmount:              "0",
			expectedAmountAfterRounding: "0",
		},
		{
			originalAmount:              "0.1",
			expectedAmountAfterRounding: "1",
		},
		{
			originalAmount:              "93456",
			expectedAmountAfterRounding: "93456",
		},
		{
			originalAmount:              "1.1",
			expectedAmountAfterRounding: "2",
		},
		{
			originalAmount:              "606.01",
			expectedAmountAfterRounding: "607",
		},
		{
			originalAmount:              "100662.99",
			expectedAmountAfterRounding: "100663",
		},
		{
			originalAmount:              "1704.51",
			expectedAmountAfterRounding: "1705",
		},
		{
			originalAmount:              "1360369.61",
			expectedAmountAfterRounding: "1360370",
		},
	}

	for _, hufAmount := range hufAmounts {
		t.Run("round HUF payments up", func(t *testing.T) {
			paymentInstruction := createHUFPaymentInstructionWithAmount(hufAmount.originalAmount)
			// When
			_, err := makePayment.Execute(context.Background(), paymentInstruction.IncomingInstruction)
			assert.NoError(err)

			// Then
			assert.Equal(hufAmount.expectedAmountAfterRounding, sentPaymentInstruction.IncomingInstruction.Payment.Amount)
		})
	}
}

func alwaysValidValidatorMock() *mocks2.ValidatorMock {
	return &mocks2.ValidatorMock{ValidateIncomingInstructionFunc: func(incomingInstruction models.IncomingInstruction) validation.IncomingInstructionValidationResult {
		return validation.Valid
	}}
}

func newPaymentInstructionRepoMock() *mocks.StorePaymentInstructionToRepoMock {
	return &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, paymentInstruction models.PaymentInstruction) error {
		return nil
	}}
}

func newEmptyMetricsClientMock() *mocks.MetricsClientMock {
	return &mocks.MetricsClientMock{
		CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
		},
		HistogramFunc: func(ctx context.Context, name string, value float64, tags []string) {
		},
	}
}

func createHUFPaymentInstructionWithAmount(amount string) models.PaymentInstruction {
	ii := testhelpers.NewIncomingInstructionBuilder().
		WithPaymentCurrencyISOCode(models.HUF).
		WithPaymentAmount(amount).
		Build()

	return testhelpers.NewPaymentInstructionBuilder().
		WithIncomingInstruction(ii).
		Build()
}
