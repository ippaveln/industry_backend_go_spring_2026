package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	clock            Clock
	lastMeasuredTime time.Time
	mutex            sync.Mutex
	ratePerSec       float64
	tokens           float64
	burst            int
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:            clock,
		lastMeasuredTime: clock.Now(),
		ratePerSec:       ratePerSec,
		tokens:           float64(burst),
		burst:            burst,
	}
}

func (l *Limiter) Allow() bool {
	if l.burst <= 0 {
		return false
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	curTime := l.clock.Now()
	l.tokens += l.ratePerSec * curTime.Sub(l.lastMeasuredTime).Seconds()
	if float64(l.burst) < l.tokens {
		l.tokens = float64(l.burst)
	}

	l.lastMeasuredTime = curTime

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}
