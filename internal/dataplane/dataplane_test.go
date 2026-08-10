package dataplane

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/proxy"
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

	backend := http.HandlerFunc(func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		params, ok := router.ParamsFromContext(req.Context())
		if !ok {
			t.Fatal("params missing from request context")
		}

		if got := params.Get("id"); got != "42" {
			t.Fatalf("id = %q, want %q", got, "42")
		}

		w.WriteHeader(http.StatusOK)
	})

	registry := proxy.NewRegistryFromHandlers(
		map[string]http.Handler{
			"users": backend,
		},
	)

	dp := New(r, registry, cfg)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/42",
		nil,
	)

	rec := httptest.NewRecorder()

	dp.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusOK,
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

	backend := http.HandlerFunc(func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	registry := proxy.NewRegistryFromHandlers(
		map[string]http.Handler{
			"users": backend,
		},
	)

	dp := New(r, registry, cfg)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		dp.ServeHTTP(rec, req)
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

	backend := http.HandlerFunc(func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	registry := proxy.NewRegistryFromHandlers(
		map[string]http.Handler{
			"users": backend,
		},
	)

	dp := New(r, registry, cfg)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/12345",
		nil,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		dp.ServeHTTP(rec, req)
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

	backend := http.HandlerFunc(func(
		w http.ResponseWriter,
		req *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	registry := proxy.NewRegistryFromHandlers(
		map[string]http.Handler{
			"users": backend,
		},
	)

	dp := New(r, registry, cfg)

	req := httptest.NewRequest(
		http.MethodGet,
		"/does-not-exist",
		nil,
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		dp.ServeHTTP(rec, req)
	}
}
