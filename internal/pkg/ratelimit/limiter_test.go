package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Allow_Success(t *testing.T) {
	limiter := NewLimiter(100, 10) // 100 req/sec, burst 10
	defer limiter.ResetAll()

	// Should allow first request
	if !limiter.Allow("test-key") {
		t.Error("Expected first request to be allowed")
	}
}

func TestLimiter_Allow_Multiple_Keys(t *testing.T) {
	limiter := NewLimiter(2, 1) // 2 req/sec, burst 1
	defer limiter.ResetAll()

	// Different keys should have separate limits
	key1 := "user:1"
	key2 := "user:2"

	if !limiter.Allow(key1) {
		t.Error("Expected first request for key1 to be allowed")
	}
	if !limiter.Allow(key2) {
		t.Error("Expected first request for key2 to be allowed")
	}
}

func TestLimiter_RateLimit_Exceeded(t *testing.T) {
	// Very strict rate limit for testing: 1 req/sec, burst 1
	limiter := NewLimiter(1, 1)
	defer limiter.ResetAll()

	key := "test-key"

	// First request should be allowed (within burst)
	if !limiter.Allow(key) {
		t.Error("Expected first request to be allowed")
	}

	// Second request should be denied (no more burst tokens, rate limit rate is 1/sec)
	if limiter.Allow(key) {
		t.Error("Expected second request to be denied")
	}
}

func TestLimiter_RateLimit_Recovery(t *testing.T) {
	// 10 req/sec, burst 1
	limiter := NewLimiter(10, 1)
	defer limiter.ResetAll()

	key := "test-key"

	// First request
	if !limiter.Allow(key) {
		t.Error("Expected first request to be allowed")
	}

	// Wait 150ms (enough for tokens to recover at 10/sec)
	time.Sleep(150 * time.Millisecond)

	// Should allow next request
	if !limiter.Allow(key) {
		t.Error("Expected request after recovery to be allowed")
	}
}

func TestLimiter_Reset_Key(t *testing.T) {
	limiter := NewLimiter(1, 1)
	defer limiter.ResetAll()

	key := "test-key"

	// Exhaust rate limit
	limiter.Allow(key)
	if limiter.Allow(key) {
		t.Error("Expected request to be denied")
	}

	// Reset the key
	limiter.Reset(key)

	// Should allow again
	if !limiter.Allow(key) {
		t.Error("Expected request after reset to be allowed")
	}
}

func TestLimiter_ResetAll(t *testing.T) {
	limiter := NewLimiter(1, 1)

	key1 := "key1"
	key2 := "key2"

	// Exhaust both
	limiter.Allow(key1)
	limiter.Allow(key2)

	// Reset all
	limiter.ResetAll()

	// Both should work again
	if !limiter.Allow(key1) {
		t.Error("Expected key1 to work after ResetAll")
	}
	if !limiter.Allow(key2) {
		t.Error("Expected key2 to work after ResetAll")
	}
}

func TestLimiter_BurstSize(t *testing.T) {
	// 100 req/sec, burst 5
	limiter := NewLimiter(100, 5)
	defer limiter.ResetAll()

	key := "test-key"

	// Should allow up to burst size (5 requests)
	for i := 0; i < 5; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Expected request %d to be allowed (within burst)", i+1)
		}
	}

	// 6th request should be denied (no burst tokens left, rate < 1/millisecond)
	if limiter.Allow(key) {
		t.Error("Expected 6th request to be denied (exceeded burst)")
	}
}

func TestLimiter_Concurrent_Keys(t *testing.T) {
	limiter := NewLimiter(100, 10)
	defer limiter.ResetAll()

	// Create limiters for different keys concurrently
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(keyID int) {
			key := "key:" + string(rune(keyID))
			// Each should get their own limit
			if !limiter.Allow(key) {
				t.Errorf("Expected request for key:%d to be allowed", keyID)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestLimiter_HighRate(t *testing.T) {
	// 1000 req/sec, burst 50
	limiter := NewLimiter(1000, 50)
	defer limiter.ResetAll()

	key := "test-key"

	// Should allow burst size requests
	for i := 0; i < 50; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Expected request %d to be allowed (within burst)", i+1)
		}
	}
}
