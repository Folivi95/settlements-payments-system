package auth

import (
	"fmt"
	"net/http"

	"github.com/saltpay/settlements-payments-system/internal/adapters/env"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/middleware"
)

func NewAuthMiddleWare(envName env.EnvName, enablePostPaymentEndpoint bool, authorisedUsers AuthorisedUsers, permittedTestUsers []string) (middleware.HTTPMiddleware, error) {
	switch envName {
	case env.ProductionGlobalPlatform:
		fallthrough
	case env.IntegrationGlobalPlatform:
		return func(handler http.Handler) http.Handler {
			return nonLocalAuthMiddleware(handler, authorisedUsers, enablePostPaymentEndpoint, permittedTestUsers)
		}, nil
	case env.Local, env.Tilt:
		return noOpAuthenticationMiddleWare, nil
	default:
		return nil, fmt.Errorf("unrecognised environment %q", envName)
	}
}
