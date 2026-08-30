package retry

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecutorWithTransportRetries(t *testing.T) {
	attempts := 0

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++

			if attempts < 3 {
				http.Error(
					w,
					"temporary failure",
					http.StatusServiceUnavailable,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		}),
	)
	defer server.Close()

	policy := DefaultPolicy()

	backoff := ExponentialBackoff{
		Base: 0,
		Max:  0,
	}

	executor := NewExecutor(policy, backoff)

	client := &http.Client{}

	req, err := http.NewRequest(
		http.MethodGet,
		server.URL,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := executor.Execute(
		req.Method,
		func() (*http.Response, error) {
			return client.Do(req)
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if attempts != 3 {
		t.Fatalf(
			"backend attempts = %d, want 3",
			attempts,
		)
	}
}
