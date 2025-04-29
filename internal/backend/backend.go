package backend

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

// BackendServer represents a server, which accepts requests from load balancer.
type BackendServer struct {
	URL         *url.URL
	healthy     atomic.Bool
	connections atomic.Int64
	proxy       *httputil.ReverseProxy
}

func (s *BackendServer) Healthy() bool {
	return s.healthy.Load()
}

// GetConnections returns current connections count (atomic).
func (s *BackendServer) GetConnections() int64 {
	return s.connections.Load()
}

// ServeHTTP passes the request to the backend server using reverse proxy.
func (s *BackendServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.connections.Add(1)
	defer s.connections.Add(-1)

	s.proxy.ServeHTTP(w, r)
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (*BackendServer, error)
	UpdateBackends(backends []*BackendServer)
}

// BackendServersFromArray creates an array of backend servers from config URLs.
func BackendServersFromArray(backends []string) ([]*BackendServer, error) {
	res := make([]*BackendServer, 0, len(backends))

	for _, b := range backends {
		parsedURL, err := url.Parse(b)
		if err != nil {
			return nil, fmt.Errorf("error parsing backend url: %w", err)
		}

		srv := &BackendServer{
			URL:   parsedURL,
			proxy: httputil.NewSingleHostReverseProxy(parsedURL),
		}

		srv.healthy.Store(true)

		res = append(res, srv)
	}

	return res, nil
}
