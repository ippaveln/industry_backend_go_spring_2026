package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	tokens     float64
	burst      int
	clock      Clock
	lastTime   time.Time
	ratePerSec float64
	mu         sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		tokens:     float64(burst),
		burst:      burst,
		clock:      clock,
		lastTime:   clock.Now(),
		ratePerSec: ratePerSec,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	delta := now.Sub(l.lastTime).Seconds()
	l.tokens += float64(delta) * l.ratePerSec
	l.tokens = min(l.tokens, float64(l.burst))
	l.lastTime = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	} else {
		return false
	}
}
