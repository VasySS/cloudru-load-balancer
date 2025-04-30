// Package balancer contains algorithms for balancing incoming requests between backends.
package balancer

import (
	"net/http"
	"net/url"
)

// BackendServer defines the interface for backend servers.
//
//go:generate go tool mockery --name=BackendServer
type BackendServer interface {
	Address() *url.URL
	Healthy() bool
	GetConnections() int64
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (BackendServer, error)
	UpdateBackends(backends []BackendServer)
}
