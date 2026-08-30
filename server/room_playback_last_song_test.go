package main

import (
	"testing"
	"time"
)

func TestPlaybackStopsAfterLastSong(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	song := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=test1",
		Title:    "Last Song",
		Duration: 1,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(song); err != nil {
		t.Fatalf("add song failed: %v", err)
	}

	if _, err := room.Controller.StartNext(); err != nil {
		t.Fatalf("start next failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	state := room.Playback.State()

	if state.Playing {
		t.Fatal("expected playback to stop after the last song")
	}
}
