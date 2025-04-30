package balancer_test

import (
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VasySS/cloudru-load-balancer/internal/balancer"
	"github.com/VasySS/cloudru-load-balancer/internal/balancer/mocks"
)

func TestLeastConnections(t *testing.T) {
	t.Parallel()

	t.Run("get a healthy backend with least connections", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		b1.On("GetConnections").Return(int64(30))
		b1.On("Healthy").Return(true)
		b1.On("Address").Return(&url.URL{Host: "backend1"}).Maybe()

		b2 := mocks.NewBackendServer(t)
		b2.On("GetConnections").Return(int64(10))
		b2.On("Healthy").Return(false)
		b2.On("Address").Return(&url.URL{Host: "backend2"}).Maybe()

		b3 := mocks.NewBackendServer(t)
		b3.On("GetConnections").Return(int64(20))
		b3.On("Healthy").Return(true)
		b3.On("Address").Return(&url.URL{Host: "backend3"}).Maybe()

		lc := balancer.NewLeastConnections([]balancer.BackendServer{b1, b2, b3})

		selected, err := lc.Next()
		require.NoError(t, err)
		assert.Equal(t, b3, selected)
	})

	t.Run("returns error for empty backends", func(t *testing.T) {
		t.Parallel()

		lc := balancer.NewLeastConnections(nil)

		_, err := lc.Next()
		require.ErrorIs(t, err, balancer.ErrNoBackends)
	})

	t.Run("returns error when no healthy backend is available", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		b1.On("GetConnections").Return(int64(0))
		b1.On("Healthy").Return(false)
		b1.On("Address").Return(&url.URL{Host: "backend1"}).Maybe()

		b2 := mocks.NewBackendServer(t)
		b2.On("GetConnections").Return(int64(0))
		b2.On("Healthy").Return(false)
		b2.On("Address").Return(&url.URL{Host: "backend2"}).Maybe()

		lc := balancer.NewLeastConnections([]balancer.BackendServer{b1, b2})

		_, err := lc.Next()
		require.ErrorIs(t, err, balancer.ErrNoHealthyBackends)
	})

	t.Run("update backends array successfully", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		lc := balancer.NewLeastConnections([]balancer.BackendServer{b1})

		b2 := mocks.NewBackendServer(t)
		b2.On("GetConnections").Return(int64(0))
		b2.On("Healthy").Return(true)
		b2.On("Address").Return(&url.URL{Host: "backend2"}).Maybe()

		lc.UpdateBackends([]balancer.BackendServer{b2})

		selected, err := lc.Next()
		require.NoError(t, err)
		assert.Equal(t, b2, selected)
	})

	t.Run("try to concurrently access Next with UpdateBackends", func(t *testing.T) {
		t.Parallel()

		lc := balancer.NewLeastConnections(nil)

		var wg sync.WaitGroup

		for range 100 {
			wg.Add(2)

			go func() {
				defer wg.Done()

				_, _ = lc.Next()
			}()

			go func() {
				defer wg.Done()

				b := mocks.NewBackendServer(t)
				b.On("GetConnections").Return(int64(1)).Maybe()
				b.On("Healthy").Return(true).Maybe()
				b.On("Address").Return(&url.URL{Host: "backend1"}).Maybe()

				lc.UpdateBackends([]balancer.BackendServer{b})
			}()
		}

		wg.Wait()
	})
}
