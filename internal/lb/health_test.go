package lb

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Prateet-Github/sentinel/internal/core"
)

func TestHealthCheckerHTTP(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{
			name:   "healthy backend",
			status: http.StatusOK,
			want:   true,
		},
		{
			name:   "server error",
			status: http.StatusInternalServerError,
			want:   false,
		},
		{
			name:   "service unavailable",
			status: http.StatusServiceUnavailable,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					if r.URL.Path != "/health" {
						t.Fatalf(
							"path = %q, want /health",
							r.URL.Path,
						)
					}

					w.WriteHeader(tt.status)
				}),
			)
			defer server.Close()

			checker := NewHealthChecker()

			backend := &core.Backend{
				URL:             server.URL,
				HealthCheckPath: "/health",
			}

			got := checker.Check(backend)

			if got != tt.want {
				t.Fatalf(
					"Check() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestHealthCheckerConnectionFailure(t *testing.T) {
	checker := NewHealthChecker()

	backend := &core.Backend{
		URL:             "http://127.0.0.1:1",
		HealthCheckPath: "/health",
	}

	if checker.Check(backend) {
		t.Fatal("Check() = true, want false")
	}
}

func TestHealthCheckerTimeout(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	checker := NewHealthChecker()

	backend := &core.Backend{
		URL:             server.URL,
		HealthCheckPath: "/health",
	}

	if checker.Check(backend) {
		t.Fatal("Check() = true, want false")
	}
}
