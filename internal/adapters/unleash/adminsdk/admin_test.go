//go:build unit
// +build unit

package adminsdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/matryer/is"
)

const (
	unleashServiceURL = "https://unleash.com/random/url"
	serviceToken      = "service token"
	featureName       = "dummy feature"
)

func TestAdminSDK_FeatureOn(t *testing.T) {
	is := is.New(t)

	t.Run("should toggle feature on successfully", func(t *testing.T) {
		var (
			ctx        = context.Background()
			httpClient = &httpClientMock{statusCode: 200}
			adminSDK   = NewAdminSDK(unleashServiceURL, serviceToken, httpClient)
		)

		err := adminSDK.FeatureOn(ctx, featureName)

		is.NoErr(err)
		is.Equal(httpClient.calls, 1)
	})

	t.Run("should return error when unleash http api returns status code != of 200", func(t *testing.T) {
		type test struct {
			name       string
			statusCode int
		}

		tests := []test{
			{
				name:       "when api returns 400 status code",
				statusCode: 400,
			},
			{
				name:       "when api returns 404 status code",
				statusCode: 404,
			},
			{
				name:       "when api returns 500 status code",
				statusCode: 500,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var (
					ctx        = context.Background()
					httpClient = &httpClientMock{statusCode: tc.statusCode}
					adminSDK   = NewAdminSDK(unleashServiceURL, serviceToken, httpClient)
				)

				expectation := fmt.Errorf("error status enabling feature (%s). status %d", featureName, tc.statusCode)

				err := adminSDK.FeatureOn(ctx, featureName)

				is.Equal(err, expectation)
				is.Equal(httpClient.calls, 1)
			})
		}
	})
}

func TestAdminSDK_FeatureOff(t *testing.T) {
	is := is.New(t)

	t.Run("should toggle feature off successfully", func(t *testing.T) {
		var (
			ctx        = context.Background()
			httpClient = &httpClientMock{statusCode: 200}
			adminSDK   = NewAdminSDK(unleashServiceURL, serviceToken, httpClient)
		)

		err := adminSDK.FeatureOff(ctx, featureName)

		is.NoErr(err)
		is.Equal(httpClient.calls, 1)
	})

	t.Run("should return error when unleash http api returns status code != of 200", func(t *testing.T) {
		type test struct {
			name       string
			statusCode int
		}

		tests := []test{
			{
				name:       "when api returns 400 status code",
				statusCode: 400,
			},
			{
				name:       "when api returns 404 status code",
				statusCode: 404,
			},
			{
				name:       "when api returns 500 status code",
				statusCode: 500,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				var (
					ctx        = context.Background()
					httpClient = &httpClientMock{statusCode: tc.statusCode}
					adminSDK   = NewAdminSDK(unleashServiceURL, serviceToken, httpClient)
				)

				expectation := fmt.Errorf("error status disabling feature (%s). status %d", featureName, tc.statusCode)

				err := adminSDK.FeatureOff(ctx, featureName)

				is.Equal(err, expectation)
				is.Equal(httpClient.calls, 1)
			})
		}
	})
}

type httpClientMock struct {
	DoFunc     func(req *http.Request) (*http.Response, error)
	calls      int
	statusCode int
}

func (c *httpClientMock) Do(_ *http.Request) (*http.Response, error) {
	c.calls++
	return &http.Response{
		StatusCode: c.statusCode,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}
