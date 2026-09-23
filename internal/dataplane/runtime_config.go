package dataplane

import (
	"sync"
	"sync/atomic"
)

type RuntimeConfig struct {
	current atomic.Pointer[RuntimeState]

	ready chan struct{}
	once  sync.Once
}

func NewRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		ready: make(chan struct{}),
	}
}

func (c *RuntimeConfig) Load() *RuntimeState {
	return c.current.Load()
}

func (c *RuntimeConfig) Store(state *RuntimeState) {
	c.current.Store(state)

	// mark the runtime ready exactly once
	c.once.Do(func() {
		close(c.ready)
	})
}

func (c *RuntimeConfig) WaitReady() {
	<-c.ready
}
