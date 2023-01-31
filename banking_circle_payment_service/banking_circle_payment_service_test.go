//go:build unit
// +build unit

package banking_circle_payment_service_test

// todo: separate this file into a file for each use_case because it's getting big and move to use_case

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/pkg/errors"

	bcmodels "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports/mocks"
	. "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/use_cases"

	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

const (
	paymentID        = models.ProviderPaymentID("banking-circle-payment-id")
	bankingReference = models.BankingReference("banking-reference")
)

var (
	dummyMetrics = testdoubles.DummyMetricsClient{}
	now          = time.Now()
	dummyNowFunc = func() time.Time { return now }
)

func TestBankingCircleMakePaymentUseCase_Execute(t *testing.T) {
	sourceAccounts := testSourceAccounts()

	t.Run("Happy Path", func(t *testing.T) {
		is := is.New(t)
		t.Run("call BC make payment with the correct DTO for the given instruction", func(t *testing.T) {
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error) {
				return spe.ResponseDto{
					PaymentID: paymentID,
					Status:    string(ports.PendingProcessing),
				}, nil
			}}

			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error {
				return nil
			}}

			makeBcPaymentUseCase := NewMakeBankingCirclePayment(MakeBankingCirclePaymentOptions{
				PaymentAPI:         mockBankingCircleAPIClient,
				SourceAccounts:     sourceAccounts,
				MetricsClient:      dummyMetrics,
				PaymentNotifier:    mockPaymentNotifier,
				SubmissionNotifier: mockPaymentNotifier,
				Now:                dummyNowFunc,
			})
			t.Run("Currency EUR, and high-risk", func(t *testing.T) {
				ctx := context.Background()
				incomingPaymentInstruction, expectedRequestDto := validPaymentInstructionAndExpectedRequestDto(string(models.EUR), "978", true)

				_, err := makeBcPaymentUseCase.Execute(ctx, incomingPaymentInstruction)

				is.NoErr(err)
				is.Equal(len(mockBankingCircleAPIClient.RequestPaymentCalls()), 1)
				is.Equal(mockBankingCircleAPIClient.RequestPaymentCalls()[0].Request, expectedRequestDto)
			})
			t.Run("EUR, low-risk", func(t *testing.T) {
				ctx := context.Background()
				incomingPaymentInstruction, expectedRequestDto := validPaymentInstructionAndExpectedRequestDto(string(models.EUR), "978", false)

				_, err := makeBcPaymentUseCase.Execute(ctx, incomingPaymentInstruction)

				is.NoErr(err)
				is.Equal(len(mockBankingCircleAPIClient.RequestPaymentCalls()), 2)
				is.Equal(mockBankingCircleAPIClient.RequestPaymentCalls()[1].Request, expectedRequestDto)
			})
			t.Run("Currency CZK, and NO high-risk", func(t *testing.T) {
				ctx := context.Background()
				incomingPaymentInstruction, expectedRequestDto := validPaymentInstructionAndExpectedRequestDto(string(models.CZK), "203", false)

				_, err := makeBcPaymentUseCase.Execute(ctx, incomingPaymentInstruction)

				is.NoErr(err)
				is.Equal(len(mockBankingCircleAPIClient.RequestPaymentCalls()), 3)
				is.Equal(mockBankingCircleAPIClient.RequestPaymentCalls()[2].Request, expectedRequestDto)
			})
		})
		t.Run("notifying payment statuses", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(context.Context, spe.RequestDto, *[]string) (spe.ResponseDto, error) {
				return spe.ResponseDto{
					PaymentID: paymentID,
					Status:    string(ports.PendingProcessing),
				}, nil
			}}

			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}
			mockSubmissionNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			makeBcPaymentUseCase := NewMakeBankingCirclePayment(MakeBankingCirclePaymentOptions{
				PaymentAPI:         mockBankingCircleAPIClient,
				SourceAccounts:     sourceAccounts,
				MetricsClient:      dummyMetrics,
				PaymentNotifier:    mockPaymentNotifier,
				SubmissionNotifier: mockSubmissionNotifier,
				Now:                dummyNowFunc,
			})

			validPaymentReq, _ := validPaymentInstructionAndExpectedRequestDto(string(models.EUR), "978", true)

			_, err := makeBcPaymentUseCase.Execute(ctx, validPaymentReq)
			is.NoErr(err)
			is.Equal(len(mockSubmissionNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockSubmissionNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Submitted,
				PaymentInstruction:       validPaymentReq,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
			})
		})
	})
	t.Run("Unhappy Path", func(t *testing.T) {
		is := is.New(t)

		t.Run("Given calling BC RequestPayment with DTO containing source account that does not exist Then send Failure event with NoSourceAccountError code", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(context.Context, spe.RequestDto, *[]string) (spe.ResponseDto, error) {
				return spe.ResponseDto{}, nil
			}}

			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			makeBcPaymentUseCase := NewMakeBankingCirclePayment(MakeBankingCirclePaymentOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				SourceAccounts:  sourceAccounts,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			badISOCode := "FOO"
			invalidPaymentReq, _ := validPaymentInstructionAndExpectedRequestDto(badISOCode, "978", true)

			_, err := makeBcPaymentUseCase.Execute(ctx, invalidPaymentReq)
			is.True(err != nil)
			_, isSourceAccountError := err.(NoSourceAccountError)
			is.True(isSourceAccountError)

			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       invalidPaymentReq,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: "",
				FailureReason: models.FailureReason{
					Code: models.NoSourceAccount,
					Message: NoSourceAccountError{
						IsoCode:  badISOCode,
						HighRisk: true,
					}.Error(),
				},
			})
		})
		t.Run("Given calling BC RequestPayment results transient error Then send Failure event with TransportFailure code", func(t *testing.T) {
			transportError := errors.New("oh damn")
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{RequestPaymentFunc: func(context.Context, spe.RequestDto, *[]string) (spe.ResponseDto, error) {
				return spe.ResponseDto{}, transportError
			}}

			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error {
				return nil
			}}

			makeBcPaymentUseCase := NewMakeBankingCirclePayment(MakeBankingCirclePaymentOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				SourceAccounts:  sourceAccounts,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			validPaymentReq, _ := validPaymentInstructionAndExpectedRequestDto(string(models.EUR), "978", false)

			_, err := makeBcPaymentUseCase.Execute(ctx, validPaymentReq)
			is.Equal(err, transportError)
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       validPaymentReq,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: "",
				FailureReason: models.FailureReason{
					Code: models.TransportFailure,
					Message: TransportError{
						UnderlyingError: transportError,
						ID:              validPaymentReq.ID(),
						ContractNumber:  validPaymentReq.ContractNumber(),
					}.Error(),
				},
			})
		})
	})
}

func TestBankingCircleCheckPaymentStatusUseCase_Execute(t *testing.T) {
	incomingPaymentInstruction, _ := validPaymentInstructionAndExpectedRequestDto(string(models.EUR), "978", true)
	is := is.New(t)

	t.Run("Happy Path", func(t *testing.T) {
		t.Run("Given BC CheckPaymentStatus is called Then payment status has been checked", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.Processed, nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(mockBankingCircleAPIClient.CheckPaymentStatusCalls()[0].PaymentID, paymentID)
			is.Equal(len(mockBankingCircleAPIClient.CheckPaymentStatusCalls()), 1)
		})
		t.Run("Given BC CheckPaymentStatus returns Success response we should Then send Processed event", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.Processed, nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Processed,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason:            models.FailureReason{},
			})
		})
		t.Run("Given BC CheckPaymentStatus transient error should not break the loop but just try again and Then send Failure event with StuckInPending code", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := stubAPIClientThatWillReturnUnprocessedTwiceThenProcessed()
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Processed,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason:            models.FailureReason{},
			})
		})
	})
	t.Run("Unhappy Path", func(t *testing.T) {
		t.Run("Given BC CheckPaymentStatus returns Rejected status we Then send StuckInPending payment event", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.PendingProcessing, nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			err := checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now)
			is.True(strings.Contains(err.Error(), string(ports.PendingProcessing)))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason: models.FailureReason{
					Code:    models.StuckInPending,
					Message: BankingCircleError{Status: ports.PendingProcessing}.Error(),
				},
			})
		})
		t.Run("Given BC CheckPaymentStatus returns Rejected status Then send RejectedCode payment event", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.Rejected, nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason: models.FailureReason{
					Code:    models.RejectedCode,
					Message: BankingCircleError{Status: ports.Rejected}.Error(),
				},
			})
		})
		t.Run("Given BC CheckPaymentStatus returns MissingFunds status Then send MissingFunds payment event", func(t *testing.T) {
			ctx := context.Background()
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.MissingFunding, nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason: models.FailureReason{
					Code:    models.MissingFunding,
					Message: BankingCircleError{Status: ports.MissingFunding}.Error(),
				},
			})
		})
		t.Run("Given BC CheckPaymentStatus returns an unexpected status we Then send Failure event including UnhandledPaymentProviderStatus code", func(t *testing.T) {
			ctx := context.Background()
			status := "UnexpectedResponse"
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(providerPaymentId models.ProviderPaymentID) (ports.PaymentStatus, error) {
				return ports.PaymentStatus(status), nil
			}}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason: models.FailureReason{
					Code:    models.UnhandledPaymentProviderStatus,
					Message: BankingCircleError{Status: ports.PaymentStatus(status)}.Error(),
				},
			})
		})

		t.Run("Given BC CheckPaymentStatus returns an error we should Then send TransportFailure event", func(t *testing.T) {
			ctx := context.Background()
			transportError := errors.New("oh damn")
			mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{CheckPaymentStatusFunc: func(paymentID models.ProviderPaymentID) (ports.PaymentStatus, error) { return "", transportError }}
			mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

			checkBcPaymentStatusUseCase := NewCheckBankingCirclePaymentStatus(CheckBankingCirclePaymentStatusOptions{
				PaymentAPI:      mockBankingCircleAPIClient,
				MetricsClient:   dummyMetrics,
				PaymentNotifier: mockPaymentNotifier,
				Now:             dummyNowFunc,
			})

			is.NoErr(checkBcPaymentStatusUseCase.Execute(ctx, incomingPaymentInstruction, paymentID, bankingReference, now))
			is.Equal(len(mockPaymentNotifier.SendPaymentStatusCalls()), 1)
			is.Equal(mockPaymentNotifier.SendPaymentStatusCalls()[0].Event, models.PaymentProviderEvent{
				CreatedOn:                now,
				Type:                     models.Failure,
				PaymentInstruction:       incomingPaymentInstruction,
				PaymentProviderName:      models.BC,
				PaymentProviderPaymentID: paymentID,
				BankingReference:         bankingReference,
				FailureReason: models.FailureReason{
					Code: models.TransportFailure,
					Message: TransportError{
						UnderlyingError: transportError,
						ID:              incomingPaymentInstruction.ID(),
						ContractNumber:  incomingPaymentInstruction.ContractNumber(),
					}.Error(),
				},
			})
		})
	})
}

func TestBankingCircleGetRejectionReportUseCase_Execute(t *testing.T) {
	is := is.New(t)
	executionDate := "2021-10-06"
	expectedReport := bcmodels.RejectionReport{}
	t.Run("Given a valid date, BC should return a rejection report", func(t *testing.T) {
		mockBankingCircleAPIClient := &mocks.BankingCircleAPIMock{GetRejectionReportFunc: func(date string) (bcmodels.RejectionReport, error) {
			return expectedReport, nil
		}}
		mockPaymentNotifier := &mocks.PaymentNotifierMock{SendPaymentStatusFunc: func(context.Context, models.PaymentProviderEvent) error { return nil }}

		checkBcGetRejectionReportUseCase := NewGetBankingCircleRejectionReport(GetBankingCircleRejectionReportOptions{
			PaymentAPI:      mockBankingCircleAPIClient,
			MetricsClient:   dummyMetrics,
			PaymentNotifier: mockPaymentNotifier,
			Now:             dummyNowFunc,
		})

		report, err := checkBcGetRejectionReportUseCase.Execute(executionDate)
		is.NoErr(err)
		is.Equal(report, expectedReport)
	})
}

func testSourceAccounts() bcmodels.SourceAccounts {
	return []bcmodels.SourceAccount{
		{
			Currency:   "EUR",
			IsHighRisk: true,
			AccountDetails: bcmodels.AccountDetails{
				Iban:      "IBAN_EUR_HR",
				AccountID: "64879db2-a8ba-34ee-c0d2381ebe4a",
			},
		},
		{
			Currency:   "EUR",
			IsHighRisk: false,
			AccountDetails: bcmodels.AccountDetails{
				Iban:      "IBAN_EUR",
				AccountID: "0e89a45e-9c7a-34ee-0b85d2abaf4e",
			},
		},
		{
			Currency:   "CZK",
			IsHighRisk: false,
			AccountDetails: bcmodels.AccountDetails{
				Iban:      "IBAN_CZK",
				AccountID: "18752399-01ff-cb5a-02a1eeb02d01",
			},
		},
	}
}

func stubAPIClientThatWillReturnUnprocessedTwiceThenProcessed() ports.BankingCircleAPI {
	stub := &stubAPIClient{0}

	return stub
}

type stubAPIClient struct {
	checkPaymentStatusCallCount int
}

func (s *stubAPIClient) RequestPayment(context.Context, spe.RequestDto, *[]string) (spe.ResponseDto, error) {
	return spe.ResponseDto{
		PaymentID: "banking-circle-payment-id",
		Status:    string(ports.PendingProcessing),
	}, nil
}

func (s *stubAPIClient) CheckPaymentStatus(paymentRequestID models.ProviderPaymentID) (ports.PaymentStatus, error) {
	fmt.Println("--- CheckPaymentStatus called", paymentRequestID, s.checkPaymentStatusCallCount)
	if paymentRequestID != "banking-circle-payment-id" {
		return "", errors.New("expected paymentRequestId as \"banking-circle-payment-id\"")
	}
	s.checkPaymentStatusCallCount += 1
	if s.checkPaymentStatusCallCount <= 3 {
		return ports.PendingProcessing, nil
	}

	return ports.Processed, nil
}

func (s *stubAPIClient) GetRejectionReport(string) (bcmodels.RejectionReport, error) {
	return bcmodels.RejectionReport{Rejections: []bcmodels.Rejection{}}, nil
}

func (s *stubAPIClient) CheckAccountBalance(string) (bcmodels.AccountBalance, error) {
	return bcmodels.AccountBalance{}, nil
}

func validPaymentInstructionAndExpectedRequestDto(currency string, currencyIsoNumber string, isHighRisk bool) (models.PaymentInstruction, spe.RequestDto) {
	fileName := ""

	if isHighRisk {
		fileName = "a-ufx-_HR_file-name.xml"
	} else {
		fileName = "a-ufx-file-name.xml"
	}

	instruction := models.IncomingInstruction{
		Merchant: models.Merchant{
			ContractNumber: "9876862",
			Name:           "A merchant",
			Email:          "merchant@merchant.com",
			Address: models.Address{
				Country:      "EUR",
				City:         "TestCity",
				AddressLine1: "Some street name",
				AddressLine2: "",
			},
			Account: models.Account{
				AccountNumber:        "IS123456789012345678901234",
				Swift:                "",
				Country:              "",
				SwiftReferenceNumber: "",
			},
			HighRisk: isHighRisk,
		},
		Metadata: models.Metadata{
			Source:   "Way4",
			Filename: "UFX",
			FileType: fileName,
		},
		Payment: models.Payment{
			Sender: models.Sender{
				Name:          "SAXO",
				AccountNumber: "the-sender-iban",
				BranchCode:    "",
			},
			Amount: "1234.56",
			Currency: models.Currency{
				IsoCode:   models.CurrencyCode(currency),
				IsoNumber: currencyIsoNumber,
			},
			ExecutionDate: time.Date(2021, time.May, 25, 0, 0, 0, 0, time.UTC),
		},
	}
	paymentReq := models.NewPaymentInstruction(instruction)

	accountNumber, _ := testSourceAccounts().FindAccountNumber(currency, isHighRisk)

	expectedRequestDto := spe.RequestDto{
		RequestedExecutiondate: time.Date(2021, time.May, 25, 0, 0, 0, 0, time.UTC),
		DebtorAccount: spe.DebtorAccount{
			Account:              accountNumber.Iban,
			FinancialInstitution: "",
			Country:              "",
		},
		DebtorViban:           "",
		DebtorReference:       "Settlm 9876862 20210525",
		DebtorNarrativeToSelf: "9876862",
		CurrencyOfTransfer:    currency,
		Amount: spe.Amount{
			Currency: currency,
			Amount:   1234.56,
		},
		ChargeBearer: "SHA",
		RemittanceInformation: spe.RemittanceInformation{
			Line1: "9876862",
			Line2: "Paydate 20210525",
			Line3: "Settlm",
			Line4: "",
		},
		CreditorID: "",
		CreditorAccount: spe.CreditorAccount{
			Account:              "IS123456789012345678901234",
			FinancialInstitution: "",
			Country:              "",
		},
		CreditorName: "A merchant",
		CreditorAddress: spe.CreditorAddress{
			Line1: "Some street name",
			Line2: "EUR TestCity",
			Line3: "",
		},
	}

	return paymentReq, expectedRequestDto
}
