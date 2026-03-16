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
	burst      int
	tokens     float64
	lastTime   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		tokens:     float64(burst),
		lastTime:   clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.burst < 1.0 {
		return false
	}
	now := l.clock.Now()
	secondsFromLastTime := now.Sub(l.lastTime).Seconds()
	if secondsFromLastTime > 0 {
		l.tokens += secondsFromLastTime * l.ratePerSec
		if l.tokens > float64(l.burst) {
			l.tokens = float64(l.burst)
		}
		l.lastTime = now
	}
	if l.tokens >= 1.0 {
		l.tokens -= 1.0
		return true
	}
	return false
}
