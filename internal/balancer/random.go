package balancer

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
)

var _ Balancer = (*Random)(nil)

// Random implements random balancing.
type Random struct {
	mu       sync.Mutex
	backends []*BackendServer
}

// NewRandom creates a new Random balancer.
func NewRandom(backends []*BackendServer) *Random {
	return &Random{
		backends: backends,
	}
}

// Next returns a random backend server.
func (r *Random) Next() (*BackendServer, error) {
	nextServerIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(r.backends))))
	if err != nil {
		return nil, fmt.Errorf("error getting random backend: %w", err)
	}

	return r.backends[nextServerIdx.Int64()], nil
}

// UpdateBackends updates the list of available backends.
func (r *Random) UpdateBackends(backends []*BackendServer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.backends = backends
}
