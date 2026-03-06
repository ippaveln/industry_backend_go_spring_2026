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
	tokens     float64
	lastUpdate time.Time
	mutex      sync.Mutex
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: clock.Now(),
	}
}

func (l *Limiter) Allow() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	currentTime := l.clock.Now()
	timePassed := currentTime.Sub(l.lastUpdate)
	l.lastUpdate = currentTime

	l.tokens += timePassed.Seconds() * l.ratePerSec
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
