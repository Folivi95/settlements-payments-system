// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
	"sync"
)

// Ensure, that ReplayPaymentMock does implement ports.ReplayPayment.
// If this is not the case, regenerate this file with moq.
var _ ports.ReplayPayment = &ReplayPaymentMock{}

// ReplayPaymentMock is a mock implementation of ports.ReplayPayment.
//
// 	func TestSomethingThatUsesReplayPayment(t *testing.T) {
//
// 		// make and configure a mocked ports.ReplayPayment
// 		mockedReplayPayment := &ReplayPaymentMock{
// 			ExecuteFunc: func(ctx context.Context, currency string, file string) error {
// 				panic("mock out the Execute method")
// 			},
// 		}
//
// 		// use mockedReplayPayment in code that requires ports.ReplayPayment
// 		// and then make assertions.
//
// 	}
type ReplayPaymentMock struct {
	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(ctx context.Context, currency string, file string) error

	// calls tracks calls to the methods.
	calls struct {
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Currency is the currency argument value.
			Currency string
			// File is the file argument value.
			File string
		}
	}
	lockExecute sync.RWMutex
}

// Execute calls ExecuteFunc.
func (mock *ReplayPaymentMock) Execute(ctx context.Context, currency string, file string) error {
	if mock.ExecuteFunc == nil {
		panic("ReplayPaymentMock.ExecuteFunc: method is nil but ReplayPayment.Execute was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Currency string
		File     string
	}{
		Ctx:      ctx,
		Currency: currency,
		File:     file,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	return mock.ExecuteFunc(ctx, currency, file)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//     len(mockedReplayPayment.ExecuteCalls())
func (mock *ReplayPaymentMock) ExecuteCalls() []struct {
	Ctx      context.Context
	Currency string
	File     string
} {
	var calls []struct {
		Ctx      context.Context
		Currency string
		File     string
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}
