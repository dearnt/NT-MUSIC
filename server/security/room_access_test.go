package security

import "testing"

func TestCanJoinRoom(t *testing.T) {
	if err := CanJoinRoom("ABC123", 0); err != nil {
		t.Fatalf("expected room join to be allowed: %v", err)
	}
}

func TestCanJoinRoomRejectsInvalidCode(t *testing.T) {
	if err := CanJoinRoom("", 0); err != ErrRoomAccessDenied {
		t.Fatalf("expected room access denied, got %v", err)
	}
}

func TestCanJoinRoomRejectsFullRoom(t *testing.T) {
	if err := CanJoinRoom("ABC123", MaxRoomUsers); err != ErrRoomFull {
		t.Fatalf("expected room full, got %v", err)
	}
}

func TestCanLeaveRoom(t *testing.T) {
	if !CanLeaveRoom("user-1") {
		t.Fatal("expected valid user to leave")
	}

	if CanLeaveRoom("") {
		t.Fatal("expected empty user id to be rejected")
	}
}
