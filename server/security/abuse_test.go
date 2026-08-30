package security

import "testing"

func TestAbuseGuardBlocksRepeatedJoins(t *testing.T) {
	guard := NewAbuseGuard()

	for i := 0; i < 5; i++ {
		if err := guard.CheckJoin("session-1"); err != nil {
			t.Fatalf("join %d should be allowed: %v", i+1, err)
		}
	}

	if err := guard.CheckJoin("session-1"); err != ErrActionBlocked {
		t.Fatalf("expected join to be blocked, got %v", err)
	}
}

func TestAbuseGuardSeparatesSessions(t *testing.T) {
	guard := NewAbuseGuard()

	for i := 0; i < 5; i++ {
		if err := guard.CheckJoin("session-1"); err != nil {
			t.Fatalf("join %d should be allowed: %v", i+1, err)
		}
	}

	if err := guard.CheckJoin("session-2"); err != nil {
		t.Fatalf("different session should be allowed: %v", err)
	}
}

func TestAbuseGuardBlocksActions(t *testing.T) {
	guard := NewAbuseGuard()

	for i := 0; i < 30; i++ {
		if err := guard.CheckAction("session-1"); err != nil {
			t.Fatalf("action %d should be allowed: %v", i+1, err)
		}
	}

	if err := guard.CheckAction("session-1"); err != ErrActionBlocked {
		t.Fatalf("expected action to be blocked, got %v", err)
	}
}

func TestAbuseGuardBlocksMessages(t *testing.T) {
	guard := NewAbuseGuard()

	for i := 0; i < 60; i++ {
		if err := guard.CheckMessage("session-1"); err != nil {
			t.Fatalf("message %d should be allowed: %v", i+1, err)
		}
	}

	if err := guard.CheckMessage("session-1"); err != ErrActionBlocked {
		t.Fatalf("expected message to be blocked, got %v", err)
	}
}
