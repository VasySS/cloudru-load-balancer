// Package leakybucket implements leaky bucket algorithm.
package leakybucket

import (
	"context"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// Repository defines an interface to save client data.
type Repository interface {
	SaveClient(ctx context.Context, client ratelimit.ClientInfo) error
}

// LeakyBucket implements a leaky bucket algorihtm.
type LeakyBucket struct {
	repo Repository
}

// New creates a new LeakyBucket.
func New(repo Repository) *LeakyBucket {
	return &LeakyBucket{
		repo: repo,
	}
}

// ClientAllowed checks if client is allowed to make a request.
func (b *LeakyBucket) ClientAllowed(_ ratelimit.ClientInfo) bool {
	// TODO ...
	return true
}
