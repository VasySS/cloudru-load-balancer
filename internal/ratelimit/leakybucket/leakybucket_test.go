package leakybucket_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/leakybucket"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/leakybucket/mocks"
)

func TestClientAllowed_Overflow(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewRepository(t)
	lb := leakybucket.NewUserBucket(mockRepo, 2, 1, time.Second*2)

	const user = "user1"

	allowed := lb.ClientAllowed(user)
	assert.True(t, allowed, "expected first request to be allowed")

	allowed = lb.ClientAllowed(user)
	assert.True(t, allowed, "expected second request to be allowed")

	allowed = lb.ClientAllowed(user)
	assert.False(t, allowed, "expected third request to be denied")

	time.Sleep(time.Second * 3)

	allowed = lb.ClientAllowed(user)
	assert.True(t, allowed, "expected client to be allowed after tokens are leaked")
}

func TestClientAllowed_AfterLeaking(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewRepository(t)
	lb := leakybucket.NewUserBucket(mockRepo, 100, 1, time.Second*2)

	id := "user1"

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			allowed := lb.ClientAllowed(id)
			assert.True(t, allowed)
		}()
	}

	wg.Wait()

	time.Sleep(time.Second * 3)

	allowed := lb.ClientAllowed(id)
	assert.True(t, allowed)
}

func TestClientAllowed_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewRepository(t)
	lb := leakybucket.NewUserBucket(mockRepo, 100, 1, time.Second*2)

	const (
		user1 = "user1"
		user2 = "user2"
	)

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			assert.True(t, lb.ClientAllowed(user1))
		}()

		go func() {
			defer wg.Done()

			assert.True(t, lb.ClientAllowed(user2))
		}()
	}

	wg.Wait()

	assert.False(t, lb.ClientAllowed(user1))
	assert.False(t, lb.ClientAllowed(user2))

	time.Sleep(time.Second * 3)

	assert.True(t, lb.ClientAllowed(user1))
	assert.True(t, lb.ClientAllowed(user2))
}
