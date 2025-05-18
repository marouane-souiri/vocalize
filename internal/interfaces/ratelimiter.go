package interfaces

import (
	"time"

	"github.com/marouane-souiri/vocalize/internal/domain"
)

type RateLimiter interface {
	UpdateLimit(route string, limit *domain.RateLimit)
	IsRateLimited(route string) bool
	RetryAfter(route string) time.Duration
}
