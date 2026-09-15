package lb

import (
	"testing"

	"sync"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestLeastConnectionsEmpty(t *testing.T) {
	pool := NewBackendPool(nil, DefaultCircuitBreakerConfig())

	strategy := &LeastConnectionStrategy{}

	got := strategy.Select(pool)

	if got != -1 {
		t.Fatalf("got %d, want -1", got)
	}
}

func TestLeastConnectionsSelectsLowest(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.connections[0].Store(10)
	pool.connections[1].Store(3)
	pool.connections[2].Store(7)

	strategy := &LeastConnectionStrategy{}

	got := strategy.Select(pool)

	if got != 1 {
		t.Fatalf("got backend index %d, want 1", got)
	}
}

func TestLeastConnectionsSkipsUnhealthy(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.connections[0].Store(1)
	pool.connections[1].Store(10)

	pool.states[0].Store(uint32(BackendUnhealthy))

	strategy := &LeastConnectionStrategy{}

	got := strategy.Select(pool)

	if got != 1 {
		t.Fatalf("got backend index %d, want 1", got)
	}
}

func TestLeastConnectionsAllUnavailable(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.states[0].Store(uint32(BackendUnhealthy))
	pool.states[1].Store(uint32(BackendUnhealthy))

	strategy := &LeastConnectionStrategy{}

	got := strategy.Select(pool)

	if got != -1 {
		t.Fatalf("got %d, want -1", got)
	}
}

func TestBackendPoolConnections(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	const goroutines = 100
	const increments = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < increments; j++ {
				pool.IncrementConnections(0)
			}
		}()
	}

	wg.Wait()

	if got := pool.connections[0].Load(); got != 100000 {
		t.Fatalf("got %d connections, want 100000", got)
	}

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < increments; j++ {
				pool.DecrementConnections(0)
			}
		}()
	}

	wg.Wait()

	if got := pool.connections[0].Load(); got != 0 {
		t.Fatalf("got %d connections, want 0", got)
	}
}
