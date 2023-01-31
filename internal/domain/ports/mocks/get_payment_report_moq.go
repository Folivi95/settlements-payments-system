// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
	"sync"
	"time"
)

// Ensure, that GetPaymentReportMock does implement ports.GetPaymentReport.
// If this is not the case, regenerate this file with moq.
var _ ports.GetPaymentReport = &GetPaymentReportMock{}

// GetPaymentReportMock is a mock implementation of ports.GetPaymentReport.
//
// 	func TestSomethingThatUsesGetPaymentReport(t *testing.T) {
//
// 		// make and configure a mocked ports.GetPaymentReport
// 		mockedGetPaymentReport := &GetPaymentReportMock{
// 			GetCurrencyReportFunc: func(ctx context.Context, timeMoqParam time.Time) (models.PaymentCurrencyReport, error) {
// 				panic("mock out the GetCurrencyReport method")
// 			},
// 			GetPaymentByMidFunc: func(ctx context.Context, mid string, timeMoqParam time.Time) (models.PaymentInstruction, error) {
// 				panic("mock out the GetPaymentByMid method")
// 			},
// 			GetReportFunc: func(ctx context.Context, timeMoqParam time.Time) (models.PaymentReport, error) {
// 				panic("mock out the GetReport method")
// 			},
// 		}
//
// 		// use mockedGetPaymentReport in code that requires ports.GetPaymentReport
// 		// and then make assertions.
//
// 	}
type GetPaymentReportMock struct {
	// GetCurrencyReportFunc mocks the GetCurrencyReport method.
	GetCurrencyReportFunc func(ctx context.Context, timeMoqParam time.Time) (models.PaymentCurrencyReport, error)

	// GetPaymentByMidFunc mocks the GetPaymentByMid method.
	GetPaymentByMidFunc func(ctx context.Context, mid string, timeMoqParam time.Time) (models.PaymentInstruction, error)

	// GetReportFunc mocks the GetReport method.
	GetReportFunc func(ctx context.Context, timeMoqParam time.Time) (models.PaymentReport, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetCurrencyReport holds details about calls to the GetCurrencyReport method.
		GetCurrencyReport []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// TimeMoqParam is the timeMoqParam argument value.
			TimeMoqParam time.Time
		}
		// GetPaymentByMid holds details about calls to the GetPaymentByMid method.
		GetPaymentByMid []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Mid is the mid argument value.
			Mid string
			// TimeMoqParam is the timeMoqParam argument value.
			TimeMoqParam time.Time
		}
		// GetReport holds details about calls to the GetReport method.
		GetReport []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// TimeMoqParam is the timeMoqParam argument value.
			TimeMoqParam time.Time
		}
	}
	lockGetCurrencyReport sync.RWMutex
	lockGetPaymentByMid   sync.RWMutex
	lockGetReport         sync.RWMutex
}

// GetCurrencyReport calls GetCurrencyReportFunc.
func (mock *GetPaymentReportMock) GetCurrencyReport(ctx context.Context, timeMoqParam time.Time) (models.PaymentCurrencyReport, error) {
	if mock.GetCurrencyReportFunc == nil {
		panic("GetPaymentReportMock.GetCurrencyReportFunc: method is nil but GetPaymentReport.GetCurrencyReport was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		TimeMoqParam time.Time
	}{
		Ctx:          ctx,
		TimeMoqParam: timeMoqParam,
	}
	mock.lockGetCurrencyReport.Lock()
	mock.calls.GetCurrencyReport = append(mock.calls.GetCurrencyReport, callInfo)
	mock.lockGetCurrencyReport.Unlock()
	return mock.GetCurrencyReportFunc(ctx, timeMoqParam)
}

// GetCurrencyReportCalls gets all the calls that were made to GetCurrencyReport.
// Check the length with:
//     len(mockedGetPaymentReport.GetCurrencyReportCalls())
func (mock *GetPaymentReportMock) GetCurrencyReportCalls() []struct {
	Ctx          context.Context
	TimeMoqParam time.Time
} {
	var calls []struct {
		Ctx          context.Context
		TimeMoqParam time.Time
	}
	mock.lockGetCurrencyReport.RLock()
	calls = mock.calls.GetCurrencyReport
	mock.lockGetCurrencyReport.RUnlock()
	return calls
}

// GetPaymentByMid calls GetPaymentByMidFunc.
func (mock *GetPaymentReportMock) GetPaymentByMid(ctx context.Context, mid string, timeMoqParam time.Time) (models.PaymentInstruction, error) {
	if mock.GetPaymentByMidFunc == nil {
		panic("GetPaymentReportMock.GetPaymentByMidFunc: method is nil but GetPaymentReport.GetPaymentByMid was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		Mid          string
		TimeMoqParam time.Time
	}{
		Ctx:          ctx,
		Mid:          mid,
		TimeMoqParam: timeMoqParam,
	}
	mock.lockGetPaymentByMid.Lock()
	mock.calls.GetPaymentByMid = append(mock.calls.GetPaymentByMid, callInfo)
	mock.lockGetPaymentByMid.Unlock()
	return mock.GetPaymentByMidFunc(ctx, mid, timeMoqParam)
}

// GetPaymentByMidCalls gets all the calls that were made to GetPaymentByMid.
// Check the length with:
//     len(mockedGetPaymentReport.GetPaymentByMidCalls())
func (mock *GetPaymentReportMock) GetPaymentByMidCalls() []struct {
	Ctx          context.Context
	Mid          string
	TimeMoqParam time.Time
} {
	var calls []struct {
		Ctx          context.Context
		Mid          string
		TimeMoqParam time.Time
	}
	mock.lockGetPaymentByMid.RLock()
	calls = mock.calls.GetPaymentByMid
	mock.lockGetPaymentByMid.RUnlock()
	return calls
}

// GetReport calls GetReportFunc.
func (mock *GetPaymentReportMock) GetReport(ctx context.Context, timeMoqParam time.Time) (models.PaymentReport, error) {
	if mock.GetReportFunc == nil {
		panic("GetPaymentReportMock.GetReportFunc: method is nil but GetPaymentReport.GetReport was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		TimeMoqParam time.Time
	}{
		Ctx:          ctx,
		TimeMoqParam: timeMoqParam,
	}
	mock.lockGetReport.Lock()
	mock.calls.GetReport = append(mock.calls.GetReport, callInfo)
	mock.lockGetReport.Unlock()
	return mock.GetReportFunc(ctx, timeMoqParam)
}

// GetReportCalls gets all the calls that were made to GetReport.
// Check the length with:
//     len(mockedGetPaymentReport.GetReportCalls())
func (mock *GetPaymentReportMock) GetReportCalls() []struct {
	Ctx          context.Context
	TimeMoqParam time.Time
} {
	var calls []struct {
		Ctx          context.Context
		TimeMoqParam time.Time
	}
	mock.lockGetReport.RLock()
	calls = mock.calls.GetReport
	mock.lockGetReport.RUnlock()
	return calls
}
