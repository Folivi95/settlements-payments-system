//go:build unit
// +build unit

package functional_tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store/postgresql"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports/mocks"
	"github.com/saltpay/settlements-payments-system/internal/domain/use_cases"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

/*
	Given a valid payment instruction
	When we attempt a payment with it
	Then the payment instruction must be submitted to a Payment Provider
	And the submitted payment instruction must have the "created" and "submitted to payment provider" events
*/

func TestMakePayment_Execute(t *testing.T) {
	t.Run("With a valid incoming instruction that should be handled by Banking Circle", func(t *testing.T) {
		t.Run("creates a payment instruction and submits it to the Banking Circle payment provider with the relevant events", func(t *testing.T) {
			var (
				ctx                          = context.Background()
				metricCountCounter           = 0
				metricTotalCounter           = 0
				incomingInstruction          = testhelpers.NewIncomingInstructionBuilder().WithPaymentAmount("50").Build()
				paymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(context.Context, models.PaymentInstruction) error {
					return nil
				}}
				mockStore = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}}

				metricsClient = &mocks.MetricsClientMock{
					CountFunc: func(ctx context.Context, name string, value int64, _ []string) {
						if name == use_cases.PaymentsAmount {
							metricTotalCounter += int(value)
						}
						if name == use_cases.PaymentsCount {
							metricCountCounter++
						}
					},
					HistogramFunc: func(context.Context, string, float64, []string) {},
				}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				metricsClient,
				paymentProviderRequestSender,
				mockStore,
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, incomingInstruction)
			require.NoError(t, err)
			assert.NotEmpty(t, id)

			paymentInstFromCall := paymentProviderRequestSender.SendPaymentInstructionCalls()[0].PaymentInstruction
			assert.Equal(t, paymentInstFromCall.GetStatus(), models.SubmittedForProcessing)
			assert.Equal(t, paymentInstFromCall.Version(), 2)
			assert.Equal(t, len(paymentInstFromCall.Events()), 2)
			assert.Equal(t, paymentInstFromCall.Events()[0].Type, models.DomainReceived)
			assert.Equal(t, paymentInstFromCall.Events()[1].Type, models.DomainSubmittedToPaymentProvider)
			assert.Equal(t, paymentInstFromCall.Events()[1].Details, models.DomainSubmittedToPaymentProviderEventDetails{
				PaymentProviderType: models.BankingCircle,
			})

			assert.Equal(t, 1, metricCountCounter)
			assert.Equal(t, 50, metricTotalCounter)
		})

		t.Run("when an account number is in the incorrect format, it increments the metric", func(t *testing.T) {
			var (
				ctx                          = context.Background()
				metricCountCounter           = 0
				incorrectAccountNumber       = "de 111 11111 gb 11111111"
				incomingInstruction          = testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber(incorrectAccountNumber).Build()
				paymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(context.Context, models.PaymentInstruction) error {
					return nil
				}}
				mockStore = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}}

				metricsClient = &mocks.MetricsClientMock{
					CountFunc: func(ctx context.Context, name string, value int64, _ []string) {
						if name == use_cases.AccountNumberIncorrectFormatCounter {
							metricCountCounter += int(value)
						}
					},
					HistogramFunc: func(context.Context, string, float64, []string) {},
				}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				metricsClient,
				paymentProviderRequestSender,
				mockStore,
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, incomingInstruction)
			assert.NoError(t, err)
			assert.NotEmpty(t, id)

			paymentInstFromCall := paymentProviderRequestSender.SendPaymentInstructionCalls()[0].PaymentInstruction
			assert.Equal(t, paymentInstFromCall.GetStatus(), models.SubmittedForProcessing)

			assert.Equal(t, metricCountCounter, 1)
		})
	})

	t.Run("With a valid incoming instruction that should be handled by Islandsbanki", func(t *testing.T) {
		t.Run("creates a payment instruction with an ISK payment and submits it to the Islandsbanki payment provider with the relevant events", func(t *testing.T) {
			var (
				ctx                = context.Background()
				metricCountCounter = 0
				metricTotalCounter = 0
				iskCurrency        = models.Currency{
					IsoCode:   "ISK",
					IsoNumber: "352",
				}
				incomingInstruction = testhelpers.NewIncomingInstructionBuilder().
							WithMerchantAccountNumber("IS140159260076545510730339").
							WithMerchantAccountSwift("").
							WithMerchantRegNumber("").
							WithMetadataSender("RB").
							WithCurrency(iskCurrency).
							WithPaymentAmount("50").
							Build()
				paymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(context.Context, models.PaymentInstruction) error {
					return nil
				}}
				metricsClient = &mocks.MetricsClientMock{
					CountFunc: func(ctx context.Context, name string, value int64, _ []string) {
						if name == use_cases.PaymentsAmount {
							metricTotalCounter += int(value)
						}
						if name == use_cases.PaymentsCount {
							metricCountCounter++
						}
					},
					HistogramFunc: func(context.Context, string, float64, []string) {},
				}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				metricsClient,
				paymentProviderRequestSender,
				&mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}},
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, incomingInstruction)
			assert.NoError(t, err)
			assert.NotEmpty(t, len(id) > 0)

			paymentInstFromCall := paymentProviderRequestSender.SendPaymentInstructionCalls()[0].PaymentInstruction
			assert.Equal(t, paymentInstFromCall.GetStatus(), models.SubmittedForProcessing)
			assert.Equal(t, paymentInstFromCall.Version(), 2)
			assert.Equal(t, len(paymentInstFromCall.Events()), 2)
			assert.Equal(t, paymentInstFromCall.Events()[0].Type, models.DomainReceived)
			assert.Equal(t, paymentInstFromCall.Events()[1].Type, models.DomainSubmittedToPaymentProvider)
			assert.Equal(t, paymentInstFromCall.Events()[1].Details, models.DomainSubmittedToPaymentProviderEventDetails{
				PaymentProviderType: models.Islandsbanki,
			})

			assert.Equal(t, 1, metricCountCounter)
			assert.Equal(t, 50, metricTotalCounter)
		})

		t.Run("creates a payment instruction with non ISK payment and submits it to the Islandsbanki payment provider with the relevant events", func(t *testing.T) {
			var (
				ctx                = context.Background()
				metricCountCounter = 0
				metricTotalCounter = 0
				nonIskCurrency     = models.Currency{
					IsoCode:   "USD",
					IsoNumber: "840",
				}
				incomingInstruction = testhelpers.NewIncomingInstructionBuilder().
							WithMerchantAccountNumber("IS140159260076545510730339").
							WithMerchantAccountSwift("").
							WithMerchantRegNumber("").
							WithMetadataSender("ISB").
							WithCurrency(nonIskCurrency).
							WithPaymentAmount("50").
							Build()

				paymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(context.Context, models.PaymentInstruction) error {
					return nil
				}}
				metricsClient = &mocks.MetricsClientMock{
					CountFunc: func(ctx context.Context, name string, value int64, _ []string) {
						if name == use_cases.PaymentsAmount {
							metricTotalCounter += int(value)
						}
						if name == use_cases.PaymentsCount {
							metricCountCounter++
						}
					},
					HistogramFunc: func(context.Context, string, float64, []string) {},
				}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				metricsClient,
				paymentProviderRequestSender,
				&mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}},
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, incomingInstruction)
			require.NoError(t, err)
			assert.NotEmpty(t, id)

			paymentInstFromCall := paymentProviderRequestSender.SendPaymentInstructionCalls()[0].PaymentInstruction
			assert.Equal(t, paymentInstFromCall.GetStatus(), models.SubmittedForProcessing)
			assert.Equal(t, paymentInstFromCall.Version(), 2)

			assert.Equal(t, len(paymentInstFromCall.Events()), 2)
			assert.Equal(t, paymentInstFromCall.Events()[0].Type, models.DomainReceived)
			assert.Equal(t, paymentInstFromCall.Events()[1].Type, models.DomainSubmittedToPaymentProvider)
			assert.Equal(t, paymentInstFromCall.Events()[1].Details, models.DomainSubmittedToPaymentProviderEventDetails{
				PaymentProviderType: models.Islandsbanki,
			})

			assert.Equal(t, metricCountCounter, 1)
			assert.Equal(t, metricTotalCounter, 50)
		})
	})

	t.Run("With a payment instruction that is invalid", func(t *testing.T) {
		t.Run("it does not submit the instruction to the Banking Circle payment provider", func(t *testing.T) {
			var (
				ctx                               = context.Background()
				incomingInstruction               = testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber("").Build()
				bankingCirclePaymentRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(context.Context, models.PaymentInstruction) error {
					return nil
				}}
				mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				&mocks.MetricsClientMock{},
				bankingCirclePaymentRequestSender,
				mockPaymentInstructionRepo,
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, incomingInstruction)
			assert.Empty(t, id)
			assert.Equal(t, len(bankingCirclePaymentRequestSender.SendPaymentInstructionCalls()), 0)

			validationResult, isValidationResult := err.(validation.IncomingInstructionValidationResult)
			assert.True(t, isValidationResult)
			assert.True(t, validationResult.Failed())
			assert.Equal(t, validationResult.GetErrors(), []string{"AccountNumber should not be empty"})
		})

		t.Run("raises a domain rejection event if validation fails", func(t *testing.T) {
			var (
				ctx                        = context.Background()
				invalidIncomingInstruction = testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber("").Build()
				mockPaymentInstructionRepo = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					return nil
				}}
			)

			makePaymentUseCase := use_cases.NewMakePayment(
				&mocks.MetricsClientMock{},
				&mocks.PaymentProviderRequestSenderMock{},
				mockPaymentInstructionRepo,
				validation.IncomingInstructionValidator{},
			)

			id, err := makePaymentUseCase.Execute(ctx, invalidIncomingInstruction)
			assert.Empty(t, id)

			instruction := mockPaymentInstructionRepo.StoreCalls()[0].Instruction
			assert.Equal(t, len(mockPaymentInstructionRepo.StoreCalls()), 1) // payment is stored

			assert.True(t, err != nil)
			assert.Equal(t, instruction.GetStatus(), models.Rejected)

			lastEvent := instruction.Events()[len(instruction.Events())-1]
			instructionJSON, _ := invalidIncomingInstruction.ToJSON()
			assert.Equal(t, lastEvent.Type, models.DomainRejected)
			assert.Equal(t, lastEvent.Details, models.DomainRejectedEventDetails{
				FailureReason: models.PIFailureReason{
					Code: models.FailedValidation,
					Message: models.InvalidPaymentInstructionError{
						UnderlyingError: err,
						InstructionJSON: string(instructionJSON),
					}.Error(),
				},
			})
		})
	})

	t.Run("return an error when the store fails", func(t *testing.T) {
		var (
			ctx                 = context.Background()
			incomingInstruction = testhelpers.NewIncomingInstructionBuilder().Build()
			errorMessage        = "unable to store payment"
			mockStore           = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
				return fmt.Errorf(errorMessage)
			}}
			mockMetricsClient = &mocks.MetricsClientMock{
				CountFunc:     nil,
				HistogramFunc: nil,
			}
			mockPaymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: nil}
		)
		makePayment := use_cases.NewMakePayment(mockMetricsClient, mockPaymentProviderRequestSender, mockStore, validation.IncomingInstructionValidator{})

		_, err := makePayment.Execute(ctx, incomingInstruction)
		assert.EqualError(t, err, errorMessage)
	})

	t.Run("return specific error when there is a duplicate", func(t *testing.T) {
		var (
			ctx                    = context.Background()
			incomingInstructionOne = testhelpers.NewIncomingInstructionBuilder().Build()
			incomingInstructionTwo = testhelpers.NewIncomingInstructionBuilder().Build()
			errMessage             = "duplicate"
			mockStoreOne           = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
				return nil
			}}
			counter      = 0
			mockStoreTwo = &mocks.StorePaymentInstructionToRepoMock{
				StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
					counter++
					if counter == 1 {
						return postgresql.ErrDuplicate
					}
					return nil
				},
			}
			mockMetricsClient = &mocks.MetricsClientMock{
				CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
				},
				HistogramFunc: func(ctx context.Context, name string, value float64, tags []string) {
				},
			}
			mockPaymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(ctx context.Context, paymentInstruction models.PaymentInstruction) error {
				return nil
			}}
		)

		makePaymentOne := use_cases.NewMakePayment(mockMetricsClient, mockPaymentProviderRequestSender, mockStoreOne, validation.IncomingInstructionValidator{})
		_, err := makePaymentOne.Execute(ctx, incomingInstructionOne)
		require.NoError(t, err)

		makePaymentTwo := use_cases.NewMakePayment(mockMetricsClient, mockPaymentProviderRequestSender, mockStoreTwo, validation.IncomingInstructionValidator{})
		_, err = makePaymentTwo.Execute(ctx, incomingInstructionTwo)
		assert.EqualError(t, err, errMessage)
	})

	t.Run("when we get a duplicate error store that duplicate as a failed payment", func(t *testing.T) {
		var (
			ctx                 = context.Background()
			incomingInstruction = testhelpers.NewIncomingInstructionBuilder().Build()
			mockMetricsClient   = &mocks.MetricsClientMock{
				CountFunc: func(ctx context.Context, name string, value int64, tags []string) {
				},
				HistogramFunc: func(ctx context.Context, name string, value float64, tags []string) {
				},
			}
			mockPaymentProviderRequestSender = &mocks.PaymentProviderRequestSenderMock{SendPaymentInstructionFunc: func(ctx context.Context, paymentInstruction models.PaymentInstruction) error {
				return nil
			}}
			mockPaymentStore = &mocks.StorePaymentInstructionToRepoMock{StoreFunc: func(ctx context.Context, instruction models.PaymentInstruction) error {
				return postgresql.ErrDuplicate
			}}
		)

		makePayment := use_cases.NewMakePayment(mockMetricsClient, mockPaymentProviderRequestSender, mockPaymentStore, validation.IncomingInstructionValidator{})
		_, err := makePayment.Execute(ctx, incomingInstruction)
		require.Error(t, err)

		assert.Len(t, mockPaymentStore.StoreCalls(), 2)
		assert.Equal(t, models.Failed, mockPaymentStore.StoreCalls()[1].Instruction.GetStatus())
		failedPaymentEvents := mockPaymentStore.StoreCalls()[1].Instruction.Events()
		assert.NotEmpty(t, failedPaymentEvents, "payment events should not be empty")
		assert.Equal(t, models.DomainProcessingFailed, failedPaymentEvents[len(failedPaymentEvents)-1].Type)
	})
}
