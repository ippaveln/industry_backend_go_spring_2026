package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	clock      Clock
	ratePerSec float64
	burst      int
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	elapsed := now.Sub(l.lastUpdate).Seconds()

	if elapsed < 0 {
		elapsed = 0
	}

	l.tokens = min(l.tokens+elapsed*l.ratePerSec, float64(l.burst))
	l.lastUpdate = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
