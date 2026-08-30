package security

import (
	"testing"
	"time"
)

func TestWebSocketGuardClientLimit(t *testing.T) {
	guard := NewWebSocketGuard(2, 20)

	if err := guard.AddClient(); err != nil {
		t.Fatalf("first client should be allowed: %v", err)
	}

	if err := guard.AddClient(); err != nil {
		t.Fatalf("second client should be allowed: %v", err)
	}

	if err := guard.AddClient(); err != ErrTooManyClients {
		t.Fatalf("expected too many clients, got %v", err)
	}

	guard.RemoveClient()

	if err := guard.AddClient(); err != nil {
		t.Fatalf("client should be allowed after removal: %v", err)
	}
}

func TestWebSocketGuardMessageSize(t *testing.T) {
	guard := NewWebSocketGuard(10, 20)

	if err := guard.CheckMessageSize(MaxMessageSize); err != nil {
		t.Fatalf("maximum valid message should be accepted: %v", err)
	}

	if err := guard.CheckMessageSize(MaxMessageSize + 1); err != ErrMessageTooLarge {
		t.Fatalf("expected message too large, got %v", err)
	}
}

func TestWebSocketGuardRateLimit(t *testing.T) {
	guard := NewWebSocketGuard(10, 2)

	if !guard.Allow("client-1") {
		t.Fatal("first request should be allowed")
	}

	if !guard.Allow("client-1") {
		t.Fatal("second request should be allowed")
	}

	if guard.Allow("client-1") {
		t.Fatal("third request should be blocked")
	}

	time.Sleep(time.Second + 50*time.Millisecond)

	if !guard.Allow("client-1") {
		t.Fatal("request should be allowed after rate limit reset")
	}
}
