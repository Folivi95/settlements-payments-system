//go:build unit
// +build unit

package authtoken_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client/authtoken"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client/authtoken/mocks"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

const (
	firstCall  = 0
	secondCall = 1
)

func TestAuthTokenService_GetAccessToken(t *testing.T) {
	var (
		username                       = testhelpers.RandomString()
		password                       = testhelpers.RandomString()
		expiresIn                      = 300
		TokenIntervalBeforeExpire      = time.Second * time.Duration(60)
		httpClient                     = &http.Client{}
		actualUserName, actualPassword string
		authServerState                int
		authCalls                      int
		tokenHasExpired                = make(chan struct{})
		authHasComplete                = make(chan struct{})

		firstAuthResponse = http_client.AuthResponseDto{
			AccessToken: testhelpers.RandomString(),
			ExpiresIn:   fmt.Sprint(expiresIn),
		}

		secondAuthResponse = http_client.AuthResponseDto{
			AccessToken: testhelpers.RandomString(),
			ExpiresIn:   fmt.Sprint(expiresIn),
		}
	)

	t.Run("it gets the token from the API, using basic auth and will schedule a refresh", func(t *testing.T) {
		is := is.New(t)
		ctx := context.Background()

		authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCalls++
			actualUserName, actualPassword, _ = r.BasicAuth()

			switch authServerState {
			case firstCall:
				_ = json.NewEncoder(w).Encode(firstAuthResponse)
			case secondCall:
				_ = json.NewEncoder(w).Encode(secondAuthResponse)
			}
		}))
		defer authServer.Close()

		spyScheduler := &mocks.SchedulerMock{AfterFuncFunc: func(d time.Duration, authTokenRefresh func()) *time.Timer {
			go func() {
				<-tokenHasExpired
				authServerState++
				authTokenRefresh()
				authHasComplete <- struct{}{}
			}()
			return time.NewTimer(1 * time.Second)
		}}

		authService := authtoken.NewAuthTokenService(authtoken.AuthConfig{
			AuthorizationBaseURL:      authServer.URL,
			APIUsername:               username,
			APIPassword:               password,
			TokenIntervalBeforeExpire: TokenIntervalBeforeExpire, // todo: int64 really?
			Scheduler:                 spyScheduler,
		}, httpClient, testdoubles.DummyMetricsClient{})

		// first time we get token it works and we schedule an update according to expiry
		actualToken, err := authService.GetAccessToken(ctx)
		is.NoErr(err)

		is.Equal(actualToken, firstAuthResponse.AccessToken)

		is.Equal(actualUserName, username) // basic auth username sent
		is.Equal(actualPassword, password) // basic auth password sent

		is.Equal(authCalls, 1)                          // got token from auth once
		is.Equal(len(spyScheduler.AfterFuncCalls()), 1) // refresh was scheduled
		expectedScheduleTime := time.Second * time.Duration(expiresIn-int(TokenIntervalBeforeExpire.Seconds()))
		is.Equal(spyScheduler.AfterFuncCalls()[0].D, expectedScheduleTime) // refresh scheduled at correct time

		// subsequent calls are cached, don't reschedule another update
		actualToken, err = authService.GetAccessToken(ctx)
		is.NoErr(err)

		is.Equal(actualToken, firstAuthResponse.AccessToken)
		is.Equal(authCalls, 1)                          // still only one call to auth
		is.Equal(len(spyScheduler.AfterFuncCalls()), 1) // no further refreshes scheduled

		// once token expires, we get a new token from auth and schedule again
		tokenHasExpired <- struct{}{}
		<-authHasComplete

		actualToken, err = authService.GetAccessToken(ctx)
		is.NoErr(err)

		is.Equal(actualToken, secondAuthResponse.AccessToken)

		is.Equal(authCalls, 2)                          // got token from auth twice
		is.Equal(len(spyScheduler.AfterFuncCalls()), 2) // another refresh scheduled
	})

	t.Run("return an error if the auth server fails", func(t *testing.T) {
		var (
			is  = is.New(t)
			ctx = context.Background()
		)

		authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		}))

		authService := authtoken.NewAuthTokenService(authtoken.AuthConfig{
			AuthorizationBaseURL:      authServer.URL,
			APIUsername:               username,
			APIPassword:               password,
			TokenIntervalBeforeExpire: TokenIntervalBeforeExpire,
			Scheduler:                 &mocks.SchedulerMock{},
		}, httpClient, testdoubles.DummyMetricsClient{})

		actualToken, err := authService.GetAccessToken(ctx)
		is.True(err != nil)
		is.Equal(actualToken, "")
	})

	t.Run("return an error if auth server returns trash", func(t *testing.T) {
		var (
			ctx = context.Background()
			is  = is.New(t)
		)

		authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, testhelpers.RandomString())
		}))

		authService := authtoken.NewAuthTokenService(authtoken.AuthConfig{
			AuthorizationBaseURL:      authServer.URL,
			APIUsername:               username,
			APIPassword:               password,
			TokenIntervalBeforeExpire: TokenIntervalBeforeExpire,
			Scheduler:                 &mocks.SchedulerMock{},
		}, httpClient, testdoubles.DummyMetricsClient{})

		actualToken, err := authService.GetAccessToken(ctx)
		is.True(err != nil)
		is.Equal(actualToken, "")
	})
}
