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
		chiMiddleware.RequestID,
		middleware.Logger,
		chiMiddleware.Recoverer,
		middleware.ClientExtractor,
		chiMiddleware.CleanPath,
		chiMiddleware.StripSlashes,
		chiMiddleware.Compress(5),
	)

	// mux.Post("/clients", func(w http.ResponseWriter, r *http.Request) {
	// 	// ...
	// })

	return &Server{
		mux:      mux,
		limiter:  limiter,
		balancer: balancer,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clientInfo, ok := ctx.Value(middleware.ClientCtxKey{}).(string)
	if !ok {
		// TODO: add error struct
		return
	}

	if !s.limiter.ClientAllowed(clientInfo) {
		// TODO: add error struct
		return
	}

	targetBackend, err := s.balancer.Next()
	if err != nil {
		// TODO: add error struct
		return
	}

	targetBackend.ServeHTTP(w, r)
}
