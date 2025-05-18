package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter interface {
	UpdateLimit(route string, limit *RateLimit)
	IsRateLimited(route string) bool
	RetryAfter(route string) time.Duration
}

type RateLimit struct {
	Remaining  int
	ResetAfter time.Duration
	ResetAt    time.Time
	Global     bool
}

type RateLimiterImpl struct {
	mu          sync.RWMutex
	globalLimit *RateLimit
	routeLimits map[string]*RateLimit
}

func NewRateLimiter() *RateLimiterImpl {
	return &RateLimiterImpl{
		routeLimits: make(map[string]*RateLimit),
	}
}

func (r *RateLimiterImpl) UpdateLimit(route string, limit *RateLimit) {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit.ResetAt = time.Now().Add(limit.ResetAfter)

	if limit.Global {
		r.globalLimit = limit
	} else {
		r.routeLimits[route] = limit
	}
}

func (r *RateLimiterImpl) IsRateLimited(route string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()

	if r.globalLimit != nil && now.Before(r.globalLimit.ResetAt) {
		return true
	}

	if rl, ok := r.routeLimits[route]; ok && now.Before(rl.ResetAt) && rl.Remaining <= 0 {
		return true
	}

	return false
}

func (r *RateLimiterImpl) RetryAfter(route string) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()

	if r.globalLimit != nil && now.Before(r.globalLimit.ResetAt) {
		return time.Until(r.globalLimit.ResetAt)
	}

	if rl, ok := r.routeLimits[route]; ok && now.Before(rl.ResetAt) && rl.Remaining <= 0 {
		return time.Until(rl.ResetAt)
	}

	return 0
}
