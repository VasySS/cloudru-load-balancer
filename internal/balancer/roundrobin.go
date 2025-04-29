package balancer

import "sync"

var _ Balancer = (*RoundRobin)(nil)

// RoundRobin implements round robin balancing.
type RoundRobin struct {
	mu       sync.Mutex
	backends []BackendServer
}

// NewRoundRobin creates a new RoundRobin balancer.
func NewRoundRobin(backends []BackendServer) *RoundRobin {
	return &RoundRobin{
		backends: backends,
	}
}

// Next gets next backend server.
//
//nolint:ireturn
func (rb *RoundRobin) Next() (BackendServer, error) {
	//nolint:nilnil
	return nil, nil
}

// UpdateBackends updates the list of available backends.
func (rb *RoundRobin) UpdateBackends(_ []BackendServer) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
}
