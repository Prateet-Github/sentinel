package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prateet-Github/sentinel/internal/retry"
)

func TestProxyRetryIntegration(t *testing.T) {
	backendRequests := 0

	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			backendRequests++

			if backendRequests < 3 {
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
	defer backend.Close()

	p, err := New(backend.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	executor := retry.NewExecutor(
		retry.DefaultPolicy(),
		retry.ExponentialBackoff{
			Base: 0,
			Max:  0,
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	resp, err := executor.Execute(
		req.Method,
		func() (*http.Response, error) {
			return p.Attempt(req)
		},
	)

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Execute() response = nil")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	if backendRequests != 3 {
		t.Fatalf(
			"backend requests = %d, want 3",
			backendRequests,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "success" {
		t.Fatalf(
			"body = %q, want %q",
			string(body),
			"success",
		)
	}

}
