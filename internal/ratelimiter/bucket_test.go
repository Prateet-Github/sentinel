package ratelimiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucketAllowsWhenTokensAvailable(t *testing.T) {
	bucket := NewTokenBucket(10, 10)

	if !bucket.Allow() {
		t.Fatal("expected request to be allowed")
	}
}

func TestTokenBucketRejectsWhenEmpty(t *testing.T) {
	bucket := NewTokenBucket(1, 10)

	if !bucket.Allow() {
		t.Fatal("first request should be allowed")
	}

	if bucket.Allow() {
		t.Fatal("second request should be rejected")
	}
}

func TestTokenBucketRefills(t *testing.T) {
	bucket := NewTokenBucket(1, 10)

	if !bucket.Allow() {
		t.Fatal("first request should be allowed")
	}

	if bucket.Allow() {
		t.Fatal("bucket should be empty")
	}

	time.Sleep(110 * time.Millisecond)

	if !bucket.Allow() {
		t.Fatal("expected token to refill")
	}
}

func TestTokenBucketRespectsCapacity(t *testing.T) {
	bucket := NewTokenBucket(5, 10)

	time.Sleep(200 * time.Millisecond)

	allowed := 0

	for i := 0; i < 6; i++ {
		if bucket.Allow() {
			allowed++
		}
	}

	if allowed != 5 {
		t.Fatalf("expected 5 allowed requests, got %d", allowed)
	}
}

func TestTokenBucketConcurrentAccess(t *testing.T) {
	bucket := NewTokenBucket(100, 0)

	const goroutines = 1000

	var allowed atomic.Int32
	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			if bucket.Allow() {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Fatalf(
			"expected exactly 100 allowed requests, got %d",
			got,
		)
	}
}
