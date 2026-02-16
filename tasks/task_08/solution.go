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
	ratePerSec float64
	burst      int
	clock      Clock
	last       time.Time
	tokens     float64
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		ratePerSec: ratePerSec,
		burst:      burst,
		clock:      clock,
		last:       clock.Now(),
		tokens:     float64(burst),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.refill()

	if l.tokens < 1 {
		return false
	}

	l.tokens -= 1

	return true
}

func (l *Limiter) refill() {
	now := l.clock.Now()
	elapsed := now.Sub(l.last)
	if elapsed <= 0 {
		return
	}

	l.tokens += l.ratePerSec * elapsed.Seconds()
	capacity := float64(l.burst)
	if l.tokens > capacity {
		l.tokens = capacity
	}

	l.last = now
}
