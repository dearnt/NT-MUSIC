package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseYouTubeDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "seconds",
			input:    "45",
			expected: 45,
		},
		{
			name:     "minutes",
			input:    "3:20",
			expected: 200,
		},
		{
			name:     "hours",
			input:    "1:02:30",
			expected: 3750,
		},
		{
			name:     "invalid",
			input:    "invalid",
			expected: 0,
		},
		{
			name:     "empty",
			input:    "",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseYouTubeDuration(test.input)

			if result != test.expected {
				t.Fatalf("expected %d, got %d", test.expected, result)
			}
		})
	}
}

func TestYouTubeExtractorValidation(t *testing.T) {
	extractor := NewYouTubeExtractor()

	if extractor == nil {
		t.Fatal("expected extractor")
	}

	if extractor.Binary != "yt-dlp" {
		t.Fatalf("expected yt-dlp binary, got %q", extractor.Binary)
	}

	if extractor.Timeout != 30*time.Second {
		t.Fatalf("expected 30 second timeout, got %v", extractor.Timeout)
	}
}

func TestYouTubeExtractorEmptyURL(t *testing.T) {
	extractor := NewYouTubeExtractor()

	metadata, err := extractor.Extract(
		context.Background(),
		"",
	)

	if metadata != nil {
		t.Fatal("expected nil metadata")
	}

	if !errors.Is(err, ErrInvalidSongURL) {
		t.Fatalf("expected ErrInvalidSongURL, got %v", err)
	}
}

func TestYouTubeExtractorInvalidBinary(t *testing.T) {
	extractor := &YouTubeExtractor{
		Binary:  "nt-music-binary-that-does-not-exist",
		Timeout: time.Second,
	}

	metadata, err := extractor.Extract(
		context.Background(),
		"https://www.youtube.com/watch?v=test",
	)

	if metadata != nil {
		t.Fatal("expected nil metadata")
	}

	if !errors.Is(err, ErrYouTubeExtractor) {
		t.Fatalf("expected ErrYouTubeExtractor, got %v", err)
	}
}

func TestYouTubeExtractorTimeout(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "slow-extractor")

	script := "#!/bin/sh\nsleep 2\n"

	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatalf("failed to create test executable: %v", err)
	}

	extractor := &YouTubeExtractor{
		Binary:  binary,
		Timeout: 50 * time.Millisecond,
	}

	metadata, err := extractor.Extract(
		context.Background(),
		"https://www.youtube.com/watch?v=test",
	)

	if metadata != nil {
		t.Fatal("expected nil metadata")
	}

	if !errors.Is(err, ErrYouTubeTimeout) {
		t.Fatalf("expected ErrYouTubeTimeout, got %v", err)
	}
}
