package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSongServiceCreateSong(t *testing.T) {
	service := NewSongService()

	song, err := service.CreateSong(
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"user-1",
	)
	if err != nil {
		t.Fatalf("failed to create song: %v", err)
	}

	if song == nil {
		t.Fatal("expected song")
	}

	if song.ID == "" {
		t.Fatal("expected song ID")
	}

	if song.URL != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
		t.Fatalf("unexpected URL: %q", song.URL)
	}

	if song.AddedBy != "user-1" {
		t.Fatalf("expected user-1, got %q", song.AddedBy)
	}

	if song.Title != "" {
		t.Fatalf("expected empty title before metadata extraction, got %q", song.Title)
	}

	if song.Duration != 0 {
		t.Fatalf("expected zero duration before metadata extraction, got %d", song.Duration)
	}
}

func TestSongServiceYouTubeURLs(t *testing.T) {
	service := NewSongService()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://www.youtube.com/watch?v=test",
			expected: "https://www.youtube.com/watch?v=test",
		},
		{
			input:    "https://youtube.com/watch?v=test",
			expected: "https://www.youtube.com/watch?v=test",
		},
		{
			input:    "https://m.youtube.com/watch?v=test",
			expected: "https://www.youtube.com/watch?v=test",
		},
		{
			input:    "https://music.youtube.com/watch?v=test",
			expected: "https://www.youtube.com/watch?v=test",
		},
		{
			input:    "https://youtu.be/test",
			expected: "https://www.youtube.com/watch?v=test",
		},
		{
			input:    "https://www.youtu.be/test",
			expected: "https://www.youtube.com/watch?v=test",
		},
	}

	for _, tt := range tests {
		song, err := service.CreateSong(tt.input, "user-1")
		if err != nil {
			t.Errorf("expected URL %q to be accepted: %v", tt.input, err)
			continue
		}

		if song.URL != tt.expected {
			t.Errorf("expected URL %q, got %q", tt.expected, song.URL)
		}
	}
}

func TestSongServiceRejectsInvalidURLs(t *testing.T) {
	service := NewSongService()

	urls := []string{
		"",
		"   ",
		"not-a-url",
		"https://example.com/video",
		"https://vimeo.com/12345",
		"https://soundcloud.com/test",
		"https://youtube.com/",
		"https://youtu.be/",
		"https://youtube.com/watch",
	}

	for _, songURL := range urls {
		song, err := service.CreateSong(songURL, "user-1")

		if err == nil {
			t.Errorf("expected URL %q to be rejected", songURL)
			continue
		}

		if song != nil {
			t.Errorf("expected nil song for URL %q", songURL)
		}

		if songURL == "" || strings.TrimSpace(songURL) == "" {
			if !errors.Is(err, ErrEmptySongURL) {
				t.Errorf("expected ErrEmptySongURL for %q, got %v", songURL, err)
			}
		} else if !errors.Is(err, ErrInvalidSongURL) {
			t.Errorf("expected ErrInvalidSongURL for %q, got %v", songURL, err)
		}
	}
}

func TestSongServiceRequiresUser(t *testing.T) {
	service := NewSongService()

	song, err := service.CreateSong(
		"https://www.youtube.com/watch?v=test",
		"",
	)

	if song != nil {
		t.Fatal("expected nil song")
	}

	if !errors.Is(err, ErrEmptyUserID) {
		t.Fatalf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestSongServiceGeneratesUniqueIDs(t *testing.T) {
	service := NewSongService()

	first, err := service.CreateSong(
		"https://www.youtube.com/watch?v=first",
		"user-1",
	)
	if err != nil {
		t.Fatalf("failed to create first song: %v", err)
	}

	second, err := service.CreateSong(
		"https://www.youtube.com/watch?v=second",
		"user-1",
	)
	if err != nil {
		t.Fatalf("failed to create second song: %v", err)
	}

	if first.ID == second.ID {
		t.Fatal("expected unique song IDs")
	}
}

func TestNormalizeYouTubeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "youtu.be with playlist",
			input:    "https://youtu.be/NtvoxOIQ9g0?list=RDGiyiITii080",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
		{
			name:     "youtube watch with playlist",
			input:    "https://www.youtube.com/watch?v=NtvoxOIQ9g0&list=RDGiyiITii080",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
		{
			name:     "youtube mobile",
			input:    "https://m.youtube.com/watch?v=NtvoxOIQ9g0",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
		{
			name:     "youtube music",
			input:    "https://music.youtube.com/watch?v=NtvoxOIQ9g0",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
		{
			name:     "youtube standard",
			input:    "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
		{
			name:     "youtu.be without query",
			input:    "https://youtu.be/NtvoxOIQ9g0",
			expected: "https://www.youtube.com/watch?v=NtvoxOIQ9g0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeYouTubeURL(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNormalizeYouTubeURLRejectsInvalidURLs(t *testing.T) {
	tests := []string{
		"",
		"not-a-url",
		"https://example.com/video",
		"https://youtube.com/",
		"https://youtube.com/watch",
		"https://youtu.be/",
	}

	for _, input := range tests {
		result, err := normalizeYouTubeURL(input)

		if err == nil {
			t.Errorf("expected error for %q", input)
		}

		if result != "" {
			t.Errorf("expected empty result for %q, got %q", input, result)
		}
	}
}
