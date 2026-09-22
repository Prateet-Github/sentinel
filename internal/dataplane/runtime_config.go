package dataplane

import (
	"sync/atomic"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

type RuntimeConfig struct {
	current atomic.Pointer[controlv1.ConfigSnapshot]
}

func NewRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{}
}

func (c *RuntimeConfig) Load() *controlv1.ConfigSnapshot {
	return c.current.Load()
}

func (c *RuntimeConfig) Store(
	snapshot *controlv1.ConfigSnapshot,
) {
	c.current.Store(snapshot)
}
