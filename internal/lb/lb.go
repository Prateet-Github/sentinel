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

	strategy SelectionStrategy
}

type BackendSelection struct {
	Backend *core.Backend
	Breaker *circuitbreaker.CircuitBreaker
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	ResetTimeout     time.Duration
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 3,
		ResetTimeout:     10 * time.Second,
	}
}

func NewBackendPool(
	backends []*core.Backend,
	breakerConfig CircuitBreakerConfig,

) *BackendPool {
	states := make([]atomic.Uint32, len(backends))
	failures := make([]atomic.Uint32, len(backends))
	success := make([]atomic.Uint32, len(backends))
	breakers := make([]*circuitbreaker.CircuitBreaker, len(backends))

	for i := range states {
		states[i].Store(uint32(BackendHealthy))
		breakers[i] = circuitbreaker.New(
			breakerConfig.FailureThreshold,
			breakerConfig.ResetTimeout,
		)
	}

	return &BackendPool{
		backends: backends,
		states:   states,
		failures: failures,
		success:  success,
		breakers: breakers,
		strategy: &RoundRobinStrategy{},
	}
}

func (p *BackendPool) Next() *BackendSelection {
	if len(p.backends) == 0 {
		return nil
	}

	idx := p.strategy.Select(p)

	if idx == -1 {
		return nil
	}

	return &BackendSelection{
		Backend: p.backends[idx],
		Breaker: p.breakers[idx],
	}
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
