//go:build unit
// +build unit

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	is2 "github.com/matryer/is"
	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/internal/adapters/env"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/middleware/auth"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

var (
	bearerToken     = testhelpers.RandomString()
	authorisedUsers = map[string]string{bearerToken: testhelpers.RandomString()}
	expectedStatus  = http.StatusTeapot
	stubHandler     = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(expectedStatus)
	})
)

func TestNewAuthMiddleWare(t *testing.T) {
	var (
		unauthorisedRequestToHealthcheck    = httptest.NewRequest(http.MethodGet, "/health_check", nil)
		unauthorisedRequestToCreatePayment  = httptest.NewRequest(http.MethodPost, "/payments", nil)
		unauthorisedRequestToGetPaymentByID = httptest.NewRequest(http.MethodGet, "/payments/123", nil)
		unauthorisedRequestToUnprocessedDLQ = httptest.NewRequest(http.MethodGet, "/internal/dead-letter-queues/unprocessed", nil)
		unauthorisedRequestBCReport         = httptest.NewRequest(http.MethodGet, "/bc-report/2021-10-11", nil)
		unauthorisedRequestCurrencyReport   = httptest.NewRequest(http.MethodGet, "/payments/currencies-report", nil)
		unauthorisedRequestMid              = httptest.NewRequest(http.MethodGet, "/mid/123/2021-10-11", nil)
		unauthorisedRequestToReplayPayment  = httptest.NewRequest(http.MethodPost, "/replay-payment?action=pay_currency_from_file&currency=EUR&file=test-ufx.xml", nil)
		authorisedRequestToCreatePayment    = reqWithBearer(httptest.NewRequest(http.MethodPost, "/payments", nil), bearerToken)
		authorisedRequestToGetPayment       = reqWithBearer(httptest.NewRequest(http.MethodGet, "/payments/123", nil), bearerToken)
		authorisedRequestToUnprocessedDLQ   = reqWithBearer(httptest.NewRequest(http.MethodGet, "/internal/dead-letter-queues/unprocessed", nil), bearerToken)
		authorisedRequestBCReport           = reqWithBearer(httptest.NewRequest(http.MethodGet, "/bc-report/2021-10-11", nil), bearerToken)
		authorisedRequestCurrencyReport     = reqWithBearer(httptest.NewRequest(http.MethodGet, "/payments/currencies-report", nil), bearerToken)
		authorisedRequestMid                = reqWithBearer(httptest.NewRequest(http.MethodGet, "/mid/123/2021-10-11", nil), bearerToken)
		authorisedRequestToReplayPayment    = reqWithBearer(httptest.NewRequest(http.MethodPost, "/replay-payment?action=pay_currency_from_file&currency=EUR&file=test-ufx.xml", nil), bearerToken)

		assertWeHadAccess = func(t *testing.T, res *httptest.ResponseRecorder) {
			t.Helper()
			if res.Code != expectedStatus {
				t.Errorf("expected %d, got %d", expectedStatus, res.Code)
			}
		}
	)

	t.Run("In a local env", func(t *testing.T) {
		authMiddleware, _ := auth.NewAuthMiddleWare(env.Local, false, authorisedUsers, nil)
		localAuthMiddleWare := authMiddleware(stubHandler)

		t.Run("we have access to healthcheck", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, authorisedRequestToCreatePayment)
			assertWeHadAccess(t, res)
		})

		t.Run("we have access to GET payment by id", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestToGetPaymentByID)
			assertWeHadAccess(t, res)
		})

		t.Run("we have access to dead letter queue information", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestToUnprocessedDLQ)
			assertWeHadAccess(t, res)
		})
		t.Run("we have access to banking circle report", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestBCReport)
			assertWeHadAccess(t, res)
		})
		t.Run("we have access to currencies report", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestCurrencyReport)
			assertWeHadAccess(t, res)
		})
		t.Run("we have access to mid by date", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestMid)
			assertWeHadAccess(t, res)
		})
		t.Run("we have access to replay payment", func(t *testing.T) {
			res := httptest.NewRecorder()
			localAuthMiddleWare.ServeHTTP(res, unauthorisedRequestToReplayPayment)
			assertWeHadAccess(t, res)
		})
	})

	t.Run("in an integration env", func(t *testing.T) {
		middleware, _ := auth.NewAuthMiddleWare(env.IntegrationGlobalPlatform, false, authorisedUsers, nil)
		integrationAuthMiddleware := middleware(stubHandler)

		t.Run("we have access to /health_check without auth", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestToHealthcheck)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users cannot create payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestToCreatePayment)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("unauthorised users cannot get payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestToGetPaymentByID)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can get payments", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestToGetPayment)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users  view dlq information", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestToUnprocessedDLQ)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can view dlq information", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestToUnprocessedDLQ)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users cannot view banking circle's rejection report", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestBCReport)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can view banking circle's rejection report", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestBCReport)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access a currency report", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestCurrencyReport)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to currency report", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestCurrencyReport)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access a mid data", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestMid)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to a mid data", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestMid)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access to replay payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, unauthorisedRequestToReplayPayment)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to replay payments", func(t *testing.T) {
			res := httptest.NewRecorder()
			integrationAuthMiddleware.ServeHTTP(res, authorisedRequestToReplayPayment)
			assertWeHadAccess(t, res)
		})
	})

	t.Run("in a production env", func(t *testing.T) {
		middleWare, _ := auth.NewAuthMiddleWare(env.ProductionGlobalPlatform, false, authorisedUsers, nil)
		productionAuthMiddleware := middleWare(stubHandler)

		t.Run("we have access to /health_check without auth", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestToHealthcheck)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users cannot create payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestToCreatePayment)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("unauthorised users cannot get payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestToGetPaymentByID)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can get payments", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestToGetPayment)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users cannot view dlq information", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestToUnprocessedDLQ)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can view dlq information", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestToUnprocessedDLQ)
			assertWeHadAccess(t, res)
		})

		t.Run("unauthorised users cannot view banking circle's rejection report", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestBCReport)
			is.Equal(res.Code, http.StatusUnauthorized)
		})

		t.Run("authorised users can view banking circle's rejection report", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestBCReport)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access a currency report", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestCurrencyReport)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to currency report", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestCurrencyReport)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access a mid data", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestMid)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to a mid data", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestMid)
			assertWeHadAccess(t, res)
		})
		t.Run("unauthorised users cannot access to replay payments", func(t *testing.T) {
			is := is2.New(t)
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, unauthorisedRequestToReplayPayment)
			is.Equal(res.Code, http.StatusUnauthorized)
		})
		t.Run("authorized users have access to replay payments", func(t *testing.T) {
			res := httptest.NewRecorder()
			productionAuthMiddleware.ServeHTTP(res, authorisedRequestToReplayPayment)
			assertWeHadAccess(t, res)
		})
		t.Run("if the bearer token is wrong, forbid the user", func(t *testing.T) {
			is := is2.New(t)

			res := httptest.NewRecorder()
			forbiddenRequestToCreatePayment := httptest.NewRequest(http.MethodGet, "/payments/123", nil)
			forbiddenRequestToCreatePayment.Header.Set("Authorization", "Bearer "+testhelpers.RandomString())

			productionAuthMiddleware.ServeHTTP(res, forbiddenRequestToCreatePayment)
			is.Equal(res.Code, http.StatusForbidden)
		})
	})

	t.Run("only authorised users are allowed to make payments when endpoint is enabled", func(t *testing.T) {
		td := []struct {
			name              string
			isAuthorisedUser  bool
			isEndpointEnabled bool
			expectedStatus    int
		}{
			{
				name:              "authorised user, endpoint enabled; allow access",
				isAuthorisedUser:  true,
				isEndpointEnabled: true,
				expectedStatus:    expectedStatus,
			},
			{
				name:              "authorised user, endpoint disabled; deny access",
				isAuthorisedUser:  true,
				isEndpointEnabled: false,
				expectedStatus:    http.StatusUnauthorized,
			},
			{
				name:              "unauthorised user, endpoint enabled; deny access",
				isAuthorisedUser:  false,
				isEndpointEnabled: true,
				expectedStatus:    http.StatusUnauthorized,
			},
			{
				name:              "unauthorised user, endpoint disabled; deny access",
				isAuthorisedUser:  false,
				isEndpointEnabled: false,
				expectedStatus:    http.StatusUnauthorized,
			},
		}

		for _, tdd := range td {
			t.Run(tdd.name, func(t *testing.T) {
				is := is2.New(t)

				// the authorised or unauthorised req
				request := unauthorisedRequestToCreatePayment
				if tdd.isAuthorisedUser {
					request = authorisedRequestToCreatePayment
				}

				middleWare, _ := auth.NewAuthMiddleWare(env.ProductionGlobalPlatform, tdd.isEndpointEnabled, authorisedUsers, nil)
				authMiddleware := middleWare(stubHandler)

				res := httptest.NewRecorder()
				authMiddleware.ServeHTTP(res, request)
				is.Equal(res.Code, tdd.expectedStatus)
			})
		}
	})

	t.Run("authorised user is allowed to access test endpoints", func(t *testing.T) {
		// Given a user with a token
		username := "testUser"
		token := "testToken"

		// And given that token is contained by the authorised users map
		authorisedUsers := map[string]string{token: username}

		// And given that user is permitted to call test endpoints
		permittedTestUsers := []string{username}

		// And we are on production with proper middleware setup
		authMiddleWare, _ := auth.NewAuthMiddleWare(env.ProductionGlobalPlatform, true, authorisedUsers, permittedTestUsers)
		handler := authMiddleWare(stubHandler)

		// When request is made to test endpoint with the token
		request := reqWithBearer(httptest.NewRequest(http.MethodPost, "/test/upload/file", nil), token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		// Then response should be teapot
		assert.Equal(t, expectedStatus, recorder.Code)
	})

	t.Run("unauthorised user is not allowed to access test endpoints", func(t *testing.T) {
		// Given a user with a token
		username := "testUser"
		token := "testToken"

		// And given that token is contained by the authorised users map
		authorisedUsers := map[string]string{token: username}

		// And given that user is not permitted to call test endpoints
		permittedTestUsers := []string{"anotherUser"}

		// And we are on production with proper middleware setup
		authMiddleWare, _ := auth.NewAuthMiddleWare(env.ProductionGlobalPlatform, true, authorisedUsers, permittedTestUsers)
		handler := authMiddleWare(stubHandler)

		// When request is made to test endpoint with the token
		request := reqWithBearer(httptest.NewRequest(http.MethodPost, "/test/upload/file", nil), token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		// Then response should be 401 Unauthorised
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func BenchmarkAllowOperationalEndpointsMiddleware(b *testing.B) {
	middleWare, _ := auth.NewAuthMiddleWare(env.ProductionGlobalPlatform, false, authorisedUsers, nil)
	productionAuthMiddleware := middleWare(stubHandler)
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payments/123", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		productionAuthMiddleware.ServeHTTP(res, req)
	}
}

func reqWithBearer(req *http.Request, token string) *http.Request {
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}
