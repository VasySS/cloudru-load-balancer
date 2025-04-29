package balancer

import (
	"errors"
	"math"
	"sync/atomic"
)

var _ Balancer = (*LeastConnections)(nil)

// LeastConnections implements least connections balancing.
type LeastConnections struct {
	backends atomic.Pointer[[]BackendServer]
}

// NewLeastConnections creates a new LeastConnections balancer.
func NewLeastConnections(backends []BackendServer) *LeastConnections {
	lc := &LeastConnections{}
	lc.UpdateBackends(backends)

	return lc
}

// Next gets next backend server.
func (lc *LeastConnections) Next() (BackendServer, error) {
	backends := *lc.backends.Load()

	if len(backends) == 0 {
		return nil, errors.New("no backends available")
	}

	var selected BackendServer
	var minConns int64 = math.MaxInt64

	for _, backend := range backends {
		backendConns := backend.GetConnections()

		if backend.Healthy() && backendConns < minConns {
			selected = backend
			minConns = backendConns

			if minConns == 0 {
				break
			}
		}
	}

	if selected == nil {
		return nil, errors.New("no healthy backends available")
	}

	return selected, nil
}

// UpdateBackends updates the list of available backends.
func (lc *LeastConnections) UpdateBackends(backends []BackendServer) {
	// create a new slice and copy to prevent external modification
	copied := make([]BackendServer, len(backends))
	copy(copied, backends)

	lc.backends.Store(&copied)
}
