// Package balancer contains algorithms for balancing incoming requests between backends.
package balancer

import (
	"net/http"
)

// BackendServer defines the interface for backend servers.
//
//go:generate go tool mockery --name=BackendServer
type BackendServer interface {
	Healthy() bool
	GetConnections() int64
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (BackendServer, error)
	UpdateBackends(backends []BackendServer)
}
