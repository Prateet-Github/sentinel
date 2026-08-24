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
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = HalfOpen
			return true
		}

		return false

	case HalfOpen:
		return true
	}

	return false
}
