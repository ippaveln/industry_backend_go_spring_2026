package main

import (
	"sync"
	"time"
)

type TokenBucket interface {
	Allow() (ok bool)
}

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	currTokens float64
	ratePerSec float64
	burst      int
	clock      Clock
	mutex      sync.Mutex
	lastRefill time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	limiter := &Limiter{
		currTokens: float64(burst),
		ratePerSec: ratePerSec,
		burst:      burst,
		clock:      clock,
		lastRefill: clock.Now(),
	}

	return limiter
}

func (l *Limiter) Allow() (ok bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.RefillTokenBuket()

	if l.burst < 1 || l.currTokens < 1 {
		return false
	}

	l.currTokens--

	return true
}

func (l *Limiter) RefillTokenBuket() {
	elapsedTime := l.clock.Now().Sub(l.lastRefill)
	rechargedTokens := elapsedTime.Seconds() * l.ratePerSec

	l.lastRefill = l.clock.Now()

	if rechargedTokens > float64(l.burst) {
		l.currTokens = float64(l.burst)
		return
	}
	l.currTokens += rechargedTokens
}
