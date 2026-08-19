package lb

type BackendState uint8

const (
	BackendHealthy BackendState = iota
	BackendUnhealthy
)
