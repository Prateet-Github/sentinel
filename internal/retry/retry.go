package retry

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
