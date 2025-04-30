// Package proxy contains logic for creating a reverse proxy and handling incoming requests.
package proxy

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/VasySS/cloudru-load-balancer/internal/balancer"
	"github.com/VasySS/cloudru-load-balancer/internal/http/proxy/middleware"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// Server implements ServeHTTP interface and represents a reverse proxy server.
type Server struct {
	mux      *chi.Mux
	limiter  ratelimit.Limiter
	balancer balancer.Balancer
}

// New creates a new reverse proxy with rate limiter and balancer.
func New(limiter ratelimit.Limiter, balancer balancer.Balancer) *Server {
	mux := chi.NewMux()

	mux.Use(
		chiMiddleware.Heartbeat("/health"),
		chiMiddleware.RequestID,
		middleware.Logger,
		chiMiddleware.Recoverer,
		middleware.ClientExtractor,
		chiMiddleware.CleanPath,
		chiMiddleware.StripSlashes,
		chiMiddleware.Compress(5),
	)

	mux.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientInfo, ok := r.Context().Value(middleware.ClientCtxKey{}).(string)
		if !ok {
			http.Error(w, "unable to identify client", http.StatusInternalServerError)
			return
		}

		if !limiter.ClientAllowed(clientInfo) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		targetBackend, err := balancer.Next()
		if err != nil {
			http.Error(w, "no available backends", http.StatusServiceUnavailable)
			return
		}

		targetBackend.ServeHTTP(w, r)
	}))

	return &Server{
		mux:      mux,
		limiter:  limiter,
		balancer: balancer,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
