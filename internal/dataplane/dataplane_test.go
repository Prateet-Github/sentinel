package dataplane

import (
	"net/http"
	"net/http/httptest"
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
