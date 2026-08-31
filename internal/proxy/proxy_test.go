package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// for successful response
func TestProxyAttemptSuccess(t *testing.T) {
	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("hello"))
		}),
	)
	defer backend.Close()

	p, err := New(backend.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/hello",
		nil,
	)

	resp, err := p.Attempt(req)
	if err != nil {
		t.Fatalf("Attempt() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Attempt() response = nil")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(body) != "hello" {
		t.Fatalf(
			"body = %q, want %q",
			string(body),
			"hello",
		)
	}
}

// 503 remains 503
func TestProxyAttemptReturnsUpstreamErrorStatus(t *testing.T) {
	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(
				w,
				"temporary failure",
				http.StatusServiceUnavailable,
			)
		}),
	)
	defer backend.Close()

	p, err := New(backend.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	resp, err := p.Attempt(req)
	if err != nil {
		t.Fatalf("Attempt() error = %v", err)
	}

	if resp == nil {
		t.Fatal("Attempt() response = nil")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want %d",
			resp.StatusCode,
			http.StatusServiceUnavailable,
		)
	}
}

// for network errors
func TestProxyAttemptNetworkError(t *testing.T) {
	backend := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	)

	backendURL := backend.URL
	backend.Close()

	p, err := New(backendURL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/users",
		nil,
	)

	resp, err := p.Attempt(req)

	if err == nil {
		t.Fatal("Attempt() error = nil, want error")
	}

	if resp != nil {
		t.Fatal("Attempt() response != nil, want nil")
	}
}
