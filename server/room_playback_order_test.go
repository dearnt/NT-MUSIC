package main

import (
	"testing"
	"time"
)

func TestPlaybackKeepsQueueOrder(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	songs := []*Song{
		{
			ID:       "song-1",
			URL:      "https://youtube.com/watch?v=test1",
			Title:    "First",
			Duration: 1,
			AddedBy:  owner.ID,
		},
		{
			ID:       "song-2",
			URL:      "https://youtube.com/watch?v=test2",
			Title:    "Second",
			Duration: 1,
			AddedBy:  owner.ID,
		},
		{
			ID:       "song-3",
			URL:      "https://youtube.com/watch?v=test3",
			Title:    "Third",
			Duration: 1,
			AddedBy:  owner.ID,
		},
	}

	for _, song := range songs {
		if err := room.Queue.Add(song); err != nil {
			t.Fatalf("add song failed: %v", err)
		}
	}

	if _, err := room.Controller.StartNext(); err != nil {
		t.Fatalf("start first song failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond)

	state := room.Playback.State()
	if state.SongID != "song-2" {
		t.Fatalf("expected second song, got %q", state.SongID)
	}

	time.Sleep(1100 * time.Millisecond)

	state = room.Playback.State()
	if state.SongID != "song-3" {
		t.Fatalf("expected third song, got %q", state.SongID)
	}

	time.Sleep(1100 * time.Millisecond)

	state = room.Playback.State()
	if state.Playing {
		t.Fatal("expected playback to stop after last song")
	}
}
