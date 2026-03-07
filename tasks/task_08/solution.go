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
	lastTime   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		tokens:     float64(burst),
		lastTime:   clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	passed := now.Sub(l.lastTime).Seconds()

	l.lastTime = now

	l.tokens += passed * l.ratePerSec
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}
