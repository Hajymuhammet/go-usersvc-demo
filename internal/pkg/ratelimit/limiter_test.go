package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Allow_Success(t *testing.T) {
	limiter := NewLimiter(100, 10)
	defer limiter.ResetAll()

	if !limiter.Allow("test-key") {
		t.Error("Expected first request to be allowed")
	}
}

func TestLimiter_Allow_Multiple_Keys(t *testing.T) {
	limiter := NewLimiter(2, 1)
	defer limiter.ResetAll()

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
	limiter := NewLimiter(1, 1)
	defer limiter.ResetAll()

	key := "test-key"

	if !limiter.Allow(key) {
		t.Error("Expected first request to be allowed")
	}

	if limiter.Allow(key) {
		t.Error("Expected second request to be denied")
	}
}

func TestLimiter_RateLimit_Recovery(t *testing.T) {
	limiter := NewLimiter(10, 1)
	defer limiter.ResetAll()

	key := "test-key"
	if !limiter.Allow(key) {
		t.Error("Expected first request to be allowed")
	}

	time.Sleep(150 * time.Millisecond)

	if !limiter.Allow(key) {
		t.Error("Expected request after recovery to be allowed")
	}
}

func TestLimiter_Reset_Key(t *testing.T) {
	limiter := NewLimiter(1, 1)
	defer limiter.ResetAll()

	key := "test-key"

	limiter.Allow(key)
	if limiter.Allow(key) {
		t.Error("Expected request to be denied")
	}
	limiter.Reset(key)

	if !limiter.Allow(key) {
		t.Error("Expected request after reset to be allowed")
	}
}

func TestLimiter_ResetAll(t *testing.T) {
	limiter := NewLimiter(1, 1)

	key1 := "key1"
	key2 := "key2"

	limiter.Allow(key1)
	limiter.Allow(key2)

	limiter.ResetAll()
	if !limiter.Allow(key1) {
		t.Error("Expected key1 to work after ResetAll")
	}
	if !limiter.Allow(key2) {
		t.Error("Expected key2 to work after ResetAll")
	}
}

func TestLimiter_BurstSize(t *testing.T) {
	limiter := NewLimiter(100, 5)
	defer limiter.ResetAll()

	key := "test-key"
	for i := 0; i < 5; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Expected request %d to be allowed (within burst)", i+1)
		}
	}

	if limiter.Allow(key) {
		t.Error("Expected 6th request to be denied (exceeded burst)")
	}
}

func TestLimiter_Concurrent_Keys(t *testing.T) {
	limiter := NewLimiter(100, 10)
	defer limiter.ResetAll()

	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(keyID int) {
			key := "key:" + string(rune(keyID))
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
	limiter := NewLimiter(1000, 50)
	defer limiter.ResetAll()

	key := "test-key"
	for i := 0; i < 50; i++ {
		if !limiter.Allow(key) {
			t.Errorf("Expected request %d to be allowed (within burst)", i+1)
		}
	}
}
