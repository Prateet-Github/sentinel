package ratelimiter

import (
	"fmt"
	"sync"
	"sync/atomic"
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

func TestRateLimiterConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(100, 0)

	const (
		goroutines = 1000
		keys       = 10
	)

	var wg sync.WaitGroup
	var allowed atomic.Int32

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("client-%d", i%keys)

			if limiter.Allow(key) {
				allowed.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if got := allowed.Load(); got != 1000 {
		t.Fatalf("expected 1000 allowed requests, got %d", got)
	}
}

func BenchmarkRateLimiterAllow(b *testing.B) {
	limiter := NewRateLimiter(1_000_000, 1_000_000)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.Allow("client-1")
	}
}

func BenchmarkRateLimiterParallel(b *testing.B) {
	limiter := NewRateLimiter(1_000_000, 1_000_000)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			limiter.Allow("client-1")
		}
	})
}
