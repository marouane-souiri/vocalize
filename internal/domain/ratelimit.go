package domain

import "time"

type RateLimit struct {
	Remaining  int
	ResetAfter time.Duration
	ResetAt    time.Time
	Global     bool
}
