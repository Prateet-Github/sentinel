package lb

import (
	"sync/atomic"
	"time"

	"github.com/Prateet-Github/sentinel/internal/circuitbreaker"
	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendPool struct {
	backends []*core.Backend
	states   []atomic.Uint32

	failures []atomic.Uint32
	success  []atomic.Uint32

	breakers []*circuitbreaker.CircuitBreaker

	next atomic.Uint64
}

type BackendSelection struct {
	Backend *core.Backend
	Breaker *circuitbreaker.CircuitBreaker
}

func NewBackendPool(backends []*core.Backend) *BackendPool {
	states := make([]atomic.Uint32, len(backends))
	failures := make([]atomic.Uint32, len(backends))
	success := make([]atomic.Uint32, len(backends))
	breakers := make([]*circuitbreaker.CircuitBreaker, len(backends))

	for i := range states {
		states[i].Store(uint32(BackendHealthy))
		breakers[i] = circuitbreaker.New(3, 10*time.Second)
	}

	return &BackendPool{
		backends: backends,
		states:   states,
		failures: failures,
		success:  success,
		breakers: breakers,
	}
}

func (p *BackendPool) Next() *BackendSelection {
	if len(p.backends) == 0 {
		return nil
	}

	index := p.next.Add(1) - 1

	for i := uint64(0); i < uint64(len(p.backends)); i++ {
		idx := (index + i) % uint64(len(p.backends))

		if BackendState(p.states[idx].Load()) != BackendHealthy {
			continue
		}

		if !p.breakers[idx].Allow() {
			continue
		}

		return &BackendSelection{
			Backend: p.backends[idx],
			Breaker: p.breakers[idx],
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
