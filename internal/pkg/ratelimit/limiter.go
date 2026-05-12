package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      float64
	burst    int
	cleanup  time.Duration
}

func NewLimiter(rps float64, burst int) *Limiter {
	l := &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	go l.cleanupLimiters()

	return l
}
func (l *Limiter) Allow(key string) bool {
	limiter := l.getLimiter(key)
	return limiter.Allow()
}

func (l *Limiter) getLimiter(key string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, exists := l.limiters[key]; exists {
		return limiter
	}
	limiter = rate.NewLimiter(rate.Limit(l.rps), l.burst)
	l.limiters[key] = limiter

	return limiter
}

func (l *Limiter) cleanupLimiters() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		for key, limiter := range l.limiters {
			if !limiter.AllowN(time.Now(), 0) {
				continue
			}
			if limiter.Tokens() == float64(l.burst) {
				delete(l.limiters, key)
			}
		}
		l.mu.Unlock()
	}
}

func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.limiters, key)
}

func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limiters = make(map[string]*rate.Limiter)
}
