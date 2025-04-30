// Package proxy contains logic for creating a reverse proxy and handling incoming requests.
package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/VasySS/cloudru-load-balancer/internal/balancer"
	"github.com/VasySS/cloudru-load-balancer/internal/http/proxy/middleware"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// ResponseError struct implements RFC 7807/RFC 9457 for http error responses.
type ResponseError struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func (e ResponseError) Error() string {
	return e.Title + ": " + e.Detail
}

func writeError(w http.ResponseWriter, title, detail string, status int) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	err := ResponseError{
		Title:  title,
		Status: status,
		Detail: detail,
	}

	//nolint:errchkjson
	_ = json.NewEncoder(w).Encode(err)
}

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
			writeError(w,
				"Server error",
				"Unable to identify client",
				http.StatusInternalServerError,
			)

			return
		}

		if !limiter.ClientAllowed(clientInfo) {
			writeError(w,
				"Rate limit exceeded",
				"Rate limit exceeded for this client, try again later",
				http.StatusTooManyRequests,
			)

			return
		}

		targetBackend, err := balancer.Next()
		if err != nil {
			// http.Error(w, "no available backends", http.StatusServiceUnavailable)
			writeError(w,
				"Server error",
				"Unable to find available backend",
				http.StatusServiceUnavailable,
			)

			return
		}

		r.Host = targetBackend.Address().Host

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
