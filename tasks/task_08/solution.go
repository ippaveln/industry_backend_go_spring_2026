package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	clock          Clock
	tokens         float64
	mu             sync.Mutex
	lastRefillTime time.Time
	ratePerSec     float64
	burst          float64
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:          clock,
		tokens:         float64(burst),
		lastRefillTime: clock.Now(),
		ratePerSec:     ratePerSec,
		burst:          float64(burst),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock.Now()
	elapsed := now.Sub(l.lastRefillTime).Seconds()
	l.tokens = min(elapsed*l.ratePerSec+l.tokens, l.burst)
	l.lastRefillTime = now
	if l.tokens >= 1 {
		l.tokens = max(l.tokens-1, 0)
		return true
	}
	return false
}
