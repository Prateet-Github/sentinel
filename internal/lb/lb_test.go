package lb

import (
	"sync"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestBackendPoolRoundRobin(t *testing.T) {
	backends := []*core.Backend{
		{
			Name: "backend-1",
			URL:  "http://127.0.0.1:9001",
		},
		{
			Name: "backend-2",
			URL:  "http://127.0.0.1:9002",
		},
		{
			Name: "backend-3",
			URL:  "http://127.0.0.1:9003",
		},
	}

	pool := NewBackendPool(backends)

	want := []string{
		"backend-1",
		"backend-2",
		"backend-3",
		"backend-1",
		"backend-2",
		"backend-3",
	}

	for i, expected := range want {
		got := pool.Next()

		if got == nil {
			t.Fatalf("Next() returned nil at iteration %d", i)
		}

		if got.Name != expected {
			t.Fatalf(
				"iteration %d: got %q, want %q",
				i,
				got.Name,
				expected,
			)
		}
	}
}

func TestBackendPoolEmpty(t *testing.T) {
	pool := NewBackendPool(nil)

	if got := pool.Next(); got != nil {
		t.Fatalf("Next() = %v, want nil", got)
	}
}

func TestBackendPoolConcurrent(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends)

	const goroutines = 100
	const requestsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				if backend := pool.Next(); backend == nil {
					t.Error("Next() returned nil")
				}
			}
		}()
	}

	wg.Wait()
}

func BenchmarkBackendPoolNext(b *testing.B) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.Next()
	}
}
