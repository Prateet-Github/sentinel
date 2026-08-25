package dataplane

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/lb"
	"github.com/Prateet-Github/sentinel/internal/router"
)

func TestDataplaneParameterPropagation(t *testing.T) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users/:id",
				Backend: "users",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	// backend := &core.Backend{
	// 	Name: "users",
	// 	URL:  "http://localhost:9000",
	// }

	// pool := lb.NewBackendPool([]*core.Backend{
	// 	backend,
	// })

	// loadBalancer := lb.NewLoadBalancer(
	// 	map[string]*lb.BackendPool{
	// 		"users": pool,
	// 	},
	// )

	// dp := New(r, loadBalancer, cfg)

	var params router.Params

	route, ok := r.Match(
		http.MethodGet,
		"/users/42",
		&params,
	)

	if !ok {
		t.Fatal("expected route to match")
	}

	if route.Backend != "users" {
		t.Fatalf(
			"backend = %q, want %q",
			route.Backend,
			"users",
		)
	}

	if got := params.Get("id"); got != "42" {
		t.Fatalf(
			"param id = %q, want %q",
			got,
			"42",
		)
	}
}

func BenchmarkDataplaneStatic(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users",
				Backend: "users",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	backends := []*core.Backend{
		{
			Name: "users",
			URL:  "http://localhost:9000",
		},
	}

	pool := lb.NewBackendPool(backends)

	loadBalancer := lb.NewLoadBalancer(
		map[string]*lb.BackendPool{
			"users": pool,
		},
	)

	dp := New(r, loadBalancer, cfg)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	params := router.Params{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route, ok := r.Match(
			req.Method,
			req.URL.Path,
			&params,
		)

		if !ok {
			b.Fatal("route did not match")
		}

		_, ok = loadBalancer.Get(route.Backend)
		if !ok {
			b.Fatal("backend pool not found")
		}

		_ = dp
	}
}

func BenchmarkDataplaneParameter(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users/:id",
				Backend: "users",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	pool := lb.NewBackendPool([]*core.Backend{
		{
			Name: "users",
			URL:  "http://localhost:9000",
		},
	})

	loadBalancer := lb.NewLoadBalancer(
		map[string]*lb.BackendPool{
			"users": pool,
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/12345",
		nil,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var params router.Params

		route, ok := r.Match(
			req.Method,
			req.URL.Path,
			&params,
		)

		if !ok {
			b.Fatal("route did not match")
		}

		pool, ok := loadBalancer.Get(route.Backend)
		if !ok {
			b.Fatal("backend pool not found")
		}

		_ = pool.Next()
	}
}

func BenchmarkDataplaneMiss(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users",
				Backend: "users",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var params router.Params

		_, ok := r.Match(
			http.MethodGet,
			"/does-not-exist",
			&params,
		)

		if ok {
			b.Fatal("expected route miss")
		}
	}
}

func TestDataplaneLoadBalancing(t *testing.T) {
	backends := []struct {
		name     string
		response string
	}{
		{
			name:     "backend-1",
			response: "backend-1",
		},
		{
			name:     "backend-2",
			response: "backend-2",
		},
		{
			name:     "backend-3",
			response: "backend-3",
		},
	}

	servers := make([]*httptest.Server, 0, len(backends))

	for _, backend := range backends {
		response := backend.response

		server := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, response)
			}),
		)

		servers = append(servers, server)
	}

	defer func() {
		for _, server := range servers {
			server.Close()
		}
	}()

	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  http.MethodGet,
				Path:    "/users",
				Backend: "users-service",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	poolBackends := make([]*core.Backend, 0, len(servers))

	for i, server := range servers {
		poolBackends = append(poolBackends, &core.Backend{
			Name: "users-service",
			URL:  server.URL,
		})

		_ = i
	}

	pool := lb.NewBackendPool(poolBackends)

	loadBalancer := lb.NewLoadBalancer(
		map[string]*lb.BackendPool{
			"users-service": pool,
		},
	)

	dp := New(r, loadBalancer, cfg)

	want := []string{
		"backend-1",
		"backend-2",
		"backend-3",
		"backend-1",
		"backend-2",
		"backend-3",
	}

	for i, expected := range want {
		req := httptest.NewRequest(
			http.MethodGet,
			"/users",
			nil,
		)

		rec := httptest.NewRecorder()

		dp.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf(
				"request %d: status = %d, want %d",
				i,
				rec.Code,
				http.StatusOK,
			)
		}

		got := strings.TrimSpace(rec.Body.String())

		if got != expected {
			t.Fatalf(
				"request %d: response = %q, want %q",
				i,
				got,
				expected,
			)
		}
	}
}

func TestCircuitBreakerIntegration(t *testing.T) {
	backendRequests := 0

	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			backendRequests++

			http.Error(
				w,
				"backend failure",
				http.StatusInternalServerError,
			)
		}),
	)
	defer backend.Close()

	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  http.MethodGet,
				Path:    "/users",
				Backend: "users-service",
			},
		},
	}

	r := router.NewRadixRouter(cfg)

	pool := lb.NewBackendPool([]*core.Backend{
		{
			Name: "users-service",
			URL:  backend.URL,
		},
	})

	loadBalancer := lb.NewLoadBalancer(
		map[string]*lb.BackendPool{
			"users-service": pool,
		},
	)

	dp := New(r, loadBalancer, cfg)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(
			http.MethodGet,
			"/users",
			nil,
		)

		rec := httptest.NewRecorder()

		dp.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf(
				"request %d: status = %d, want %d",
				i+1,
				rec.Code,
				http.StatusInternalServerError,
			)
		}
	}

	if backendRequests != 3 {
		t.Fatalf(
			"backend requests = %d, want 3",
			backendRequests,
		)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	rec := httptest.NewRecorder()

	dp.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf(
			"request 4: status = %d, want %d",
			rec.Code,
			http.StatusBadGateway,
		)
	}

	if backendRequests != 3 {
		t.Fatalf(
			"backend was called after circuit opened: got %d requests, want 3",
			backendRequests,
		)
	}
}
