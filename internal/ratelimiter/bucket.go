package ratelimiter

import (
	"sync/atomic"
	"time"
)

type bucketState struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucket struct {
	capacity   float64
	refillRate float64
	state      atomic.Pointer[bucketState]
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	b := &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
	}

	b.state.Store(&bucketState{
		tokens:     capacity,
		lastRefill: time.Now(),
	})

	return b
}

func (b *TokenBucket) Allow() bool {
	for {
		old := b.state.Load()

		now := time.Now()

		elapsed := now.Sub(old.lastRefill).Seconds()

		tokens := old.tokens + elapsed*b.refillRate

		if tokens > b.capacity {
			tokens = b.capacity
		}

		if tokens < 1 {
			return false
		}

		next := &bucketState{
			tokens:     tokens - 1,
			lastRefill: now,
		}

		if b.state.CompareAndSwap(old, next) {
			return true
		}
	}
}
