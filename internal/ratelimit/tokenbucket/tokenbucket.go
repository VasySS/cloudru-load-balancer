// Package tokenbucket implements token bucket algorithm.
package tokenbucket

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit"
)

// Repository defines an interface to save client data.
//
//go:generate go tool mockery --name=Repository
type Repository interface {
	SaveClient(ctx context.Context, client ratelimit.ClientInfo) error
}

type bucket struct {
	capacity    atomic.Int64
	refillRate  atomic.Int64
	tokens      atomic.Int64
	lastUpdated atomic.Value // time.Time
}

// UserBucket implements a token bucket algorithm per user.
type UserBucket struct {
	repo       Repository
	capacity   int
	refillRate int
	mu         sync.RWMutex
	buckets    map[string]*bucket
	ticker     *time.Ticker
	stopChan   chan struct{}
}

// NewUserBucket creates new token bucket with individual user rate limits.
func NewUserBucket(repo Repository, capacity, refillRate int, refillInterval time.Duration) *UserBucket {
	tb := &UserBucket{
		repo:       repo,
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*bucket),
		ticker:     time.NewTicker(refillInterval),
		stopChan:   make(chan struct{}),
	}

	go tb.startRefiller()

	return tb
}

// Stop stops the bucket's refill.
func (tb *UserBucket) Stop() {
	close(tb.stopChan)
	tb.ticker.Stop()
}

// ClientAllowed checks if client is allowed to make a request.
func (tb *UserBucket) ClientAllowed(identifier string) bool {
	b := tb.getOrCreateBucket(identifier)

	// CAS loop
	for {
		current := b.tokens.Load()

		slog.Debug("token bucket check",
			slog.String("id", identifier),
			slog.Int64("tokens", current),
		)

		if current <= 0 {
			return false
		}

		if b.tokens.CompareAndSwap(current, current-1) {
			return true
		}
	}
}

func (tb *UserBucket) getOrCreateBucket(identifier string) *bucket {
	tb.mu.RLock()
	existingBucket, ok := tb.buckets[identifier]
	tb.mu.RUnlock()

	if ok {
		return existingBucket
	}

	// TODO - add data retrieval from db
	capacity := tb.capacity

	newBucket := &bucket{}
	newBucket.tokens.Store(int64(capacity))
	newBucket.capacity.Store(int64(capacity))
	newBucket.refillRate.Store(int64(tb.refillRate))
	newBucket.lastUpdated.Store(time.Now().UTC())

	tb.mu.Lock()
	defer tb.mu.Unlock()

	// check if the bucket was created between locks
	if existing, ok := tb.buckets[identifier]; ok {
		return existing
	}

	tb.buckets[identifier] = newBucket

	return newBucket
}

func (tb *UserBucket) startRefiller() {
	for {
		select {
		case <-tb.ticker.C:
			tb.refillBuckets()
		case <-tb.stopChan:
			return
		}
	}
}

func (tb *UserBucket) refillBuckets() {
	tb.mu.Lock()
	buckets := tb.buckets
	tb.mu.Unlock()

	now := time.Now().UTC()

	for _, bucket := range buckets {
		if bucket.tokens.Load() == bucket.capacity.Load() {
			continue
		}

		lastUpdated, ok := bucket.lastUpdated.Load().(time.Time)
		if !ok {
			return
		}

		elapsed := now.Sub(lastUpdated)
		tokensToAdd := int64(elapsed.Seconds()) * bucket.refillRate.Load()

		// CAS loop
		for {
			currentTokens := bucket.tokens.Load()
			newTokens := currentTokens + tokensToAdd

			if capacity := bucket.capacity.Load(); newTokens > capacity {
				newTokens = capacity
			}

			if bucket.tokens.CompareAndSwap(currentTokens, newTokens) {
				break
			}
		}

		bucket.lastUpdated.Store(now)
	}
}
