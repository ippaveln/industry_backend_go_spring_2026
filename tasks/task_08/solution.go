package main

import (
	"math"
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	mu sync.Mutex

	clock  Clock
	rate   float64
	burst  int64
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	if ratePerSec < 0 {
		ratePerSec = 0
	}
	if burst < 0 {
		burst = 0
	}

	now := clock.Now()

	return &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  int64(burst),
		tokens: float64(burst),
		last:   now,
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	if now.Before(l.last) {
		l.last = now
		l.tokens = math.Min(float64(l.burst), l.tokens)
		return l.tokens >= 1 && l.consume(1)
	}

	duration := now.Sub(l.last).Seconds()
	added := duration * l.rate

	l.tokens += added
	if l.tokens > float64(l.burst) {
		l.tokens = float64(l.burst)
	}

	l.last = now

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}

func (l *Limiter) consume(n float64) bool {
	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
