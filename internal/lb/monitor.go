package lb

type HealthMonitor struct {
	checker *HealthChecker
}

func NewHealthMonitor(checker *HealthChecker) *HealthMonitor {
	return &HealthMonitor{
		checker: checker,
	}
}

// will check health of all backends in the pool and record the result in pool
func (m *HealthMonitor) CheckPool(pool *BackendPool) {
	for i, backend := range pool.backends {
		healthy := m.checker.Check(backend)
		pool.RecordResult(i, healthy)
	}
}
