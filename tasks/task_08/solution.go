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
	burst      float64
	tokens     float64
	lastRefill time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	b := burst
	if b < 0 {
		b = 0
	}

	now := clock.Now()
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      float64(b),
		tokens:     float64(b),
		lastRefill: now,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	l.refill(now)

	if l.tokens < 1 {
		return false
	}

	l.tokens -= 1
	return true
}

func (l *Limiter) refill(now time.Time) {
	if l.burst <= 0 {
		l.tokens = 0
		l.lastRefill = now
		return
	}

	if now.Before(l.lastRefill) {
		l.lastRefill = now
		return
	}

	if l.ratePerSec <= 0 {
		l.lastRefill = now
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
		return
	}

	elapsed := now.Sub(l.lastRefill).Seconds()
	if elapsed > 0 {
		l.tokens += elapsed * l.ratePerSec
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
	}

	l.lastRefill = now
}
