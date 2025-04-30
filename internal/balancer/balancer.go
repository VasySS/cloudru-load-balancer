// Package balancer contains algorithms for balancing incoming requests between backends.
package balancer

import (
	"errors"
	"net/http"
	"net/url"
)

var (
	// ErrNoBackends is returned when there are no backends available (none were set).
	ErrNoBackends = errors.New("no backends available")
	// ErrNoHealthyBackends is returned when there are no healthy backends available.
	ErrNoHealthyBackends = errors.New("no healthy backends available")
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
