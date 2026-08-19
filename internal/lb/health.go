package lb

import "github.com/Prateet-Github/sentinel/internal/core"

type BackendState uint8

const (
	BackendHealthy BackendState = iota
	BackendUnhealthy
)

type HealthChecker struct{}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

func (h *HealthChecker) Check(backend *core.Backend) bool {
	return false
}
