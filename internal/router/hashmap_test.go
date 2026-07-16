package router

import (
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func BenchmarkHashMapRouter(b *testing.B) {
	cfg := &core.Config{
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users",
				Backend: "users-service",
			},
		},
	}

	router := NewHashMapRouter(cfg)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = router.Match("GET", "/users")
	}
}
