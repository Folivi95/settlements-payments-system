package auth

import (
	"net/http"
	"regexp"
	"strings"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"
)

func nonLocalAuthMiddleware(handler http.Handler, authorisedUsers AuthorisedUsers, enablePostPaymentsEndpoint bool, permittedTestUsers []string) http.Handler {
	postPaymentsAPI := regexp.MustCompile(`^/payments$`)
	healthCheck := regexp.MustCompile(`^/health_check$`)
	metrics := regexp.MustCompile(`^/metrics$`)
	members := regexp.MustCompile(`^/internal/team$`)
	testAPI := regexp.MustCompile(`^/test`)
	handlerWithBearerAuth := withBearerAuth(handler, authorisedUsers)
	handlerWithTestAuthorisation := withTestAuthorisation(handler, authorisedUsers, permittedTestUsers)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// disable the POST /payments endpoint if configured so
		if postPaymentsAPI.MatchString(r.URL.Path) && !enablePostPaymentsEndpoint {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if healthCheck.MatchString(r.URL.Path) || metrics.MatchString(r.URL.Path) || members.MatchString(r.URL.Path) {
			handler.ServeHTTP(w, r)
			return
		}

		if testAPI.MatchString(r.URL.Path) {
			handlerWithTestAuthorisation.ServeHTTP(w, r)
			return
		}

		handlerWithBearerAuth.ServeHTTP(w, r)
	})
}

func noOpAuthenticationMiddleWare(handler http.Handler) http.Handler {
	return handler
}

type AuthorisedUsers = map[string]string

func withBearerAuth(handler http.Handler, authorisedUsers AuthorisedUsers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("Authorization")

		if value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(value, " ")
		token := bearerToken[1]

		username := authorisedUsers[token]
		if username == "" {
			http.Error(w, "Authorization bearer token is invalid", http.StatusForbidden)
			return
		}
		zapctx.Info(r.Context(), "Request authenticated", zap.String("Path", r.URL.Path), zap.String("User", username), zap.String("Method", r.Method))

		handler.ServeHTTP(w, r)
	})
}

func withTestAuthorisation(handler http.Handler, authorisedUsers AuthorisedUsers, permittedTestUsers []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("Authorization")

		if value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(value, " ")
		token := bearerToken[1]

		username := authorisedUsers[token]
		if username == "" {
			http.Error(w, "Authorization bearer token is invalid", http.StatusForbidden)
			return
		}

		for _, permittedUser := range permittedTestUsers {
			if username == permittedUser {
				zapctx.Info(r.Context(), "Request authenticated", zap.String("Path", r.URL.Path), zap.String("User", username), zap.String("Method", r.Method))
				handler.ServeHTTP(w, r)
				return
			}
		}
		zapctx.Warn(r.Context(), "Request unauthorized", zap.String("Path", r.URL.Path), zap.String("User", username), zap.String("Method", r.Method))
		http.Error(w, "User is not permitted to make this request", http.StatusUnauthorized)
	})
}
