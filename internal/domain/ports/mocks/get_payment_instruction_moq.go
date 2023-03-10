// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
	"sync"
)

// Ensure, that GetPaymentInstructionMock does implement ports.GetPaymentInstruction.
// If this is not the case, regenerate this file with moq.
var _ ports.GetPaymentInstruction = &GetPaymentInstructionMock{}

// GetPaymentInstructionMock is a mock implementation of ports.GetPaymentInstruction.
//
// 	func TestSomethingThatUsesGetPaymentInstruction(t *testing.T) {
//
// 		// make and configure a mocked ports.GetPaymentInstruction
// 		mockedGetPaymentInstruction := &GetPaymentInstructionMock{
// 			ExecuteFunc: func(ctx context.Context, paymentInstructionID models.PaymentInstructionID) (models.PaymentInstruction, error) {
// 				panic("mock out the Execute method")
// 			},
// 			RetrieveByCorrelationIDFunc: func(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error) {
// 				panic("mock out the RetrieveByCorrelationID method")
// 			},
// 		}
//
// 		// use mockedGetPaymentInstruction in code that requires ports.GetPaymentInstruction
// 		// and then make assertions.
//
// 	}
type GetPaymentInstructionMock struct {
	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(ctx context.Context, paymentInstructionID models.PaymentInstructionID) (models.PaymentInstruction, error)

	// RetrieveByCorrelationIDFunc mocks the RetrieveByCorrelationID method.
	RetrieveByCorrelationIDFunc func(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error)

	// calls tracks calls to the methods.
	calls struct {
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// PaymentInstructionID is the paymentInstructionID argument value.
			PaymentInstructionID models.PaymentInstructionID
		}
		// RetrieveByCorrelationID holds details about calls to the RetrieveByCorrelationID method.
		RetrieveByCorrelationID []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// CorrelationID is the correlationID argument value.
			CorrelationID string
		}
	}
	lockExecute                 sync.RWMutex
	lockRetrieveByCorrelationID sync.RWMutex
}

// Execute calls ExecuteFunc.
func (mock *GetPaymentInstructionMock) Execute(ctx context.Context, paymentInstructionID models.PaymentInstructionID) (models.PaymentInstruction, error) {
	if mock.ExecuteFunc == nil {
		panic("GetPaymentInstructionMock.ExecuteFunc: method is nil but GetPaymentInstruction.Execute was just called")
	}
	callInfo := struct {
		Ctx                  context.Context
		PaymentInstructionID models.PaymentInstructionID
	}{
		Ctx:                  ctx,
		PaymentInstructionID: paymentInstructionID,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	return mock.ExecuteFunc(ctx, paymentInstructionID)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//     len(mockedGetPaymentInstruction.ExecuteCalls())
func (mock *GetPaymentInstructionMock) ExecuteCalls() []struct {
	Ctx                  context.Context
	PaymentInstructionID models.PaymentInstructionID
} {
	var calls []struct {
		Ctx                  context.Context
		PaymentInstructionID models.PaymentInstructionID
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}

// RetrieveByCorrelationID calls RetrieveByCorrelationIDFunc.
func (mock *GetPaymentInstructionMock) RetrieveByCorrelationID(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error) {
	if mock.RetrieveByCorrelationIDFunc == nil {
		panic("GetPaymentInstructionMock.RetrieveByCorrelationIDFunc: method is nil but GetPaymentInstruction.RetrieveByCorrelationID was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		CorrelationID string
	}{
		Ctx:           ctx,
		CorrelationID: correlationID,
	}
	mock.lockRetrieveByCorrelationID.Lock()
	mock.calls.RetrieveByCorrelationID = append(mock.calls.RetrieveByCorrelationID, callInfo)
	mock.lockRetrieveByCorrelationID.Unlock()
	return mock.RetrieveByCorrelationIDFunc(ctx, correlationID)
}

// RetrieveByCorrelationIDCalls gets all the calls that were made to RetrieveByCorrelationID.
// Check the length with:
//     len(mockedGetPaymentInstruction.RetrieveByCorrelationIDCalls())
func (mock *GetPaymentInstructionMock) RetrieveByCorrelationIDCalls() []struct {
	Ctx           context.Context
	CorrelationID string
} {
	var calls []struct {
		Ctx           context.Context
		CorrelationID string
	}
	mock.lockRetrieveByCorrelationID.RLock()
	calls = mock.calls.RetrieveByCorrelationID
	mock.lockRetrieveByCorrelationID.RUnlock()
	return calls
}
