package circuitbreaker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreakerClosedAllowsRequests(t *testing.T) {
	cb := New(3, time.Second)

	if !cb.Allow() {
		t.Fatal("Allow() = false, want true")
	}
}

func TestCircuitBreakerOpensAfterFailureThreshold(t *testing.T) {
	cb := New(3, time.Second)

	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.Allow() {
		t.Fatal("circuit opened before reaching threshold")
	}

	cb.RecordFailure()

	if cb.Allow() {
		t.Fatal("circuit should reject requests after threshold")
	}
}

func TestCircuitBreakerSuccessResetsFailures(t *testing.T) {
	cb := New(3, time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()

	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.Allow() {
		t.Fatal("success should reset failure count")
	}
}

func TestCircuitBreakerTransitionsToHalfOpen(t *testing.T) {
	cb := New(1, 20*time.Millisecond)

	cb.RecordFailure()

	if cb.Allow() {
		t.Fatal("open circuit should reject requests")
	}

	time.Sleep(30 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("circuit should allow request after reset timeout")
	}

	if cb.state != HalfOpen {
		t.Fatalf(
			"state = %v, want HalfOpen",
			cb.state,
		)
	}
}

func TestCircuitBreakerHalfOpenSuccessClosesCircuit(t *testing.T) {
	cb := New(1, 20*time.Millisecond)

	cb.RecordFailure()
	time.Sleep(30 * time.Millisecond)

	cb.Allow()

	cb.RecordSuccess()

	if cb.state != Closed {
		t.Fatalf(
			"state = %v, want Closed",
			cb.state,
		)
	}

	if !cb.Allow() {
		t.Fatal("closed circuit should allow requests")
	}
}

func TestCircuitBreakerHalfOpenFailureReopensCircuit(t *testing.T) {
	cb := New(1, 20*time.Millisecond)

	cb.RecordFailure()
	time.Sleep(30 * time.Millisecond)

	cb.Allow()

	cb.RecordFailure()

	if cb.state != Open {
		t.Fatalf(
			"state = %v, want Open",
			cb.state,
		)
	}

	if cb.Allow() {
		t.Fatal("reopened circuit should reject requests")
	}
}

func TestCircuitBreakerHalfOpenAllowsSingleProbe(t *testing.T) {
	cb := New(1, 20*time.Millisecond)

	// Open the circuit
	cb.RecordFailure()

	// Wait for the reset timeout
	time.Sleep(30 * time.Millisecond)

	// First request becomes the Half-Open probe
	if !cb.Allow() {
		t.Fatal("first half-open probe should be allowed")
	}

	// While the probe is in flight, other requests must be rejected
	if cb.Allow() {
		t.Fatal("second request should be rejected while probe is in flight")
	}
}

func TestCircuitBreakerHalfOpenConcurrentSingleProbe(t *testing.T) {
	cb := New(1, 20*time.Millisecond)

	// Open the circuit.
	cb.RecordFailure()

	// Wait until the circuit can transition to HalfOpen.
	time.Sleep(30 * time.Millisecond)

	const goroutines = 100

	var wg sync.WaitGroup
	var allowed atomic.Int32

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			if cb.Allow() {
				allowed.Add(1)
			}
		}()
	}

	wg.Wait()

	if got := allowed.Load(); got != 1 {
		t.Fatalf(
			"allowed requests = %d, want 1",
			got,
		)
	}
}
