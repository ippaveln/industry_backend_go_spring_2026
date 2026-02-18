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
	tokens     float64
	clock      Clock
	ratePerSec float64
	burst      int
	last       time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	now := clock.Now()
	tokens := float64(burst)

	return &Limiter{
		tokens:     tokens,
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		last:       now,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.burst == 0 {
		return false
	}

	now := l.clock.Now()
	diff := now.Sub(l.last).Seconds()
	if diff > 0 {
		t := diff * l.ratePerSec
		l.tokens += t
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}

	}
	l.last = now

	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}
