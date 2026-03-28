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
	ratePerSec float64
	burst      int
	bucket     float64
	lastTime   time.Time
	mu         sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{clock: clock, ratePerSec: ratePerSec, burst: burst, bucket: float64(burst), lastTime: clock.Now()}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.clock.Now()
	timeDiff := now.Sub(l.lastTime)
	tokenDiff := timeDiff.Seconds() * l.ratePerSec
	l.bucket = min(l.bucket+tokenDiff, float64(l.burst))
	l.lastTime = now
	if l.bucket < 1 {
		return false
	}
	l.bucket -= 1
	return true
}
