package lb

import "github.com/Prateet-Github/sentinel/internal/core"

func BuildLoadBalancer(cfg *core.Config) *LoadBalancer {
	pools := make(map[string]*BackendPool)

	grouped := make(map[string][]*core.Backend)

	for i := range cfg.Backends {
		backend := &cfg.Backends[i]

		grouped[backend.Name] = append(
			grouped[backend.Name],
			backend,
		)
	}

	for name, backends := range grouped {
		pools[name] = NewBackendPool(backends)
	}

	return NewLoadBalancer(pools)
}
