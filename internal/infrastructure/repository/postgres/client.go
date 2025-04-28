package postgres

import (
	"context"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// SaveClient saves information about a client for load balancing.
func (r *Repository) SaveClient(_ context.Context, _ ratelimit.ClientInfo) error {
	return nil
}
