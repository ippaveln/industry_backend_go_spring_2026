package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	mu         sync.Mutex
	clock      Clock
	ratePerSec float64
	burst      int
	tokens     float64
	lastRefill time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		tokens:     float64(burst),
		lastRefill: clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.burst == 0 {
		return false
	}

	now := l.clock.Now()
	if elapsed := now.Sub(l.lastRefill).Seconds(); elapsed > 0 {
		l.tokens += elapsed * l.ratePerSec
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
	}
	l.lastRefill = now

	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}

	return false
}
