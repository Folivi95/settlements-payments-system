//go:build unit
// +build unit

package functional_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
	validationMocks "github.com/saltpay/settlements-payments-system/internal/domain/validation/mocks"
)

func TestTrackPaymentOutcome_Execute(t *testing.T) {
	dummyEventValidator := &validationMocks.PPEventValidatorMock{ValidateFunc: func(ppEvent models.PaymentProviderEvent) error {
		return nil
	}}

	t.Run("given a successful payment, we update the status, events and version of the payment instruction", func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomSuccessfulPaymentProviderEvent()
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{
				StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				},
				UpdatePaymentFunc: func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
					return nil
				},
			}
			paymentExporterProducer = &mocks.PaymentExporterProducerMock{}
		)

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)

		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Successful)
		assert.Equal(t, updatedEvents.Type, models.DomainProcessingSucceeded)
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.StuckInPending), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.StuckInPending)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}
		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainProcessingFailedEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.StuckPayment,
				Message: ppEvent.FailureReason.Message,
			},
			PaymentProviderID: ppEvent.PaymentProviderPaymentID,
		})
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.RejectedPayment), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.RejectedCode)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainRejectedFailedEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.RejectedPayment,
				Message: ppEvent.FailureReason.Message,
			},
			PaymentProviderID: ppEvent.PaymentProviderPaymentID,
		})
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.UnhandledPaymentProviderStatus), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.UnhandledPaymentProviderStatus)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainUnhandledStatusFailedEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.Unhandled,
				Message: ppEvent.FailureReason.Message,
			},
			PaymentProviderID: ppEvent.PaymentProviderPaymentID,
		})
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.TransportFailure), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.TransportFailure)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainTransportErrorEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.TransportMishap,
				Message: ppEvent.FailureReason.Message,
			},
			PaymentProviderID: ppEvent.PaymentProviderPaymentID,
		})
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.NoSourceAccount), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.NoSourceAccount)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainRejectedEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.NoSourceAcct,
				Message: ppEvent.FailureReason.Message,
			},
		})
	})

	t.Run(fmt.Sprintf("Given a failed ppEvent with a faliure code of %s, it updates the status of the payment instruction, adds the relevant event and tracks it in the payment repo", models.MissingFunding), func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.MissingFunding)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, dummyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		require.NoError(t, err)

		assert.Equal(t, len(mockPaymentInstructionRepo.UpdatePaymentCalls()), 1)
		updatedStatus := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Status
		updatedEvents := mockPaymentInstructionRepo.UpdatePaymentCalls()[0].Event
		assert.Equal(t, updatedStatus, models.Failed)

		assert.Equal(t, updatedEvents.Type, models.DomainProcessingFailed)
		assert.Equal(t, updatedEvents.Details, models.DomainRejectedEventDetails{
			FailureReason: models.PIFailureReason{
				Code:    models.MissingFunds,
				Message: ppEvent.FailureReason.Message,
			},
		})
	})

	t.Run("errors if payment provider event validation fails", func(t *testing.T) {
		var (
			ctx                        = context.Background()
			ppEvent                    = randomFailedPaymentProviderEvent(models.StuckInPending)
			mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{}
			paymentExporterProducer    = &mocks.PaymentExporterProducerMock{}
		)

		mockPaymentInstructionRepo.StoreFunc = func(ctx context.Context, instruction models.PaymentInstruction) error {
			return nil
		}
		mockPaymentInstructionRepo.UpdatePaymentFunc = func(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
			return nil
		}

		paymentExporterProducer.ReportPaymentStatusFunc = func(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
			return nil
		}

		validationError := errors.New("bad event")
		spyEventValidator := &validationMocks.PPEventValidatorMock{ValidateFunc: func(ppEvent models.PaymentProviderEvent) error {
			return validationError
		}}

		useCase := use_cases.NewTrackPaymentOutcome(mockPaymentInstructionRepo, spyEventValidator, paymentExporterProducer)
		err := useCase.Execute(ctx, ppEvent)
		assert.Equal(t, err, validationError)
		assert.Equal(t, spyEventValidator.ValidateCalls()[0].PaymentProviderEvent, ppEvent)
	})
}

func randomSuccessfulPaymentProviderEvent() models.PaymentProviderEvent {
	instruction, _, _ := testhelpers.ValidPaymentInstruction()
	(&instruction).SubmitForProcessing()
	ppEvent := models.PaymentProviderEvent{
		CreatedOn:                time.Now(),
		Type:                     models.Processed,
		PaymentInstruction:       instruction,
		PaymentProviderName:      models.BC,
		PaymentProviderPaymentID: "banking_circle_1234567890",
		FailureReason:            models.FailureReason{},
	}
	return ppEvent
}

func randomFailedPaymentProviderEvent(code models.PPEventFailureCode) models.PaymentProviderEvent {
	instruction, _, _ := testhelpers.ValidPaymentInstruction()
	(&instruction).SubmitForProcessing()
	ppEvent := models.PaymentProviderEvent{
		CreatedOn:                time.Now(),
		Type:                     models.Failure,
		PaymentInstruction:       instruction,
		PaymentProviderName:      models.BC,
		PaymentProviderPaymentID: "banking_circle_1234567890",
		FailureReason: models.FailureReason{
			Code:    code,
			Message: "blah",
		},
	}
	return ppEvent
}
