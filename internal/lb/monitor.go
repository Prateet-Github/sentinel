package lb

import (
	"context"
	"time"
)

type HealthMonitor struct {
	checker  *HealthChecker
	interval time.Duration
}

func NewHealthMonitor(
	checker *HealthChecker,
	interval time.Duration) *HealthMonitor {
	return &HealthMonitor{
		checker:  checker,
		interval: interval,
	}
}

// will check health of all backends in the pool and record the result in pool
func (m *HealthMonitor) CheckPool(pool *BackendPool) {
	for i, backend := range pool.backends {
		healthy := m.checker.Check(backend)
		pool.RecordResult(i, healthy)
	}
}

func (m *HealthMonitor) Start(
	ctx context.Context,
	pool *BackendPool,
) {
	go func() {
		m.CheckPool(pool)

		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				m.CheckPool(pool)
			}
		}
	}()
}

func (m *HealthMonitor) StartAll(
	ctx context.Context,
	lb *LoadBalancer,
) {
	for _, pool := range lb.pools {
		m.Start(ctx, pool)
	}
}
