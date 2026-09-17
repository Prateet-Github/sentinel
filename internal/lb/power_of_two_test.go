package lb

import (
	"sync"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestPowerOfTwoEmpty(t *testing.T) {
	pool := NewBackendPool(nil, DefaultCircuitBreakerConfig())

	strategy := &PowerOfTwoStrategy{}

	got := strategy.Select(pool)

	if got != -1 {
		t.Fatalf("got %d, want -1", got)
	}
}

func TestPowerOfTwoSelectsLessLoaded(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.connections[0].Store(10)
	pool.connections[1].Store(2)

	strategy := &PowerOfTwoStrategy{}

	for i := 0; i < 100; i++ {
		got := strategy.Select(pool)

		if got != 1 {
			t.Fatalf(
				"iteration %d: got backend %d, want backend 1",
				i,
				got,
			)
		}
	}
}

func TestPowerOfTwoSkipsUnavailable(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.states[0].Store(uint32(BackendUnhealthy))

	strategy := &PowerOfTwoStrategy{}

	for i := 0; i < 100; i++ {
		got := strategy.Select(pool)

		if got != 1 {
			t.Fatalf(
				"iteration %d: got backend %d, want backend 1",
				i,
				got,
			)
		}
	}
}

func TestPowerOfTwoConcurrent(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	strategy := &PowerOfTwoStrategy{}

	const goroutines = 100
	const requestsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				got := strategy.Select(pool)

				if got < 0 || got >= len(backends) {
					t.Errorf("invalid backend index: %d", got)
				}
			}
		}()
	}

	wg.Wait()
}

func TestPowerOfTwoAllUnavailable(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	pool.states[0].Store(uint32(BackendUnhealthy))
	pool.states[1].Store(uint32(BackendUnhealthy))
	pool.states[2].Store(uint32(BackendUnhealthy))

	strategy := &PowerOfTwoStrategy{}

	got := strategy.Select(pool)

	if got != -1 {
		t.Fatalf("got %d, want -1", got)
	}
}

func BenchmarkPowerOfTwoNext(b *testing.B) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	strategy := &PowerOfTwoStrategy{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		index := strategy.Select(pool)

		if index == -1 {
			b.Fatal("Select() failed")
		}

		pool.IncrementConnections(index)
		pool.DecrementConnections(index)
	}
}

func BenchmarkPowerOfTwo3(b *testing.B) {
	benchmarkPowerOfTwo(b, 3)
}

func BenchmarkPowerOfTwo10(b *testing.B) {
	benchmarkPowerOfTwo(b, 10)
}

func BenchmarkPowerOfTwo50(b *testing.B) {
	benchmarkPowerOfTwo(b, 50)
}

func BenchmarkPowerOfTwo100(b *testing.B) {
	benchmarkPowerOfTwo(b, 100)
}

func benchmarkPowerOfTwo(b *testing.B, count int) {
	backends := make([]*core.Backend, count)

	for i := range backends {
		backends[i] = &core.Backend{
			Name: "backend",
			URL:  "http://127.0.0.1:9000",
		}
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	strategy := &PowerOfTwoStrategy{}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		index := strategy.Select(pool)

		if index == -1 {
			b.Fatal("Select() failed")
		}

		pool.IncrementConnections(index)
		pool.DecrementConnections(index)
	}
}
