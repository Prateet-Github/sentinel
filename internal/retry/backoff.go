package retry

import (
	"math/rand"
	"time"
)

type Backoff interface {
	Delay(attempt int) time.Duration
}

type ExponentialBackoff struct {
	Base time.Duration
	Max  time.Duration
}

func (b ExponentialBackoff) Delay(attempt int) time.Duration {
	delay := b.Base * time.Duration(1<<(attempt-1))

	if delay > b.Max {
		return b.Max
	}

	return delay
}

type FullJitterBackoff struct {
	Base time.Duration
	Max  time.Duration
}

func (b FullJitterBackoff) Delay(attempt int) time.Duration {
	exponential := b.Base * time.Duration(1<<(attempt-1))

	if exponential > b.Max {
		exponential = b.Max
	}

	if exponential <= 0 {
		return 0
	}

	return time.Duration(rand.Int63n(int64(exponential)))
}
