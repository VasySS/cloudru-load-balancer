package balancer

import (
	"log/slog"
	"sync/atomic"
)

var _ Balancer = (*RoundRobin)(nil)

// RoundRobin implements round robin balancing.
type RoundRobin struct {
	counter  atomic.Uint64
	backends atomic.Pointer[[]BackendServer]
}

// NewRoundRobin creates a new RoundRobin balancer.
func NewRoundRobin(backends []BackendServer) *RoundRobin {
	rr := &RoundRobin{}
	rr.UpdateBackends(backends)

	return rr
}

// Next gets next backend server.
//
//nolint:ireturn
func (rr *RoundRobin) Next() (BackendServer, error) {
	backends := *rr.backends.Load()

	if len(backends) == 0 {
		return nil, ErrNoBackends
	}

	var selected BackendServer

	for range len(backends) {
		nextIdx := rr.counter.Load() % uint64(len(backends))
		nextBackend := backends[nextIdx]

		rr.counter.Add(1)

		if nextBackend.Healthy() {
			selected = nextBackend

			break
		}
	}

	if selected == nil {
		return nil, ErrNoHealthyBackends
	}

	slog.Debug("selected backend using round robin",
		slog.String("addr", selected.Address().Host),
	)

	return selected, nil
}

// UpdateBackends updates the list of available backends.
func (rr *RoundRobin) UpdateBackends(backends []BackendServer) {
	// create a new slice and copy to prevent external modification
	copied := make([]BackendServer, len(backends))
	copy(copied, backends)

	rr.backends.Store(&copied)
}
