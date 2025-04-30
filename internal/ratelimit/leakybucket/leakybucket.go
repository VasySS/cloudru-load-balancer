// Package leakybucket implements leaky bucket algorithm.
package leakybucket

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// Repository defines an interface to save client data.
type Repository interface {
	SaveClient(ctx context.Context, client ratelimit.ClientInfo) error
}

type bucket struct {
	mu          sync.Mutex
	tokens      int
	capacity    int
	leakRate    int
	lastUpdated time.Time
}

// UserBucket implements a leaky bucket algorihtm per user.
type UserBucket struct {
	repo     Repository
	capacity int
	leakRate int
	mu       sync.RWMutex
	buckets  map[string]*bucket
}

// NewUserBucket creates a new LeakyBucket.
func NewUserBucket(repo Repository, capacity, leakRate int) *UserBucket {
	lb := &UserBucket{
		repo:     repo,
		capacity: capacity,
		leakRate: leakRate,
		buckets:  make(map[string]*bucket),
	}

	return lb
}

// ClientAllowed checks if client is allowed to make a request.
func (lb *UserBucket) ClientAllowed(identifier string) bool {
	b := lb.getOrCreateBucket(identifier)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UTC()
	elapsed := now.Sub(b.lastUpdated)

	leakedTokens := int(elapsed.Seconds() * float64(b.leakRate))
	newTokens := max(b.tokens-leakedTokens, 0)

	if newTokens+1 > b.capacity {
		slog.Debug("leaky bucket overflow", slog.String("client", identifier))

		return false
	}

	b.tokens = newTokens + 1
	b.lastUpdated = now

	slog.Debug("request allowed by leaky bucket",
		slog.String("client", identifier),
		slog.Int("current tokens", b.tokens),
	)

	return true
}

// getOrCreateBucket retrieves or creates a bucket for the client.
func (lb *UserBucket) getOrCreateBucket(identifier string) *bucket {
	lb.mu.RLock()
	existingBucket, ok := lb.buckets[identifier]
	lb.mu.RUnlock()

	if ok {
		return existingBucket
	}

	// TODO - add data retrieval from db
	capacity := lb.capacity
	leakRate := lb.leakRate

	newBucket := &bucket{
		tokens:      capacity,
		capacity:    capacity,
		leakRate:    leakRate,
		lastUpdated: time.Now().UTC(),
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	// check if the bucket was created between locks
	if existing, ok := lb.buckets[identifier]; ok {
		return existing
	}

	lb.buckets[identifier] = newBucket

	return newBucket
}
