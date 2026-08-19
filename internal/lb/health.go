package lb

import (
	"net/http"
	"time"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendState uint8

const (
	BackendHealthy BackendState = iota
	BackendUnhealthy
)

type HealthChecker struct {
	client *http.Client
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: time.Second,
		},
	}
}

func (h *HealthChecker) Check(backend *core.Backend) bool {
	if backend.HealthCheckPath == "" {
		return false
	}

	url := backend.URL + backend.HealthCheckPath
	resp, err := h.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= http.StatusOK &&
		resp.StatusCode < http.StatusBadRequest

}
