package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	mu          sync.Mutex
	clock       Clock
	lastUpdated time.Time
	tokens      float64
	ratePerSec  float64
	burst       float64
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		mu:          sync.Mutex{},
		clock:       clock,
		lastUpdated: clock.Now(),
		tokens:      float64(burst),
		ratePerSec:  ratePerSec,
		burst:       float64(burst),
	}
}

func (lim *Limiter) Allow() bool {
	if lim.burst == 0 {
		return false
	}

	lim.mu.Lock()
	defer lim.mu.Unlock()

	sec := int64(lim.clock.Now().Sub(lim.lastUpdated).Seconds())
	lim.tokens = min(lim.tokens+lim.ratePerSec*float64(sec), lim.burst)
	lim.lastUpdated = lim.lastUpdated.Add(time.Duration(sec) * time.Second)
	if lim.tokens >= 1 {
		lim.tokens--
		return true
	}
	return false
}
