package lb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestHealthMonitorCheckPool(t *testing.T) {
	healthyServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}),
	)
	defer unhealthyServer.Close()

	backends := []*core.Backend{
		{
			Name:            "healthy",
			URL:             healthyServer.URL,
			HealthCheckPath: "/",
		},
		{
			Name:            "unhealthy",
			URL:             unhealthyServer.URL,
			HealthCheckPath: "/",
		},
	}

	pool := NewBackendPool(backends)

	checker := NewHealthChecker()

	monitor := NewHealthMonitor(checker)

	// 1st check: failure threshold not reached yet
	monitor.CheckPool(pool)

	if got := pool.State(0); got != BackendHealthy {
		t.Fatalf("healthy backend = %v, want healthy", got)
	}

	if got := pool.State(1); got != BackendHealthy {
		t.Fatalf("unhealthy backend after 1 check = %v, want healthy", got)
	}

	// 2nd check: unhealthy backend reaches failure threshold
	monitor.CheckPool(pool)

	if got := pool.State(0); got != BackendHealthy {
		t.Fatalf("healthy backend = %v, want healthy", got)
	}

	if got := pool.State(1); got != BackendUnhealthy {
		t.Fatalf("unhealthy backend after 2 checks = %v, want unhealthy", got)
	}
}
