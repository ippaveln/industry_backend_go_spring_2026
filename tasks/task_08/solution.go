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
	lastReFill time.Time
	ratePerSec float64
	tokens     float64
	burst      int
	mu         sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		tokens:     float64(burst),
		burst:      burst,
		lastReFill: clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.reFill()

	if l.burst != 0 && l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}

func (l *Limiter) reFill() {
	now := l.clock.Now()
	elapsed := now.Sub(l.lastReFill).Seconds()

	l.lastReFill = now
	l.tokens += elapsed * l.ratePerSec

	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}
}
