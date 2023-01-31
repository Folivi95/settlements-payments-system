// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mocks

import (
	"bytes"
	"context"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws"
	"sync"
)

// Ensure, that S3ClientMock does implement aws.S3Client.
// If this is not the case, regenerate this file with moq.
var _ aws.S3Client = &S3ClientMock{}

// S3ClientMock is a mock implementation of aws.S3Client.
//
// 	func TestSomethingThatUsesS3Client(t *testing.T) {
//
// 		// make and configure a mocked aws.S3Client
// 		mockedS3Client := &S3ClientMock{
// 			PutBucketFileFunc: func(ctx context.Context, file *bytes.Reader, filename string) error {
// 				panic("mock out the PutBucketFile method")
// 			},
// 		}
//
// 		// use mockedS3Client in code that requires aws.S3Client
// 		// and then make assertions.
//
// 	}
type S3ClientMock struct {
	// PutBucketFileFunc mocks the PutBucketFile method.
	PutBucketFileFunc func(ctx context.Context, file *bytes.Reader, filename string) error

	// calls tracks calls to the methods.
	calls struct {
		// PutBucketFile holds details about calls to the PutBucketFile method.
		PutBucketFile []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// File is the file argument value.
			File *bytes.Reader
			// Filename is the filename argument value.
			Filename string
		}
	}
	lockPutBucketFile sync.RWMutex
}

// PutBucketFile calls PutBucketFileFunc.
func (mock *S3ClientMock) PutBucketFile(ctx context.Context, file *bytes.Reader, filename string) error {
	if mock.PutBucketFileFunc == nil {
		panic("S3ClientMock.PutBucketFileFunc: method is nil but S3Client.PutBucketFile was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		File     *bytes.Reader
		Filename string
	}{
		Ctx:      ctx,
		File:     file,
		Filename: filename,
	}
	mock.lockPutBucketFile.Lock()
	mock.calls.PutBucketFile = append(mock.calls.PutBucketFile, callInfo)
	mock.lockPutBucketFile.Unlock()
	return mock.PutBucketFileFunc(ctx, file, filename)
}

// PutBucketFileCalls gets all the calls that were made to PutBucketFile.
// Check the length with:
//     len(mockedS3Client.PutBucketFileCalls())
func (mock *S3ClientMock) PutBucketFileCalls() []struct {
	Ctx      context.Context
	File     *bytes.Reader
	Filename string
} {
	var calls []struct {
		Ctx      context.Context
		File     *bytes.Reader
		Filename string
	}
	mock.lockPutBucketFile.RLock()
	calls = mock.calls.PutBucketFile
	mock.lockPutBucketFile.RUnlock()
	return calls
}