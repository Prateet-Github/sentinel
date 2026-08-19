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

func TestLoadBalancerGet(t *testing.T) {
	users := NewBackendPool([]*core.Backend{
		{URL: "http://localhost:9001"},
		{URL: "http://localhost:9002"},
	})

	orders := NewBackendPool([]*core.Backend{
		{URL: "http://localhost:9101"},
	})

	lb := NewLoadBalancer(map[string]*BackendPool{
		"users-service":  users,
		"orders-service": orders,
	})

	tests := []struct {
		name string
		want bool
	}{
		{
			name: "users-service",
			want: true,
		},
		{
			name: "orders-service",
			want: true,
		},
		{
			name: "unknown-service",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, ok := lb.Get(tt.name)

			if ok != tt.want {
				t.Fatalf("Get(%q) ok = %v, want %v", tt.name, ok, tt.want)
			}

			if tt.want && pool == nil {
				t.Fatalf("Get(%q) returned nil pool", tt.name)
			}
		})
	}
}

func TestLoadBalancerEmpty(t *testing.T) {
	lb := NewLoadBalancer(
		map[string]*BackendPool{},
	)

	pool, ok := lb.Get("users-service")

	if ok {
		t.Fatal("expected lookup to fail")
	}

	if pool != nil {
		t.Fatal("expected nil pool")
	}
}

func TestBuildLoadBalancer(t *testing.T) {
	cfg := &core.Config{
		Backends: []core.Backend{
			{Name: "users-service", URL: "http://127.0.0.1:9000"},
			{Name: "users-service", URL: "http://127.0.0.1:9001"},
			{Name: "users-service", URL: "http://127.0.0.1:9002"},
			{Name: "orders-service", URL: "http://127.0.0.1:9100"},
		},
	}

	lb := BuildLoadBalancer(cfg)

	users, ok := lb.Get("users-service")
	if !ok {
		t.Fatal("users-service pool not found")
	}

	if len(users.backends) != 3 {
		t.Fatalf(
			"users-service backend count = %d, want 3",
			len(users.backends),
		)
	}

	orders, ok := lb.Get("orders-service")
	if !ok {
		t.Fatal("orders-service pool not found")
	}

	if len(orders.backends) != 1 {
		t.Fatalf(
			"orders-service backend count = %d, want 1",
			len(orders.backends),
		)
	}
}

func TestBackendPoolHealth(t *testing.T) {
	backends := []*core.Backend{
		{Name: "backend-1", URL: "http://127.0.0.1:9001"},
		{Name: "backend-2", URL: "http://127.0.0.1:9002"},
		{Name: "backend-3", URL: "http://127.0.0.1:9003"},
	}

	pool := NewBackendPool(backends)

	if got := pool.State(0); got != BackendHealthy {
		t.Fatalf("backend-1 state = %v, want healthy", got)
	}

	pool.SetState(1, BackendUnhealthy)

	if got := pool.State(1); got != BackendUnhealthy {
		t.Fatalf("backend-2 state = %v, want unhealthy", got)
	}

	if got := pool.State(0); got != BackendHealthy {
		t.Fatalf("backend-1 state = %v, want healthy", got)
	}
}
