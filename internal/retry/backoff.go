package retry

import (
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
