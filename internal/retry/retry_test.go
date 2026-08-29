package retry

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.MaxAttempts != 3 {
		t.Fatalf(
			"MaxAttempts = %d, want 3",
			policy.MaxAttempts,
		)
	}

	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPut,
		http.MethodDelete,
	}

	for _, method := range methods {
		if !policy.ShouldRetryMethod(method) {
			t.Fatalf(
				"method %s should be retryable",
				method,
			)
		}
	}

	nonRetryableMethods := []string{
		http.MethodPost,
		http.MethodPatch,
	}

	for _, method := range nonRetryableMethods {
		if policy.ShouldRetryMethod(method) {
			t.Fatalf(
				"method %s should not be retryable",
				method,
			)
		}
	}

	statuses := []int{
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}

	for _, status := range statuses {
		if !policy.ShouldRetryStatus(status) {
			t.Fatalf(
				"status %d should be retryable",
				status,
			)
		}
	}
}

func TestRetryPolicyShouldRetry(t *testing.T) {
	policy := DefaultPolicy()

	tests := []struct {
		name   string
		method string
		status int
		err    error
		want   bool
	}{
		{
			name:   "GET 503",
			method: http.MethodGet,
			status: http.StatusServiceUnavailable,
			want:   true,
		},
		{
			name:   "GET 502",
			method: http.MethodGet,
			status: http.StatusBadGateway,
			want:   true,
		},
		{
			name:   "GET 504",
			method: http.MethodGet,
			status: http.StatusGatewayTimeout,
			want:   true,
		},
		{
			name:   "GET 404",
			method: http.MethodGet,
			status: http.StatusNotFound,
			want:   false,
		},
		{
			name:   "GET 500",
			method: http.MethodGet,
			status: http.StatusInternalServerError,
			want:   false,
		},
		{
			name:   "GET network error",
			method: http.MethodGet,
			err:    errors.New("connection reset"),
			want:   true,
		},
		{
			name:   "POST 503",
			method: http.MethodPost,
			status: http.StatusServiceUnavailable,
			want:   false,
		},
		{
			name:   "PATCH 503",
			method: http.MethodPatch,
			status: http.StatusServiceUnavailable,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := policy.ShouldRetry(
				tt.method,
				tt.status,
				tt.err,
			)

			if got != tt.want {
				t.Fatalf(
					"ShouldRetry() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	backoff := ExponentialBackoff{
		Base: 10 * time.Millisecond,
		Max:  100 * time.Millisecond,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{
			attempt: 1,
			want:    10 * time.Millisecond,
		},
		{
			attempt: 2,
			want:    20 * time.Millisecond,
		},
		{
			attempt: 3,
			want:    40 * time.Millisecond,
		},
		{
			attempt: 4,
			want:    80 * time.Millisecond,
		},
		{
			attempt: 5,
			want:    100 * time.Millisecond,
		},
		{
			attempt: 10,
			want:    100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(
			fmt.Sprintf("attempt_%d", tt.attempt),
			func(t *testing.T) {
				got := backoff.Delay(tt.attempt)

				if got != tt.want {
					t.Fatalf(
						"Delay(%d) = %v, want %v",
						tt.attempt,
						got,
						tt.want,
					)
				}
			},
		)
	}
}
