package lb

import (
	"github.com/Prateet-Github/sentinel/internal/core"
)

func BuildLoadBalancer(cfg *core.Config) *LoadBalancer {
	pools := make(map[string]*BackendPool)

	grouped := make(map[string][]*core.Backend)

	for i := range cfg.Backends {
		backend := &cfg.Backends[i]

		grouped[backend.Service] = append(
			grouped[backend.Service],
			backend,
		)
	}

	for name, backends := range grouped {
		pools[name] = NewBackendPool(
			backends,
			DefaultCircuitBreakerConfig(),
		)
	}

	return NewLoadBalancer(pools)
}
