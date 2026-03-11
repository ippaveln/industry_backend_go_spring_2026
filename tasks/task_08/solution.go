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
	clock      Clock
	ratePerSec float64
	burst      int
	lastUpdate time.Time
	mu         sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{tokens: float64(burst), clock: clock, ratePerSec: ratePerSec, burst: burst, lastUpdate: clock.Now()}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	elapsed := now.Sub(l.lastUpdate).Seconds()
	l.tokens = min(float64(l.burst), l.tokens+(elapsed*l.ratePerSec))
	l.lastUpdate = now

	if l.tokens >= 1 {
		l.tokens--
		return true
	}

	return false
}
