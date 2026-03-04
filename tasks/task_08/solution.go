package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	mutex     sync.Mutex
	clock     Clock
	rate      float64
	burst     int
	tokens    float64
	lastCheck time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:     clock,
		rate:      ratePerSec,
		burst:     burst,
		tokens:    float64(burst),
		lastCheck: clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.burst == 0 {
		return false
	}

	now := l.clock.Now()
	elapsed := now.Sub(l.lastCheck)
	l.lastCheck = now

	tokensToAdd := l.rate * elapsed.Seconds()
	l.tokens += tokensToAdd

	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}
