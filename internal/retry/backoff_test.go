package retry

import (
	"testing"
	"time"
)

func TestFullJitterBackoff(t *testing.T) {
	backoff := FullJitterBackoff{
		Base: 10 * time.Millisecond,
		Max:  500 * time.Millisecond,
	}

	for attempt := 1; attempt <= 10; attempt++ {
		delay := backoff.Delay(attempt)

		exponential := backoff.Base * time.Duration(1<<(attempt-1))
		if exponential > backoff.Max {
			exponential = backoff.Max
		}

		if delay < 0 {
			t.Fatalf(
				"attempt %d: delay = %v, want >= 0",
				attempt,
				delay,
			)
		}

		if delay >= exponential {
			t.Fatalf(
				"attempt %d: delay = %v, want < %v",
				attempt,
				delay,
				exponential,
			)
		}
	}
}