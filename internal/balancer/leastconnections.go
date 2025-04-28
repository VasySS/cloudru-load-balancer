package balancer

import "sync"

var _ Balancer = (*LeastConnections)(nil)

// LeastConnections implements least connections balancing.
type LeastConnections struct {
	mu       sync.Mutex
	backends []*BackendServer
}

// NewLeastConnections creates a new LeastConnections balancer.
func NewLeastConnections(backends []*BackendServer) *LeastConnections {
	return &LeastConnections{
		backends: backends,
	}
}

// Next gets next backend server.
func (lc *LeastConnections) Next() (*BackendServer, error) {
	//nolint:nilnil
	return nil, nil
}

// UpdateBackends updates the list of available backends.
func (lc *LeastConnections) UpdateBackends(_ []*BackendServer) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
}
