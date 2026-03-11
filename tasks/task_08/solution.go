package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}
type Limiter struct {
	mu       sync.Mutex
	clock    Clock
	rate     float64
	burst    int
	tokens   float64
	lastTime time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	now := clock.Now()

	return &Limiter{
		clock:    clock,
		rate:     ratePerSec,
		burst:    burst,
		tokens:   float64(burst),
		lastTime: now,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.burst == 0 {
		return false
	}

	now := l.clock.Now()

	if !now.Before(l.lastTime) {
		timeDifference := now.Sub(l.lastTime).Seconds()

		if timeDifference > 0 {
			l.tokens += timeDifference * l.rate
		}

		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}

		l.lastTime = now
	}

	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}

	return false
}
