package retry

import (
	"net/http"
	"testing"
	"time"
)

type testBackoff struct {
	delays []time.Duration
	calls  int
}

func (b *testBackoff) Delay(attempt int) time.Duration {
	b.calls++
	return b.delays[attempt-1]
}

func TestExecutorSucceedsWithoutRetry(t *testing.T) {
	executor := NewExecutor(
		DefaultPolicy(),
		&testBackoff{},
	)

	attempts := 0

	resp, err := executor.Execute(
		http.MethodGet,
		func() (*http.Response, error) {
			attempts++

			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if attempts != 1 {
		t.Fatalf(
			"attempts = %d, want 1",
			attempts,
		)
	}
}

func TestExecutorRetriesRetryableFailure(t *testing.T) {
	backoff := &testBackoff{
		delays: []time.Duration{
			0,
			0,
		},
	}

	executor := NewExecutor(
		DefaultPolicy(),
		backoff,
	)

	attempts := 0

	resp, err := executor.Execute(
		http.MethodGet,
		func() (*http.Response, error) {
			attempts++

			if attempts < 3 {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
				}, nil
			}

			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"attempts = %d, want 3",
			attempts,
		)
	}

	if backoff.calls != 2 {
		t.Fatalf(
			"backoff calls = %d, want 2",
			backoff.calls,
		)
	}
}

func TestExecutorStopsAfterMaxAttempts(t *testing.T) {
	backoff := &testBackoff{
		delays: []time.Duration{
			0,
			0,
		},
	}

	executor := NewExecutor(
		DefaultPolicy(),
		backoff,
	)

	attempts := 0

	resp, err := executor.Execute(
		http.MethodGet,
		func() (*http.Response, error) {
			attempts++

			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusServiceUnavailable,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"attempts = %d, want 3",
			attempts,
		)
	}
}

func TestExecutorDoesNotRetryNonRetryableMethod(t *testing.T) {
	executor := NewExecutor(
		DefaultPolicy(),
		&testBackoff{},
	)

	attempts := 0

	_, err := executor.Execute(
		http.MethodPost,
		func() (*http.Response, error) {
			attempts++

			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
			}, nil
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if attempts != 1 {
		t.Fatalf(
			"attempts = %d, want 1",
			attempts,
		)
	}
}
