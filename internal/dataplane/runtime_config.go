package dataplane

import (
	"context"
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

	c.once.Do(func() {
		close(c.ready)
	})
}

func (c *RuntimeConfig) WaitReady(ctx context.Context) error {
	select {
	case <-c.ready:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
