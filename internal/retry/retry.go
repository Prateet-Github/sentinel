package retry

import (
	"net/http"
)

type RetryPolicy struct {
	MaxAttempts      int
	RetryableMethods map[string]bool
	RetryableStatus  map[int]bool
}

func NewPolicy(
	maxAttempts int,
	methods map[string]bool,
	statuses map[int]bool,
) RetryPolicy {
	return RetryPolicy{
		MaxAttempts:      maxAttempts,
		RetryableMethods: methods,
		RetryableStatus:  statuses,
	}
}

func DefaultPolicy() RetryPolicy {
	return NewPolicy(
		3,
		map[string]bool{
			http.MethodGet:     true,
			http.MethodHead:    true,
			http.MethodOptions: true,
			http.MethodPut:     true,
			http.MethodDelete:  true,
		},
		map[int]bool{
			http.StatusBadGateway:         true, // 502
			http.StatusServiceUnavailable: true, // 503
			http.StatusGatewayTimeout:     true, // 504
		},
	)
}

func (p RetryPolicy) ShouldRetryMethod(method string) bool {
	return p.RetryableMethods[method]
}

func (p RetryPolicy) ShouldRetryStatus(status int) bool {
	return p.RetryableStatus[status]
}

func (p RetryPolicy) ShouldRetry(
	method string,
	status int,
	err error,
) bool {
	if !p.ShouldRetryMethod(method) {
		return false
	}

	if err != nil {
		return true
	}

	return p.ShouldRetryStatus(status)
}
