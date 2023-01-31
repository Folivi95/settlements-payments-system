// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"context"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_file_listener"
	"io"
	"sync"
)

// Ensure, that FileStoreMock does implement ufx_file_listener.FileStore.
// If this is not the case, regenerate this file with moq.
var _ ufx_file_listener.FileStore = &FileStoreMock{}

// FileStoreMock is a mock implementation of ufx_file_listener.FileStore.
//
//	func TestSomethingThatUsesFileStore(t *testing.T) {
//
//		// make and configure a mocked ufx_file_listener.FileStore
//		mockedFileStore := &FileStoreMock{
//			ReadFileFunc: func(ctx context.Context, filename string) (io.Reader, error) {
//				panic("mock out the ReadFile method")
//			},
//		}
//
//		// use mockedFileStore in code that requires ufx_file_listener.FileStore
//		// and then make assertions.
//
//	}
type FileStoreMock struct {
	// ReadFileFunc mocks the ReadFile method.
	ReadFileFunc func(ctx context.Context, filename string) (io.Reader, error)

	// calls tracks calls to the methods.
	calls struct {
		// ReadFile holds details about calls to the ReadFile method.
		ReadFile []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Filename is the filename argument value.
			Filename string
		}
	}
	lockReadFile sync.RWMutex
}

// ReadFile calls ReadFileFunc.
func (mock *FileStoreMock) ReadFile(ctx context.Context, filename string) (io.Reader, error) {
	if mock.ReadFileFunc == nil {
		panic("FileStoreMock.ReadFileFunc: method is nil but FileStore.ReadFile was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		Filename string
	}{
		Ctx:      ctx,
		Filename: filename,
	}
	mock.lockReadFile.Lock()
	mock.calls.ReadFile = append(mock.calls.ReadFile, callInfo)
	mock.lockReadFile.Unlock()
	return mock.ReadFileFunc(ctx, filename)
}

// ReadFileCalls gets all the calls that were made to ReadFile.
// Check the length with:
//
//	len(mockedFileStore.ReadFileCalls())
func (mock *FileStoreMock) ReadFileCalls() []struct {
	Ctx      context.Context
	Filename string
} {
	var calls []struct {
		Ctx      context.Context
		Filename string
	}
	mock.lockReadFile.RLock()
	calls = mock.calls.ReadFile
	mock.lockReadFile.RUnlock()
	return calls
}