//go:build integration
// +build integration

package postgresql

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

func TestNewPaymentStore(t *testing.T) {
	t.Run("should successfully connect when database is available", func(t *testing.T) {
		pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
		_, err := NewPaymentStore(context.Background(), pgString, nil)
		assert.NoError(t, err)
	})

	t.Run("should fail when database is not available after polling cycle finishes", func(t *testing.T) {
		_, err := NewPaymentStore(context.Background(), "", nil)
		assert.True(t, err != nil)
	})
}

func TestStorePaymentInstruction(t *testing.T) {
	var (
		ctx      = context.Background()
		pgString = os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	)
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	require.NoError(t, err)

	t.Run("successfully save a valid payment instruction", func(t *testing.T) {
		paymentInstruction, _, err := testhelpers.ValidPaymentInstruction()
		require.NoError(t, err)

		err = paymentStore.Store(ctx, paymentInstruction)
		require.NoError(t, err)

		actualInstruction, err := paymentStore.Get(ctx, paymentInstruction.ID())
		assert.NoError(t, err)
		assert.Equal(t, actualInstruction.GetStatus(), paymentInstruction.GetStatus())
	})

	t.Run("a payment instruction that does not exist", func(t *testing.T) {
		_, err = paymentStore.Get(ctx, models.PaymentInstructionID(testhelpers2.RandomString()))
		assert.True(t, err != nil)
	})

	t.Run("when a payment is in the FAILED state, the replay of that payment with the same details should not trigger the duplicate detection", func(t *testing.T) {
		var (
			mid      = testhelpers2.RandomString()
			amount   = "100"
			currency = models.Currency{
				IsoCode:   models.GBP,
				IsoNumber: "826",
			}
			accountNumber = "GB33BUKB20201555555555"
			failedStatus  = models.Failed

			firstPayment  = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).WithStatus(failedStatus).Build()
			secondPayment = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).Build()
		)

		err := paymentStore.Store(ctx, firstPayment)
		require.NoError(t, err)

		err = paymentStore.Store(ctx, secondPayment)
		require.NoError(t, err)

		storedFirstPayment, err := paymentStore.Get(ctx, firstPayment.ID())
		require.NoError(t, err)

		storedSecondPayment, err := paymentStore.Get(ctx, secondPayment.ID())
		require.NoError(t, err)

		assert.NotEqual(t, storedFirstPayment.ID(), storedSecondPayment.ID(), "ids should be different")
		assert.Equal(t, storedSecondPayment.ContractNumber(), storedFirstPayment.ContractNumber(), "mids should be the same")
		assert.NotEqual(t, storedSecondPayment.GetStatus(), storedFirstPayment.GetStatus(), "states should not be the same")
		assert.NotEqual(t, storedSecondPayment.GetStatus(), models.Failed, "second payment should store without failure state")
	})

	t.Run("when a payment is in a state other than FAILED or REJECTED, the replay of that payment with the same details should trigger the duplication detection", func(t *testing.T) {
		var (
			mid      = testhelpers2.RandomString()
			amount   = "100"
			currency = models.Currency{
				IsoCode:   models.GBP,
				IsoNumber: "826",
			}
			accountNumber    = "GB33BUKB20201555555555"
			successfulStatus = models.Successful

			firstPayment  = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).WithStatus(successfulStatus).Build()
			secondPayment = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).Build()
			errMessage    = "duplicate"
		)

		err := paymentStore.Store(ctx, firstPayment)
		require.NoError(t, err)

		err = paymentStore.Store(ctx, secondPayment)
		require.Error(t, err, "duplicate payment should return an error")
		assert.EqualError(t, err, errMessage)
	})

	t.Run("when a payment is in a state other than FAILED or REJECTED, the replay of that payment with a different date should not trigger the duplication detection", func(t *testing.T) {
		var (
			mid      = testhelpers2.RandomString()
			amount   = "100"
			currency = models.Currency{
				IsoCode:   models.GBP,
				IsoNumber: "826",
			}
			accountNumber    = "GB33BUKB20201555555555"
			successfulStatus = models.Successful

			date             = time.Now().Format("2006-01-02")
			dateFormatted, _ = time.Parse("2006-01-02", date)
			tomorrow         = dateFormatted.AddDate(0, 0, 1)

			firstPayment  = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).WithStatus(successfulStatus).WithDate(dateFormatted).Build()
			secondPayment = testhelpers.NewPaymentInstructionBuilder().WithMid(mid).WithAmount(amount).WithCurrency(currency).WithAccountNumber(accountNumber).WithDate(tomorrow).Build()
		)

		err := paymentStore.Store(ctx, firstPayment)
		require.NoError(t, err)

		err = paymentStore.Store(ctx, secondPayment)
		require.NoError(t, err)
	})
}

func TestUpdatePayment(t *testing.T) {
	pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	require.NoError(t, err)
	t.Run("update the payment state, event and version for BC payments", func(t *testing.T) {
		var (
			ctx          = context.Background()
			createdOn    = time.Time{}
			storedStatus = models.SubmittedForProcessing
			storedEvents = []models.PaymentInstructionEvent{
				{
					Type:      models.DomainReceived,
					CreatedOn: createdOn,
					Details:   "",
				},
				{
					Type:      models.DomainSubmittedToPaymentProvider,
					CreatedOn: createdOn,
					Details:   "",
				},
			}
			submittedPaymentInstruction = testhelpers.NewPaymentInstructionBuilder().WithStatus(storedStatus).WithEvents(storedEvents).Build()
			expectedStatus              = models.Successful
			expectedEvents              = models.PaymentInstructionEvent{
				Type:      models.DomainProcessingSucceeded,
				CreatedOn: createdOn,
				Details:   "",
			}
		)

		err := paymentStore.Store(ctx, submittedPaymentInstruction)
		require.NoError(t, err)

		_, err = paymentStore.Get(ctx, submittedPaymentInstruction.ID())
		require.NoError(t, err)

		err = paymentStore.UpdatePayment(ctx, submittedPaymentInstruction.ID(), expectedStatus, expectedEvents)
		require.NoError(t, err)

		updatedInstruction, err := paymentStore.Get(ctx, submittedPaymentInstruction.ID())
		require.NoError(t, err)
		assert.Equal(t, expectedStatus, updatedInstruction.GetStatus())
		assert.Len(t, updatedInstruction.Events(), 3)
		assert.Equal(t, expectedEvents, updatedInstruction.Events()[2])
		assert.Equal(t, 3, updatedInstruction.Version())
	})
}

func TestPostgresStore_GetReport(t *testing.T) {
	pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	assert.NoError(t, err)
	err = paymentStore.CleanDBForTesting()
	assert.NoError(t, err)

	t.Run("returns a report of statuses for today", func(t *testing.T) {
		ctx := context.Background()

		// todo: this is HEINOUS, what do we do? this is howling how wonky the PI API
		// still is, sometimes methods like reject, sometimes events, and sometimes we
		// take a PI, sometimes incoming instruction
		pendingPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).Build(),
		).Build()
		pendingPI.SubmitForProcessing()

		successPI1 := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).Build(),
		).Build()
		successPI1.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPI1,
		})

		successPI2 := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).Build(),
		).Build()
		successPI2.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPI2,
		})

		successPIYesterday := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now().AddDate(0, 0, -1)).Build(),
		).Build()
		successPIYesterday.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPIYesterday,
		})

		rejectedValidationPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).Build(),
		).Build()
		rejectedValidationPI.Rejected(rejectedValidationPI.IncomingInstruction, "wtf??")

		failedPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).Build(),
		).Build()

		failedPI.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Failure,
			PaymentInstruction: failedPI,
			FailureReason: models.FailureReason{
				Code:    models.RejectedCode,
				Message: "blah",
			},
		})

		assert.NoError(t, paymentStore.Store(ctx, pendingPI))
		assert.NoError(t, paymentStore.Store(ctx, successPI1))
		assert.NoError(t, paymentStore.Store(ctx, successPI2))
		assert.NoError(t, paymentStore.Store(ctx, successPIYesterday))
		assert.NoError(t, paymentStore.Store(ctx, rejectedValidationPI))
		assert.NoError(t, paymentStore.Store(ctx, failedPI))
		// end of heinous

		toRound := time.Now()
		date := time.Date(toRound.Year(), toRound.Month(), toRound.Day(), 0, 0, 0, 0, toRound.Location())

		report, err := paymentStore.GetReport(ctx, date)
		assert.NoError(t, err)
		assert.Equal(t, models.PaymentReport{
			Stats: models.PaymentStats{
				Successful:             2,
				Failed:                 1,
				SubmittedForProcessing: 1,
				Rejected:               1,
			},
			FailedPayments: models.FailedPayments{
				FailedStats: models.FailedStats{
					Rejected:         1,
					StuckInPending:   0,
					Unhandled:        0,
					TransportMishap:  0,
					NoSourceAcct:     0,
					FailedValidation: 1,
					MissingFunds:     0,
				},
				FailedInstructions: []models.FailedInstruction{
					{ID: rejectedValidationPI.ID(), Currency: rejectedValidationPI.IncomingInstruction.IsoCode(), Mid: rejectedValidationPI.ContractNumber(), Reason: models.FailedValidation},
					{ID: failedPI.ID(), Currency: failedPI.IncomingInstruction.IsoCode(), Mid: failedPI.ContractNumber(), Reason: models.RejectedPayment},
				},
			},
		}, report)
	})
}

func TestPostgresStore_GetCurrencyReport(t *testing.T) {
	pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	assert.NoError(t, err)
	err = paymentStore.CleanDBForTesting()
	assert.NoError(t, err)
	t.Run("Generate currency report for valid payments", func(t *testing.T) {
		// todo: this is HEINOUS, what do we do? this is howling how wonky the PI API
		// still is, sometimes methods like reject, sometimes events, and sometimes we
		// take a PI, sometimes incoming instruction
		ctx := context.Background()
		pendingPI1Mid := "1234567"
		successPI1Mid := "2345678"
		successPI2Mid := "9435643"
		successPI3Mid := "2468790"
		rejectedValidationPIMid := "3245667"
		failedPIMid := "8705685"

		pendingPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithMerchantAccountNumber(pendingPI1Mid).WithPaymentAmount("50").Build(),
		).Build()
		pendingPI.SubmitForProcessing()

		successPI1 := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithMerchantAccountNumber(successPI1Mid).WithPaymentAmount("50").Build(),
		).Build()
		successPI1.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPI1,
		})

		successPI2 := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithMerchantAccountNumber(successPI2Mid).WithPaymentAmount("50").WithCurrency(models.Currency{
				IsoCode:   "CZK",
				IsoNumber: "203",
			}).Build(),
		).Build()
		successPI2.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPI2,
		})

		successPI3 := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithPaymentAmount("50").WithMerchantAccountNumber(successPI3Mid).WithHighRIsk().Build(),
		).Build()
		successPI3.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Submitted,
			PaymentInstruction: successPI3,
		})

		rejectedValidationPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithMerchantAccountNumber(rejectedValidationPIMid).WithPaymentAmount("50").Build(),
		).Build()
		rejectedValidationPI.Rejected(rejectedValidationPI.IncomingInstruction, "wtf??")

		failedPI := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentAmount("50").WithPaymentExecutionDate(time.Now()).WithMerchantAccountNumber(failedPIMid).WithCurrency(models.Currency{
				IsoCode:   "CZK",
				IsoNumber: "203",
			}).Build(),
		).Build()

		failedPI.TrackPPEvent(models.PaymentProviderEvent{
			Type:               models.Failure,
			PaymentInstruction: failedPI,
			FailureReason: models.FailureReason{
				Code:    models.RejectedCode,
				Message: "blah",
			},
		})

		assert.NoError(t, paymentStore.Store(ctx, pendingPI))
		assert.NoError(t, paymentStore.Store(ctx, successPI1))
		assert.NoError(t, paymentStore.Store(ctx, successPI2))
		assert.NoError(t, paymentStore.Store(ctx, successPI3))
		assert.NoError(t, paymentStore.Store(ctx, rejectedValidationPI))
		assert.NoError(t, paymentStore.Store(ctx, failedPI))
		// end of heinous
		expectedReportMap := map[string]models.CurrencyStats{}
		expectedReportMap["EUR"] = models.CurrencyStats{
			Successful:       1,
			Failures:         2,
			Total:            3,
			SuccessfulAmount: 50,
			FailuresAmount:   100,
			TotalAmount:      150,
		}
		expectedReportMap["CZK"] = models.CurrencyStats{
			Successful:       1,
			Failures:         1,
			Total:            2,
			SuccessfulAmount: 50,
			FailuresAmount:   50,
			TotalAmount:      100,
		}
		expectedReportMap["EUR_HR"] = models.CurrencyStats{
			Successful:       1,
			Failures:         0,
			Total:            1,
			SuccessfulAmount: 50,
			FailuresAmount:   0,
			TotalAmount:      50,
		}
		toRound := time.Now()
		date := time.Date(toRound.Year(), toRound.Month(), toRound.Day(), 0, 0, 0, 0, toRound.Location())

		report, err := paymentStore.GetCurrencyReport(ctx, date)
		assert.NoError(t, err)
		assert.Equal(t, models.PaymentCurrencyReport{CurrencyReport: expectedReportMap}, report)
	})
}

func TestGetPaymentByMid(t *testing.T) {
	pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	assert.NoError(t, err)

	t.Run("endpoint returns relevant information from successful payment", func(t *testing.T) {
		ctx := context.Background()
		// given a valid payment instruction
		paymentInstruction := testhelpers.NewPaymentInstructionBuilder().WithIncomingInstruction(
			testhelpers.NewIncomingInstructionBuilder().WithPaymentExecutionDate(time.Now()).WithMerchantContractNumber("F00").Build(),
		).Build()
		// when i store the payment instruction

		err = paymentStore.Store(ctx, paymentInstruction)
		assert.NoError(t, err)

		paymentByMid, err := paymentStore.GetPaymentByMid(ctx, paymentInstruction.ContractNumber(), time.Now())
		assert.NoError(t, err)
		assert.Equal(t, paymentInstruction.ContractNumber(), paymentByMid.ContractNumber())
	})
}

func TestGetPaymentByCorrelationID(t *testing.T) {
	pgString := os.Getenv("POSTGRES_DB_CONNECTION_STRING")
	if pgString == "" {
		t.Fatal("POSTGRES_DB_CONNECTION_STRING environment variable is not set ")
	}
	paymentStore, err := NewPaymentStore(
		context.Background(),
		pgString,
		payment_store.NewLoggingAndMetricsPaymentObservabilityForPostgres(testdoubles.DummyMetricsClient{}),
	)
	assert.NoError(t, err)

	t.Run("should successfully return payment instruction for a correlationID", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "correlationID"
		paymentInstruction := testhelpers.NewPaymentInstructionBuilder().
			WithIncomingInstruction(
				testhelpers.
					NewIncomingInstructionBuilder().
					WithPaymentExecutionDate(time.Now()).
					WithMerchantContractNumber("F00").
					WithCorrelationID(correlationID).
					Build(),
			).Build()

		err = paymentStore.Store(ctx, paymentInstruction)
		assert.NoError(t, err)

		paymentByCorrelationID, err := paymentStore.GetFromCorrelationID(ctx, correlationID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(paymentByCorrelationID))
		assert.Equal(t, paymentInstruction.IncomingInstruction.PaymentCorrelationId, paymentByCorrelationID[0].IncomingInstruction.PaymentCorrelationId)
	})

	t.Run("should return error when correlationID does not exists in a payment instruction", func(t *testing.T) {
		ctx := context.Background()
		correlationID := "unknown-correlationID"
		expectedError := PaymentInstructionMissingError{CorrelationID: correlationID}

		_, err := paymentStore.GetFromCorrelationID(ctx, correlationID)
		assert.Equal(t, err, expectedError)
	})
}
