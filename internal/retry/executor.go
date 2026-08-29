package retry

import (
	"net/http"
	"time"
)

type AttemptFunc func() (*http.Response, error)

type Executor struct {
	Policy  RetryPolicy
	Backoff Backoff
}

func NewExecutor(policy RetryPolicy, backoff Backoff) *Executor {
	return &Executor{
		Policy:  policy,
		Backoff: backoff,
	}
}

func (e *Executor) Execute(
	method string,
	attempt AttemptFunc,
) (*http.Response, error) {
	for i := 1; i <= e.Policy.MaxAttempts; i++ {
		resp, err := attempt()

		status := 0
		if resp != nil {
			status = resp.StatusCode
		}

		if !e.Policy.ShouldRetry(method, status, err) {
			return resp, err
		}

		if i == e.Policy.MaxAttempts {
			return resp, err
		}

		if err == nil && resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		time.Sleep(e.Backoff.Delay(i))
	}

	return nil, nil
}
