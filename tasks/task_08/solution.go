package main

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Limiter struct {
	mu sync.Mutex

	clock Clock

	ratePerSec float64
	burst      float64

	tokens   float64
	lastTime time.Time
}

// ratePerSec — скорость пополнения (токенов в секунду), может быть дробной (например 2.5).
// burst — максимальная ёмкость корзины в токенах.
func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	now := clock.Now()

	return &Limiter{
		clock:      clock,
		ratePerSec: ratePerSec,
		burst:      float64(burst),
		tokens:     float64(burst),
		lastTime:   now,
	}
}

// Allow возвращает true, если на текущий момент есть хотя бы 1 токен и он успешно списан.
// Иначе возвращает false.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	elapsed := now.Sub(l.lastTime).Seconds()

	if elapsed > 0 {
		l.tokens += elapsed * l.ratePerSec
		if l.tokens > l.burst {
			l.tokens = l.burst
		}
		l.lastTime = now
	}

	if l.tokens >= 1 {
		l.tokens -= 1
		return true
	}

	return false
}
