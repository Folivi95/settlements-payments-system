package middleware

import "net/http"

type HTTPMiddleware func(handler http.Handler) http.Handler
