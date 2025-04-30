// Package backend contains logic for creating backend servers.
package backend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

// Backend represents a server, which accepts requests from load balancer.
type Backend struct {
	url          *url.URL
	healthy      atomic.Bool
	healthTicker *time.Ticker
	connections  atomic.Int64
	proxy        *httputil.ReverseProxy
}

// Address returns the url of a backend.
func (b *Backend) Address() *url.URL {
	return b.url
}

// Healthy returns current health status (atomic).
func (b *Backend) Healthy() bool {
	return b.healthy.Load()
}

// StartHealthChecks starts the periodic health checks for the backend on "/health" path.
func (b *Backend) StartHealthChecks(ctx context.Context) {
	select {
	case <-ctx.Done():
		b.healthTicker.Stop()

		return
	case <-b.healthTicker.C:
		// not going to make a custom http client for this
		//nolint:noctx
		resp, err := http.Get(b.url.String() + "/health")
		if err != nil {
			b.healthy.Store(false)
			return
		}

		//nolint:errcheck
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b.healthy.Store(false)
			return
		}

		b.healthy.Store(true)
	}
}

// GetConnections returns current connections count (atomic).
func (b *Backend) GetConnections() int64 {
	return b.connections.Load()
}

// ServeHTTP passes the request to the backend server using reverse proxy.
func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.connections.Add(1)
	defer b.connections.Add(-1)

	b.proxy.ServeHTTP(w, r)
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (*Backend, error)
	UpdateBackends(backends []*Backend)
}

// NewBackendServers creates an array of backend servers from config URLs and starts health checks on them.
func NewBackendServers(ctx context.Context, backends []string, healthCheckInterval time.Duration) ([]*Backend, error) {
	res := make([]*Backend, 0, len(backends))

	for _, b := range backends {
		parsedURL, err := url.Parse(b)
		if err != nil {
			return nil, fmt.Errorf("error parsing backend url: %w", err)
		}

		srv := &Backend{
			url:          parsedURL,
			proxy:        httputil.NewSingleHostReverseProxy(parsedURL),
			healthTicker: time.NewTicker(healthCheckInterval),
		}

		srv.healthy.Store(true)

		go srv.StartHealthChecks(ctx)

		res = append(res, srv)
	}

	return res, nil
}
