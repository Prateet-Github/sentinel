package dataplane

import "sync/atomic"

type RuntimeConfig struct {
	current atomic.Pointer[RuntimeState]
}

func NewRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{}
}

func (c *RuntimeConfig) Load() *RuntimeState {
	return c.current.Load()
}

func (c *RuntimeConfig) Store(state *RuntimeState) {
	c.current.Store(state)
}
