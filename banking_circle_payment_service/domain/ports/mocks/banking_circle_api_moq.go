// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	ppEvents "github.com/saltpay/settlements-payments-system/internal/domain/models"
	"sync"
)

// Ensure, that BankingCircleAPIMock does implement ports.BankingCircleAPI.
// If this is not the case, regenerate this file with moq.
var _ ports.BankingCircleAPI = &BankingCircleAPIMock{}

// BankingCircleAPIMock is a mock implementation of ports.BankingCircleAPI.
//
// 	func TestSomethingThatUsesBankingCircleAPI(t *testing.T) {
//
// 		// make and configure a mocked ports.BankingCircleAPI
// 		mockedBankingCircleAPI := &BankingCircleAPIMock{
// 			CheckAccountBalanceFunc: func(accountID string) (models2.AccountBalance, error) {
// 				panic("mock out the CheckAccountBalance method")
// 			},
// 			CheckPaymentStatusFunc: func(paymentID ppEvents.ProviderPaymentID) (ports.PaymentStatus, error) {
// 				panic("mock out the CheckPaymentStatus method")
// 			},
// 			GetRejectionReportFunc: func(date string) (models2.RejectionReport, error) {
// 				panic("mock out the GetRejectionReport method")
// 			},
// 			RequestPaymentFunc: func(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error) {
// 				panic("mock out the RequestPayment method")
// 			},
// 		}
//
// 		// use mockedBankingCircleAPI in code that requires ports.BankingCircleAPI
// 		// and then make assertions.
//
// 	}
type BankingCircleAPIMock struct {
	// CheckAccountBalanceFunc mocks the CheckAccountBalance method.
	CheckAccountBalanceFunc func(accountID string) (models2.AccountBalance, error)

	// CheckPaymentStatusFunc mocks the CheckPaymentStatus method.
	CheckPaymentStatusFunc func(paymentID ppEvents.ProviderPaymentID) (ports.PaymentStatus, error)

	// GetRejectionReportFunc mocks the GetRejectionReport method.
	GetRejectionReportFunc func(date string) (models2.RejectionReport, error)

	// RequestPaymentFunc mocks the RequestPayment method.
	RequestPaymentFunc func(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error)

	// calls tracks calls to the methods.
	calls struct {
		// CheckAccountBalance holds details about calls to the CheckAccountBalance method.
		CheckAccountBalance []struct {
			// AccountID is the accountID argument value.
			AccountID string
		}
		// CheckPaymentStatus holds details about calls to the CheckPaymentStatus method.
		CheckPaymentStatus []struct {
			// PaymentID is the paymentID argument value.
			PaymentID ppEvents.ProviderPaymentID
		}
		// GetRejectionReport holds details about calls to the GetRejectionReport method.
		GetRejectionReport []struct {
			// Date is the date argument value.
			Date string
		}
		// RequestPayment holds details about calls to the RequestPayment method.
		RequestPayment []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Request is the request argument value.
			Request spe.RequestDto
			// Slice is the slice argument value.
			Slice *[]string
		}
	}
	lockCheckAccountBalance sync.RWMutex
	lockCheckPaymentStatus  sync.RWMutex
	lockGetRejectionReport  sync.RWMutex
	lockRequestPayment      sync.RWMutex
}

// CheckAccountBalance calls CheckAccountBalanceFunc.
func (mock *BankingCircleAPIMock) CheckAccountBalance(accountID string) (models2.AccountBalance, error) {
	if mock.CheckAccountBalanceFunc == nil {
		panic("BankingCircleAPIMock.CheckAccountBalanceFunc: method is nil but BankingCircleAPI.CheckAccountBalance was just called")
	}
	callInfo := struct {
		AccountID string
	}{
		AccountID: accountID,
	}
	mock.lockCheckAccountBalance.Lock()
	mock.calls.CheckAccountBalance = append(mock.calls.CheckAccountBalance, callInfo)
	mock.lockCheckAccountBalance.Unlock()
	return mock.CheckAccountBalanceFunc(accountID)
}

// CheckAccountBalanceCalls gets all the calls that were made to CheckAccountBalance.
// Check the length with:
//     len(mockedBankingCircleAPI.CheckAccountBalanceCalls())
func (mock *BankingCircleAPIMock) CheckAccountBalanceCalls() []struct {
	AccountID string
} {
	var calls []struct {
		AccountID string
	}
	mock.lockCheckAccountBalance.RLock()
	calls = mock.calls.CheckAccountBalance
	mock.lockCheckAccountBalance.RUnlock()
	return calls
}

// CheckPaymentStatus calls CheckPaymentStatusFunc.
func (mock *BankingCircleAPIMock) CheckPaymentStatus(paymentID ppEvents.ProviderPaymentID) (ports.PaymentStatus, error) {
	if mock.CheckPaymentStatusFunc == nil {
		panic("BankingCircleAPIMock.CheckPaymentStatusFunc: method is nil but BankingCircleAPI.CheckPaymentStatus was just called")
	}
	callInfo := struct {
		PaymentID ppEvents.ProviderPaymentID
	}{
		PaymentID: paymentID,
	}
	mock.lockCheckPaymentStatus.Lock()
	mock.calls.CheckPaymentStatus = append(mock.calls.CheckPaymentStatus, callInfo)
	mock.lockCheckPaymentStatus.Unlock()
	return mock.CheckPaymentStatusFunc(paymentID)
}

// CheckPaymentStatusCalls gets all the calls that were made to CheckPaymentStatus.
// Check the length with:
//     len(mockedBankingCircleAPI.CheckPaymentStatusCalls())
func (mock *BankingCircleAPIMock) CheckPaymentStatusCalls() []struct {
	PaymentID ppEvents.ProviderPaymentID
} {
	var calls []struct {
		PaymentID ppEvents.ProviderPaymentID
	}
	mock.lockCheckPaymentStatus.RLock()
	calls = mock.calls.CheckPaymentStatus
	mock.lockCheckPaymentStatus.RUnlock()
	return calls
}

// GetRejectionReport calls GetRejectionReportFunc.
func (mock *BankingCircleAPIMock) GetRejectionReport(date string) (models2.RejectionReport, error) {
	if mock.GetRejectionReportFunc == nil {
		panic("BankingCircleAPIMock.GetRejectionReportFunc: method is nil but BankingCircleAPI.GetRejectionReport was just called")
	}
	callInfo := struct {
		Date string
	}{
		Date: date,
	}
	mock.lockGetRejectionReport.Lock()
	mock.calls.GetRejectionReport = append(mock.calls.GetRejectionReport, callInfo)
	mock.lockGetRejectionReport.Unlock()
	return mock.GetRejectionReportFunc(date)
}

// GetRejectionReportCalls gets all the calls that were made to GetRejectionReport.
// Check the length with:
//     len(mockedBankingCircleAPI.GetRejectionReportCalls())
func (mock *BankingCircleAPIMock) GetRejectionReportCalls() []struct {
	Date string
} {
	var calls []struct {
		Date string
	}
	mock.lockGetRejectionReport.RLock()
	calls = mock.calls.GetRejectionReport
	mock.lockGetRejectionReport.RUnlock()
	return calls
}

// RequestPayment calls RequestPaymentFunc.
func (mock *BankingCircleAPIMock) RequestPayment(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error) {
	if mock.RequestPaymentFunc == nil {
		panic("BankingCircleAPIMock.RequestPaymentFunc: method is nil but BankingCircleAPI.RequestPayment was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Request spe.RequestDto
		Slice   *[]string
	}{
		Ctx:     ctx,
		Request: request,
		Slice:   slice,
	}
	mock.lockRequestPayment.Lock()
	mock.calls.RequestPayment = append(mock.calls.RequestPayment, callInfo)
	mock.lockRequestPayment.Unlock()
	return mock.RequestPaymentFunc(ctx, request, slice)
}

// RequestPaymentCalls gets all the calls that were made to RequestPayment.
// Check the length with:
//     len(mockedBankingCircleAPI.RequestPaymentCalls())
func (mock *BankingCircleAPIMock) RequestPaymentCalls() []struct {
	Ctx     context.Context
	Request spe.RequestDto
	Slice   *[]string
} {
	var calls []struct {
		Ctx     context.Context
		Request spe.RequestDto
		Slice   *[]string
	}
	mock.lockRequestPayment.RLock()
	calls = mock.calls.RequestPayment
	mock.lockRequestPayment.RUnlock()
	return calls
}
