package main

import (
	"testing"
	"time"
)

func TestPlaybackStartsSongAddedDuringPlayback(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	first := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=test1",
		Title:    "First",
		Duration: 1,
		AddedBy:  owner.ID,
	}

	second := &Song{
		ID:       "song-2",
		URL:      "https://youtube.com/watch?v=test2",
		Title:    "Second",
		Duration: 1,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(first); err != nil {
		t.Fatalf("add first song failed: %v", err)
	}

	if _, err := room.Controller.StartNext(); err != nil {
		t.Fatalf("start first song failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	if err := room.Queue.Add(second); err != nil {
		t.Fatalf("add second song failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	state := room.Playback.State()

	if state.SongID != second.ID {
		t.Fatalf("expected second song %q, got %q", second.ID, state.SongID)
	}

	if !state.Playing {
		t.Fatal("expected second song to be playing")
	}
}
