package security

import "testing"

func TestSessionIDIsUnique(t *testing.T) {
	first, err := NewSessionID()
	if err != nil {
		t.Fatalf("first session generation failed: %v", err)
	}

	second, err := NewSessionID()
	if err != nil {
		t.Fatalf("second session generation failed: %v", err)
	}

	if first == second {
		t.Fatal("session IDs should be unique")
	}

	if len(first) != sessionIDBytes*2 {
		t.Fatalf("unexpected session ID length: %d", len(first))
	}
}

func TestValidateSessionIDRejectsInvalidValues(t *testing.T) {
	tests := []string{
		"",
		"123",
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		"000000000000000000000000000000000000000000000000000000000000000G",
	}

	for _, id := range tests {
		if err := ValidateSessionID(id); err != ErrInvalidIdentity {
			t.Fatalf("expected invalid identity for %q, got %v", id, err)
		}
	}
}

func TestValidateSessionIDAcceptsValidValue(t *testing.T) {
	id := "0123456789abcdef0123456789abcdef"

	if err := ValidateSessionID(id); err != nil {
		t.Fatalf("valid session ID was rejected: %v", err)
	}
}
