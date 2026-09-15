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
	pool.strategy = &LeastConnectionStrategy{}

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

func BenchmarkLeastConnectionNext(b *testing.B) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())
	pool.strategy = &LeastConnectionStrategy{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		selection, ok := pool.Next()

		if !ok {
			b.Fatal("Next() failed")
		}

		pool.IncrementConnections(selection.Index)
		pool.DecrementConnections(selection.Index)
	}
}

func TestLeastConnectionsConcurrentSelection(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())
	pool.strategy = &LeastConnectionStrategy{}

	// simulate existing active requests
	pool.connections[0].Store(10)
	pool.connections[1].Store(2)
	pool.connections[2].Store(7)

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	selections := make([]int, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			selection, ok := pool.Next()
			if !ok {
				t.Error("Next() failed")
				return
			}

			selections[i] = selection.Index
		}(i)
	}

	wg.Wait()

	for i, selected := range selections {
		if selected != 1 {
			t.Fatalf(
				"request %d: selected backend %d, want backend 1",
				i,
				selected,
			)
		}
	}
}

func TestLeastConnectionsConcurrentLoad(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())
	pool.strategy = &LeastConnectionStrategy{}

	const requests = 1000

	var wg sync.WaitGroup
	wg.Add(requests)

	for i := 0; i < requests; i++ {
		go func() {
			defer wg.Done()

			selection, ok := pool.Next()
			if !ok {
				t.Error("Next() failed")
				return
			}

			pool.IncrementConnections(selection.Index)
		}()
	}

	wg.Wait()

	total := int64(0)

	for i := range backends {
		count := pool.connections[i].Load()

		t.Logf("backend-%d: %d connections", i+1, count)

		total += count
	}

	if total != requests {
		t.Fatalf("got %d total connections, want %d", total, requests)
	}
}

func TestLeastConnectionsPrefersLessLoadedBackend(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())
	pool.strategy = &LeastConnectionStrategy{}

	// simulate currently active requests
	pool.connections[0].Store(10)
	pool.connections[1].Store(2)
	pool.connections[2].Store(7)

	selection, ok := pool.Next()

	if !ok {
		t.Fatal("Next() failed")
	}

	if selection.Index != 1 {
		t.Fatalf(
			"got backend %d, want backend 1",
			selection.Index,
		)
	}
}

func BenchmarkLeastConnection3(b *testing.B) {
	benchmarkLeastConnection(b, 3)
}

func BenchmarkLeastConnection10(b *testing.B) {
	benchmarkLeastConnection(b, 10)
}

func BenchmarkLeastConnection50(b *testing.B) {
	benchmarkLeastConnection(b, 50)
}

func BenchmarkLeastConnection100(b *testing.B) {
	benchmarkLeastConnection(b, 100)
}

func benchmarkLeastConnection(b *testing.B, count int) {
	backends := make([]*core.Backend, count)

	for i := range backends {
		backends[i] = &core.Backend{
			Name: "backend",
			URL:  "http://127.0.0.1:9000",
		}
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())
	pool.strategy = &LeastConnectionStrategy{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		selection, ok := pool.Next()

		if !ok {
			b.Fatal("Next() failed")
		}

		pool.IncrementConnections(selection.Index)
		pool.DecrementConnections(selection.Index)
	}
}
