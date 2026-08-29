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

func (p RetryPolicy) MethodAllowed(method string) bool {
	return p.RetryableMethods[method]
}
