package ratelimiter

import (
	"testing"
)

func TestRateLimiterSeparatesKeys(t *testing.T) {
	limiter := NewRateLimiter(2, 0)

	if !limiter.Allow("client-a") {
		t.Fatal("client-a first request should be allowed")
	}

	if !limiter.Allow("client-a") {
		t.Fatal("client-a second request should be allowed")
	}

	if limiter.Allow("client-a") {
		t.Fatal("client-a third request should be rejected")
	}

	if !limiter.Allow("client-b") {
		t.Fatal("client-b should have its own bucket")
	}
}
