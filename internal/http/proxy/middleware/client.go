package middleware

import (
	"context"
	"net/http"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

const rateLimitKeyHeader = "Rate-Limit-Key"

// ClientCtxKey is a context key, used for retrieving client from context.
type ClientCtxKey struct{}

// ClientExtractor is middleware for extracting client from the request.
func ClientExtractor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		headerKey := r.Header.Get(rateLimitKeyHeader)
		if headerKey != "" {
			newCtx := context.WithValue(ctx, ClientCtxKey{}, ratelimit.ClientInfo{Identifier: headerKey})

			next.ServeHTTP(w, r.WithContext(newCtx))
		}
	})
}
