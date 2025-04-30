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

func TestRandom(t *testing.T) {
	t.Parallel()

	t.Run("get the healthy backend", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		b1.On("Healthy").Return(true).Maybe()
		b1.On("Address").Return(&url.URL{Host: "backend1"}).Maybe()

		b2 := mocks.NewBackendServer(t)
		b2.On("Healthy").Return(true).Maybe()
		b2.On("Address").Return(&url.URL{Host: "backend2"}).Maybe()

		b3 := mocks.NewBackendServer(t)
		b3.On("Healthy").Return(true).Maybe()
		b3.On("Address").Return(&url.URL{Host: "backend3"}).Maybe()

		b4 := mocks.NewBackendServer(t)
		b4.On("Healthy").Return(false).Maybe()
		b4.On("Address").Return(&url.URL{Host: "backend4"}).Maybe()

		random := balancer.NewRandom([]balancer.BackendServer{b1, b2, b3, b4})

		selected, err := random.Next()
		require.NoError(t, err)
		assert.Contains(t, []balancer.BackendServer{b1, b2, b3}, selected)
	})

	t.Run("return error when no backends are set", func(t *testing.T) {
		t.Parallel()

		random := balancer.NewRandom(nil)
		_, err := random.Next()
		require.ErrorIs(t, err, balancer.ErrNoBackends)
	})

	t.Run("return error when no healthy backends are available", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		b1.On("Healthy").Return(false)
		b1.On("Address").Return(&url.URL{Host: "backend1"}).Maybe()

		b2 := mocks.NewBackendServer(t)
		b2.On("Healthy").Return(false)
		b2.On("Address").Return(&url.URL{Host: "backend2"}).Maybe()

		random := balancer.NewRandom([]balancer.BackendServer{b1, b2})
		_, err := random.Next()
		require.ErrorIs(t, err, balancer.ErrNoHealthyBackends)
	})

	t.Run("update backends successfully", func(t *testing.T) {
		t.Parallel()

		b1 := mocks.NewBackendServer(t)
		random := balancer.NewRandom([]balancer.BackendServer{b1})

		b2 := mocks.NewBackendServer(t)
		b2.On("Healthy").Return(true)
		b2.On("Address").Return(&url.URL{Host: "backend2"})

		random.UpdateBackends([]balancer.BackendServer{b2})

		selected, err := random.Next()
		require.NoError(t, err)
		assert.Equal(t, b2, selected)
	})

	t.Run("concurrent access to Next and UpdateBackends", func(t *testing.T) {
		t.Parallel()

		random := balancer.NewRandom(nil)

		var wg sync.WaitGroup

		for range 100 {
			wg.Add(2)

			go func() {
				defer wg.Done()

				_, _ = random.Next()
			}()

			go func() {
				defer wg.Done()

				b := mocks.NewBackendServer(t)
				b.On("Healthy").Return(true).Maybe()
				b.On("Address").Return(&url.URL{Host: "backend"}).Maybe()

				random.UpdateBackends([]balancer.BackendServer{b})
			}()
		}

		wg.Wait()
	})
}
