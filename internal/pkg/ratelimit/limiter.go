package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter manages rate limits for different keys (IPs, user IDs, etc).
type Limiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      float64        // requests per second
	burst    int            // burst size
	cleanup  time.Duration  // cleanup interval for old limiters
}

// NewLimiter creates a new rate limiter with specified requests per second and burst size.
func NewLimiter(rps float64, burst int) *Limiter {
	l := &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go l.cleanupLimiters()

	return l
}

// Allow checks if a request from the given key is allowed.
func (l *Limiter) Allow(key string) bool {
	limiter := l.getLimiter(key)
	return limiter.Allow()
}

// getLimiter retrieves or creates a limiter for the given key.
func (l *Limiter) getLimiter(key string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if exists {
		return limiter
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Double-check locking
	if limiter, exists := l.limiters[key]; exists {
		return limiter
	}

	// Create new limiter for this key
	limiter = rate.NewLimiter(rate.Limit(l.rps), l.burst)
	l.limiters[key] = limiter

	return limiter
}

// cleanupLimiters periodically removes old unused limiters to prevent memory leaks.
func (l *Limiter) cleanupLimiters() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		// Remove limiters that haven't been used recently (AllowN returns 0 tokens)
		for key, limiter := range l.limiters {
			// Check if limiter has available tokens
			if !limiter.AllowN(time.Now(), 0) {
				// If it doesn't allow 0 tokens, it hasn't been reset, so keep it
				continue
			}
			// Remove unused limiters
			if limiter.Tokens() == float64(l.burst) {
				delete(l.limiters, key)
			}
		}
		l.mu.Unlock()
	}
}

// Reset resets the limiter for a given key.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.limiters, key)
}

// ResetAll resets all limiters.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.limiters = make(map[string]*rate.Limiter)
}
