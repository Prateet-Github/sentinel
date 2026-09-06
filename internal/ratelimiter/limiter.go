package ratelimiter

import "sync"

type RateLimiter struct {
	capacity   float64
	refillRate float64

	mu      sync.Mutex
	buckets map[string]*TokenBucket
}

func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*TokenBucket),
	}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()

	bucket, exists := r.buckets[key]
	if !exists {
		bucket = NewTokenBucket(r.capacity, r.refillRate)
		r.buckets[key] = bucket
	}
	r.mu.Unlock()

	return bucket.Allow()

}
