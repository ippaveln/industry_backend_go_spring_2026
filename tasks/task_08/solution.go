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
	burst      int
	fullness   float64
	ratePerSec float64
	lastUpdate time.Time
}

// ratePerSec — скорость пополнения (токенов в секунду), может быть дробной (например 2.5).
// burst — максимальная ёмкость корзины в токенах.
func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	return &Limiter{
		clock:      clock,
		burst:      burst,
		fullness:   float64(burst),
		ratePerSec: ratePerSec,
		lastUpdate: clock.Now(),
	}
}

// Allow возвращает true, если на текущий момент есть хотя бы 1 токен и он успешно списан.
// Иначе возвращает false.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	elapsed := l.clock.Now().Sub(l.lastUpdate).Seconds()
	l.fullness += elapsed * l.ratePerSec
	l.fullness = min(l.fullness, float64(l.burst))
	l.lastUpdate = l.clock.Now()
	if l.fullness >= 1 {
		l.fullness--
		return true
	}
	return false
}
