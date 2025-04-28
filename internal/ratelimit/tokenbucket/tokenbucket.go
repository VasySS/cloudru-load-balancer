// Package tokenbucket implements token bucket algorithm.
package tokenbucket

import (
	"context"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// Repository defines an interface to save client data.
type Repository interface {
	SaveClient(ctx context.Context, client ratelimit.ClientInfo) error
}

// TokenBucket implements a token bucket algorithm.
type TokenBucket struct {
	repo Repository
}

// New creates a new TokenBucket.
func New(repo Repository) *TokenBucket {
	return &TokenBucket{
		repo: repo,
	}
}

// ClientAllowed checks if client has tokens and allows/reject a request.
func (b *TokenBucket) ClientAllowed(_ ratelimit.ClientInfo) bool {
	// TODO ...
	return true
}
