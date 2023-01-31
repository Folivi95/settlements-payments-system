package functional_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackPaymentState(t *testing.T) {
	t.Run("update the state and event of a successful payment sent by isb", func(t *testing.T) {
		// Given a payment response from isb
		var (
			ctx                  = context.Background()
			paymentInstructionID = testhelpers.RandomString()
			successfulState      = models.Successful
			event                = models.PaymentInstructionEvent{
				Type:      models.DomainProcessingSucceeded,
				CreatedOn: time.Time{},
				Details:   "",
			}
			mockStore = &mocks.StorePaymentInstructionToRepoMock{
				StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				},
				UpdatePaymentFunc: func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
					return nil
				},
			}
			mockMetricsClient = &mocks.MetricsClientMock{CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
			}}
		)
		// when we execute the update payment method
		useCase := use_cases.NewUpdatePaymentState(mockStore, mockMetricsClient)
		err := useCase.Execute(ctx, paymentInstructionID, successfulState, event)
		require.NoError(t, err)

		// then we should update the payment state and events in the DB
		assert.Len(t, mockStore.UpdatePaymentCalls(), 1)
		assert.Equal(t, models.PaymentInstructionID(paymentInstructionID), mockStore.UpdatePaymentCalls()[0].ID)
		assert.Equal(t, successfulState, mockStore.UpdatePaymentCalls()[0].Status)
		assert.Equal(t, event, mockStore.UpdatePaymentCalls()[0].Event)
		assert.Len(t, mockMetricsClient.CountCalls(), 1)
	})
	t.Run("update the state of a failed payment sent by isb", func(t *testing.T) {
		var (
			ctx                  = context.Background()
			paymentInstructionID = testhelpers.RandomString()
			failedState          = models.Failed
			event                = models.PaymentInstructionEvent{
				Type:      models.DomainProcessingFailed,
				CreatedOn: time.Time{},
				Details: models.FailureReason{
					Code:    models.MissingFunding,
					Message: testhelpers.RandomString(),
				},
			}
			mockStore = &mocks.StorePaymentInstructionToRepoMock{
				StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				},
				UpdatePaymentFunc: func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
					return nil
				},
			}
			mockMetricsClient = &mocks.MetricsClientMock{CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
			}}
		)
		// when we execute the update payment method
		useCase := use_cases.NewUpdatePaymentState(mockStore, mockMetricsClient)
		err := useCase.Execute(ctx, paymentInstructionID, failedState, event)
		require.NoError(t, err)

		// then we should update the payment state and events in the DB
		assert.Len(t, mockStore.UpdatePaymentCalls(), 1)
		assert.Equal(t, models.PaymentInstructionID(paymentInstructionID), mockStore.UpdatePaymentCalls()[0].ID)
		assert.Equal(t, failedState, mockStore.UpdatePaymentCalls()[0].Status)
		assert.Equal(t, event, mockStore.UpdatePaymentCalls()[0].Event)
		assert.Len(t, mockMetricsClient.CountCalls(), 0)
	})

	t.Run("errors if updatePaymentState call fails", func(t *testing.T) {
		var (
			ctx                  = context.Background()
			updatePaymentError   = fmt.Errorf("unable to update the payment")
			paymentInstructionID = testhelpers.RandomString()
			state                = models.Successful
			event                = models.PaymentInstructionEvent{
				Type:      models.DomainProcessingSucceeded,
				CreatedOn: time.Time{},
				Details:   "",
			}
			mockStore = &mocks.StorePaymentInstructionToRepoMock{
				StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				},
				UpdatePaymentFunc: func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
					return updatePaymentError
				},
			}
			mockMetricsClient = &mocks.MetricsClientMock{CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
			}}
		)

		useCase := use_cases.NewUpdatePaymentState(mockStore, mockMetricsClient)
		err := useCase.Execute(ctx, paymentInstructionID, state, event)
		assert.Error(t, err)
		assert.Equal(t, err, updatePaymentError)

		assert.Len(t, mockStore.UpdatePaymentCalls(), 1)
	})
}
