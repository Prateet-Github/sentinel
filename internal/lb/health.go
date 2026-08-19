package lb

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type BackendState uint8

const (
	BackendHealthy BackendState = iota
	BackendUnhealthy
)

type HealthChecker struct {
	client  *http.Client
	timeout time.Duration
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: time.Second,
		},
		timeout: time.Second,
	}
}

func (h *HealthChecker) Check(backend *core.Backend) bool {
	if backend.HealthCheckPath != "" {
		return h.checkHTTP(backend)
	}

	return h.checkTCP(backend)
}

func (h *HealthChecker) checkHTTP(backend *core.Backend) bool {
	url := backend.URL + backend.HealthCheckPath

	resp, err := h.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= http.StatusOK &&
		resp.StatusCode < http.StatusInternalServerError
}

func (h *HealthChecker) checkTCP(backend *core.Backend) bool {
	u, err := url.Parse(backend.URL)
	if err != nil {
		return false
	}

	conn, err := net.DialTimeout(
		"tcp",
		u.Host,
		h.timeout,
	)
	if err != nil {
		return false
	}

	defer conn.Close()

	return true
}
