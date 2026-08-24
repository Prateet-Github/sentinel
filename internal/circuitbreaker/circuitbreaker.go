package circuitbreaker

import (
	"sync"
	"time"
)

type State uint8

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	mu sync.Mutex

	state State

	failures int

	failureThreshold int
	resetTimeout     time.Duration

	openedAt time.Time

	probeInFlight bool
}

func New(
	failureThreshold int,
	resetTimeout time.Duration,
) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		return true

	case Open:
		if time.Since(cb.openedAt) < cb.resetTimeout {
			return false
		}

		cb.state = HalfOpen
		cb.probeInFlight = true
		return true

	case HalfOpen:
		if cb.probeInFlight {
			return false
		}

		cb.probeInFlight = true
		return true
	}

	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state == HalfOpen {
		cb.state = Closed
		cb.probeInFlight = false
	}

}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case Closed:
		cb.failures++

		if cb.failures >= cb.failureThreshold {
			cb.state = Open
			cb.openedAt = time.Now()
		}

	case HalfOpen:
		cb.state = Open
		cb.openedAt = time.Now()
		cb.probeInFlight = false
	}
}
