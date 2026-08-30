package security

import (
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	if !limiter.Allow("user-1") {
		t.Fatal("first request should be allowed")
	}

	if !limiter.Allow("user-1") {
		t.Fatal("second request should be allowed")
	}

	if limiter.Allow("user-1") {
		t.Fatal("third request should be blocked")
	}
}

func TestRateLimiterSeparatesKeys(t *testing.T) {
	limiter := NewRateLimiter(1, time.Second)

	if !limiter.Allow("user-1") {
		t.Fatal("first user should be allowed")
	}

	if limiter.Allow("user-1") {
		t.Fatal("second request from first user should be blocked")
	}

	if !limiter.Allow("user-2") {
		t.Fatal("second user should have its own limit")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	limiter := NewRateLimiter(1, 20*time.Millisecond)

	if !limiter.Allow("user-1") {
		t.Fatal("first request should be allowed")
	}

	if limiter.Allow("user-1") {
		t.Fatal("request should be blocked during the window")
	}

	time.Sleep(30 * time.Millisecond)

	if !limiter.Allow("user-1") {
		t.Fatal("request should be allowed after the window")
	}
}

func TestRateLimiterRejectsEmptyKey(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second)

	if limiter.Allow("") {
		t.Fatal("empty key should not be allowed")
	}
}

func TestRateLimiterReset(t *testing.T) {
	limiter := NewRateLimiter(1, time.Second)

	if !limiter.Allow("user-1") {
		t.Fatal("first request should be allowed")
	}

	limiter.Reset("user-1")

	if !limiter.Allow("user-1") {
		t.Fatal("request should be allowed after reset")
	}
}
