package security

import "testing"

func TestValidOrigin(t *testing.T) {
	allowed := []string{"localhost:8080", "music.example.com"}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{"localhost", "http://localhost:8080", true},
		{"https host", "https://music.example.com", true},
		{"wrong host", "https://evil.example.com", false},
		{"wrong scheme", "ftp://music.example.com", false},
		{"empty origin", "", false},
		{"missing host", "https://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidOrigin(tt.origin, allowed); got != tt.want {
				t.Fatalf("ValidOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
