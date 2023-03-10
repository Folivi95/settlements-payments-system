// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"github.com/saltpay/settlements-payments-system/internal/adapters/replay_payment"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"io"
	"sync"
)

// Ensure, that UfxConverterMock does implement replay_payment.UfxConverter.
// If this is not the case, regenerate this file with moq.
var _ replay_payment.UfxConverter = &UfxConverterMock{}

// UfxConverterMock is a mock implementation of replay_payment.UfxConverter.
//
// 	func TestSomethingThatUsesUfxConverter(t *testing.T) {
//
// 		// make and configure a mocked replay_payment.UfxConverter
// 		mockedUfxConverter := &UfxConverterMock{
// 			FilterCurrencyFunc: func(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error) {
// 				panic("mock out the FilterCurrency method")
// 			},
// 		}
//
// 		// use mockedUfxConverter in code that requires replay_payment.UfxConverter
// 		// and then make assertions.
//
// 	}
type UfxConverterMock struct {
	// FilterCurrencyFunc mocks the FilterCurrency method.
	FilterCurrencyFunc func(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error)

	// calls tracks calls to the methods.
	calls struct {
		// FilterCurrency holds details about calls to the FilterCurrency method.
		FilterCurrency []struct {
			// UfxFileContents is the ufxFileContents argument value.
			UfxFileContents io.Reader
			// Currency is the currency argument value.
			Currency models.CurrencyCode
		}
	}
	lockFilterCurrency sync.RWMutex
}

// FilterCurrency calls FilterCurrencyFunc.
func (mock *UfxConverterMock) FilterCurrency(ufxFileContents io.Reader, currency models.CurrencyCode) ([]byte, error) {
	if mock.FilterCurrencyFunc == nil {
		panic("UfxConverterMock.FilterCurrencyFunc: method is nil but UfxConverter.FilterCurrency was just called")
	}
	callInfo := struct {
		UfxFileContents io.Reader
		Currency        models.CurrencyCode
	}{
		UfxFileContents: ufxFileContents,
		Currency:        currency,
	}
	mock.lockFilterCurrency.Lock()
	mock.calls.FilterCurrency = append(mock.calls.FilterCurrency, callInfo)
	mock.lockFilterCurrency.Unlock()
	return mock.FilterCurrencyFunc(ufxFileContents, currency)
}

// FilterCurrencyCalls gets all the calls that were made to FilterCurrency.
// Check the length with:
//     len(mockedUfxConverter.FilterCurrencyCalls())
func (mock *UfxConverterMock) FilterCurrencyCalls() []struct {
	UfxFileContents io.Reader
	Currency        models.CurrencyCode
} {
	var calls []struct {
		UfxFileContents io.Reader
		Currency        models.CurrencyCode
	}
	mock.lockFilterCurrency.RLock()
	calls = mock.calls.FilterCurrency
	mock.lockFilterCurrency.RUnlock()
	return calls
}
