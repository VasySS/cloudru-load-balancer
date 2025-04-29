// Package balancer contains algorithms for balancing incoming requests between backends.
package balancer

import (
	"net/http"
)

// BackendServer defines the interface for backend servers.
type BackendServer interface {
	Healthy() bool
	GetConnections() int64
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// Balancer defines an interface for balancing the load between backends.
type Balancer interface {
	Next() (BackendServer, error)
	UpdateBackends(backends []BackendServer)
}
