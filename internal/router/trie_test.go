package router

import (
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func BenchmarkTrieRouter(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users",
				Backend: "users-service",
			},
		},
	}

	r := NewTrieRouter(cfg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.Match("GET", "/users")
	}
}
