package security

import "testing"

func TestSecurityIdentity(t *testing.T) {
	id, err := NewSessionID()
	if err != nil {
		t.Fatalf("session id generation failed: %v", err)
	}

	if err := ValidateSessionID(id); err != nil {
		t.Fatalf("generated session id should be valid: %v", err)
	}

	if id == "" {
		t.Fatal("session id should not be empty")
	}
}

func TestSecurityInputLimits(t *testing.T) {
	if !ValidLength("Alice", MaxNameLength) {
		t.Fatal("valid name should be accepted")
	}

	if ValidLength("", MaxNameLength) {
		t.Fatal("empty value should be rejected")
	}

	if !ValidMessageSize(MaxMessageSize) {
		t.Fatal("maximum message size should be accepted")
	}

	if ValidMessageSize(MaxMessageSize + 1) {
		t.Fatal("oversized message should be rejected")
	}

	if !ValidQueueSize(MaxQueueSize) {
		t.Fatal("maximum queue size should be accepted")
	}

	if ValidQueueSize(MaxQueueSize + 1) {
		t.Fatal("oversized queue should be rejected")
	}
}

func TestSecurityRoomAccess(t *testing.T) {
	if err := CanJoinRoom("ABC123", 0); err != nil {
		t.Fatalf("valid room should allow access: %v", err)
	}

	if err := CanJoinRoom("", 0); err != ErrRoomAccessDenied {
		t.Fatalf("invalid room should be denied, got %v", err)
	}

	if err := CanJoinRoom("ABC123", MaxRoomUsers); err != ErrRoomFull {
		t.Fatalf("full room should be rejected, got %v", err)
	}
}

func TestSecurityOrigin(t *testing.T) {
	allowed := []string{
		"localhost:8080",
		"music.example.com",
	}

	if !ValidOrigin("http://localhost:8080", allowed) {
		t.Fatal("allowed origin should pass")
	}

	if ValidOrigin("https://evil.example.com", allowed) {
		t.Fatal("unknown origin should fail")
	}

	if ValidOrigin("ftp://music.example.com", allowed) {
		t.Fatal("invalid scheme should fail")
	}
}

func TestSecurityWebSocketGuard(t *testing.T) {
	guard := NewWebSocketGuard(1, 2)

	id, err := NewSessionID()
	if err != nil {
		t.Fatalf("session id generation failed: %v", err)
	}

	if err := guard.Connect(id); err != nil {
		t.Fatalf("first connection should succeed: %v", err)
	}

	if guard.ClientCount() != 1 {
		t.Fatalf("expected one client, got %d", guard.ClientCount())
	}

	if err := guard.Connect(id); err != ErrTooManyClients {
		t.Fatalf("second connection should be rejected, got %v", err)
	}

	guard.Disconnect()

	if guard.ClientCount() != 0 {
		t.Fatalf("expected zero clients, got %d", guard.ClientCount())
	}

	if err := guard.CheckMessageSize(MaxMessageSize + 1); err != ErrMessageTooLarge {
		t.Fatalf("oversized message should be rejected, got %v", err)
	}
}
