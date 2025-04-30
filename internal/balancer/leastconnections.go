package balancer

import (
	"errors"
	"log/slog"
	"math"
	"sync/atomic"
)

var (
	// ErrNoBackends is returned when there are no backends available (none were set).
	ErrNoBackends = errors.New("no backends available")
	// ErrNoHealthyBackends is returned when there are no healthy backends available.
	ErrNoHealthyBackends = errors.New("no healthy backends available")
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
//
//nolint:ireturn
func (lc *LeastConnections) Next() (BackendServer, error) {
	backends := *lc.backends.Load()

	if len(backends) == 0 {
		return nil, ErrNoBackends
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
		return nil, ErrNoHealthyBackends
	}

	slog.Debug("selected backend with least connections",
		slog.String("addr", selected.Address().Host),
		slog.Int64("connections", selected.GetConnections()),
	)

	return selected, nil
}

// UpdateBackends updates the list of available backends.
//
//nolint:ireturn
func (lc *LeastConnections) UpdateBackends(backends []BackendServer) {
	// create a new slice and copy to prevent external modification
	copied := make([]BackendServer, len(backends))
	copy(copied, backends)

	lc.backends.Store(&copied)
}
