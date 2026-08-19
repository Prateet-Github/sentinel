package lb

import (
	"sync/atomic"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendPool struct {
	backends []*core.Backend
	states   []atomic.Uint32
	next     atomic.Uint64
}

func NewBackendPool(backends []*core.Backend) *BackendPool {
	states := make([]atomic.Uint32, len(backends))

	for i := range states {
		states[i].Store(uint32(BackendHealthy))
	}

	return &BackendPool{
		backends: backends,
		states:   states,
	}
}

func (p *BackendPool) Next() *core.Backend {
	if len(p.backends) == 0 {
		return nil
	}

	index := p.next.Add(1) - 1

	for i := uint64(0); i < uint64(len(p.backends)); i++ {
		idx := (index + i) % uint64(len(p.backends))

		if BackendState(p.states[idx].Load()) == BackendHealthy {
			return p.backends[idx]
		}
	}

	return nil
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

func (p *BackendPool) State(index int) BackendState {
	return BackendState(p.states[index].Load())
}

func (p *BackendPool) SetState(index int, state BackendState) {
	p.states[index].Store(uint32(state))
}
