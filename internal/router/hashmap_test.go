package router

import (
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func BenchmarkHashMapRouter(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{Method: "GET", Path: "/users", Backend: "users-service"},
			{Method: "GET", Path: "/users/profile", Backend: "users-service"},
			{Method: "GET", Path: "/users/settings", Backend: "users-service"},
			{Method: "GET", Path: "/users/orders", Backend: "users-service"},
			{Method: "GET", Path: "/orders", Backend: "orders-service"},
			{Method: "GET", Path: "/orders/history", Backend: "orders-service"},
			{Method: "GET", Path: "/products", Backend: "products-service"},
			{Method: "GET", Path: "/products/latest", Backend: "products-service"},
			{Method: "GET", Path: "/health", Backend: "health-service"},
		},
	}

	r := NewHashMapRouter(cfg)

	paths := []string{
		"/users",
		"/users/profile",
		"/users/settings",
		"/users/orders",
		"/orders",
		"/orders/history",
		"/products",
		"/products/latest",
		"/health",
	}

	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = r.Match("GET", path)
			}
		})
	}
}
