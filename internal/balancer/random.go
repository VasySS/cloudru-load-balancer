package balancer

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sync/atomic"
)

var _ Balancer = (*Random)(nil)

// Random implements random balancing.
type Random struct {
	backends atomic.Pointer[[]BackendServer]
}

// NewRandom creates a new Random balancer.
func NewRandom(backends []BackendServer) *Random {
	r := &Random{}
	r.UpdateBackends(backends)

	return r
}

// Next returns a random backend server.
//
//nolint:ireturn
func (r *Random) Next() (BackendServer, error) {
	backends := *r.backends.Load()

	nextServerIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(backends))))
	if err != nil {
		return nil, fmt.Errorf("error getting random backend: %w", err)
	}

	selected := backends[nextServerIdx.Int64()]

	slog.Debug("selected backend using random",
		slog.String("addr", selected.Address().Host),
	)

	return selected, nil
}

// UpdateBackends updates the list of available backends.
func (r *Random) UpdateBackends(backends []BackendServer) {
	// create a new slice and copy to prevent external modification
	copied := make([]BackendServer, len(backends))
	copy(copied, backends)

	r.backends.Store(&copied)
}
