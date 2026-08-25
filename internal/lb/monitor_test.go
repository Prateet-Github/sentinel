package lb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	checker := NewHealthChecker()

	monitor := NewHealthMonitor(checker,
		5*time.Second)

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

func TestHealthMonitorStart(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}),
	)
	defer server.Close()

	backends := []*core.Backend{
		{
			Name:            "backend-1",
			URL:             server.URL,
			HealthCheckPath: "/",
		},
	}

	pool := NewBackendPool(backends, DefaultCircuitBreakerConfig())

	monitor := NewHealthMonitor(
		NewHealthChecker(),
		10*time.Millisecond,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor.Start(ctx, pool)

	// Wait long enough for at least two health checks
	time.Sleep(100 * time.Millisecond)

	if got := pool.State(0); got != BackendUnhealthy {
		t.Fatalf(
			"backend state = %v, want unhealthy",
			got,
		)
	}
}
