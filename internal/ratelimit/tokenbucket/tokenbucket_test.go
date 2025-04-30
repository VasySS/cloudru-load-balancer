package tokenbucket_test

import (
	"sync"
	"testing"
	"time"

	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/tokenbucket"
	"github.com/VasySS/cloudru-load-balancer/internal/ratelimit/tokenbucket/mocks"
	"github.com/stretchr/testify/assert"
)

func TestClientAllowed_TokenAvailable(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewRepository(t)
	tb := tokenbucket.NewUserBucket(mockRepo, 2, 1, time.Second)
	defer tb.Stop()

	id := "user1"

	allowed := tb.ClientAllowed(id)
	assert.True(t, allowed, "expected client to be allowed on first try")

	allowed = tb.ClientAllowed(id)
	assert.True(t, allowed, "expected client to be allowed on second try")

	allowed = tb.ClientAllowed(id)
	assert.False(t, allowed, "expected client to be denied on third attempt")

	time.Sleep(time.Second * 2)

	allowed = tb.ClientAllowed(id)
	assert.True(t, allowed, "expected client to be allowed after tokens are refilled")
}

func TestClientAllowed_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	mockRepo := mocks.NewRepository(t)

	tb := tokenbucket.NewUserBucket(mockRepo, 100, 1, time.Second*60)
	defer tb.Stop()

	const (
		user1 = "user1"
		user2 = "user2"
	)

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			assert.True(t, tb.ClientAllowed(user1))
		}()

		go func() {
			defer wg.Done()

			assert.True(t, tb.ClientAllowed(user2))
		}()
	}

	wg.Wait()

	assert.False(t, tb.ClientAllowed(user1))
	assert.False(t, tb.ClientAllowed(user2))
}
