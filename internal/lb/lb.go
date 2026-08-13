package lb

import (
	"sync/atomic"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendPool struct {
	backends []*core.Backend
	next     atomic.Uint64
}

func NewBackendPool(backends []*core.Backend) *BackendPool {
	return &BackendPool{
		backends: backends,
	}
}

func (p *BackendPool) Next() *core.Backend {
	if len(p.backends) == 0 {
		return nil
	}

	index := p.next.Add(1) - 1
	return p.backends[index%uint64(len(p.backends))]
}

type LoadBalancer struct {
	pools map[string]*BackendPool
}

func NewLoadBalancer(pools map[string]*BackendPool) *LoadBalancer {
	return &LoadBalancer{
		pools: pools,
	}
}

func (lb *LoadBalancer) Get(name string) (*BackendPool, bool) {
	pool, ok := lb.pools[name]
	return pool, ok
}

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
