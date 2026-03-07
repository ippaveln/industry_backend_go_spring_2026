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
	prevTime   time.Time
	ratePerSec float64
	tokens     float64
	capacity   int
	clock      Clock
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		prevTime:   clock.Now(),
		ratePerSec: ratePerSec,
		tokens:     float64(burst),
		capacity:   burst,
		clock:      clock,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.capacity == 0 {
		return false
	}

	currentTime := l.clock.Now()
	elapsedTime := currentTime.Sub(l.prevTime).Seconds()
	generatedTokens := elapsedTime * l.ratePerSec
	if l.tokens+generatedTokens > float64(l.capacity) {
		l.tokens = float64(l.capacity)
	} else {
		l.tokens += generatedTokens
	}

	l.prevTime = currentTime

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}
