// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	ppEvents "github.com/saltpay/settlements-payments-system/internal/domain/models"
	"sync"
)

// Ensure, that MakeBankingCirclePaymentMock does implement ports.MakeBankingCirclePayment.
// If this is not the case, regenerate this file with moq.
var _ ports.MakeBankingCirclePayment = &MakeBankingCirclePaymentMock{}

// MakeBankingCirclePaymentMock is a mock implementation of ports.MakeBankingCirclePayment.
//
// 	func TestSomethingThatUsesMakeBankingCirclePayment(t *testing.T) {
//
// 		// make and configure a mocked ports.MakeBankingCirclePayment
// 		mockedMakeBankingCirclePayment := &MakeBankingCirclePaymentMock{
// 			ExecuteFunc: func(ctx context.Context, request ppEvents.PaymentInstruction) (ppEvents.ProviderPaymentID, error) {
// 				panic("mock out the Execute method")
// 			},
// 		}
//
// 		// use mockedMakeBankingCirclePayment in code that requires ports.MakeBankingCirclePayment
// 		// and then make assertions.
//
// 	}
type MakeBankingCirclePaymentMock struct {
	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(ctx context.Context, request ppEvents.PaymentInstruction) (ppEvents.ProviderPaymentID, error)

	// calls tracks calls to the methods.
	calls struct {
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Request is the request argument value.
			Request ppEvents.PaymentInstruction
		}
	}
	lockExecute sync.RWMutex
}

// Execute calls ExecuteFunc.
func (mock *MakeBankingCirclePaymentMock) Execute(ctx context.Context, request ppEvents.PaymentInstruction) (ppEvents.ProviderPaymentID, error) {
	if mock.ExecuteFunc == nil {
		panic("MakeBankingCirclePaymentMock.ExecuteFunc: method is nil but MakeBankingCirclePayment.Execute was just called")
	}
	callInfo := struct {
		Ctx     context.Context
		Request ppEvents.PaymentInstruction
	}{
		Ctx:     ctx,
		Request: request,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	return mock.ExecuteFunc(ctx, request)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//     len(mockedMakeBankingCirclePayment.ExecuteCalls())
func (mock *MakeBankingCirclePaymentMock) ExecuteCalls() []struct {
	Ctx     context.Context
	Request ppEvents.PaymentInstruction
} {
	var calls []struct {
		Ctx     context.Context
		Request ppEvents.PaymentInstruction
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}
