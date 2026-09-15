package lb

import (
	"testing"

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
