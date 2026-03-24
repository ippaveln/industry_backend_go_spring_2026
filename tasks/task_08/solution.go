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
	last       time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      float64(burst),
		tokens:     float64(burst),
		last:       clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.burst == 0 {
		return false
	}

	now := l.clock.Now()
	elapsed := now.Sub(l.last).Seconds()
	l.last = now

	if elapsed > 0 && l.ratePerSec > 0 {
		l.tokens += elapsed * l.ratePerSec
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}

	return false
}