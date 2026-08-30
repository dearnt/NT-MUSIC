package main

import "testing"

func TestPlaybackQueueTransition(t *testing.T) {
	q := NewQueue()
	p := NewPlayback()

	first := &Song{
		ID:  "song-a",
		URL: "https://youtube.com/watch?v=a",
	}

	second := &Song{
		ID:  "song-b",
		URL: "https://youtube.com/watch?v=b",
	}

	if err := q.Add(first); err != nil {
		t.Fatalf("add first song failed: %v", err)
	}

	if err := q.Add(second); err != nil {
		t.Fatalf("add second song failed: %v", err)
	}

	song := q.Next()
	if song == nil {
		t.Fatal("expected first song")
	}

	p.Play(song.ID, 0)

	state := p.State()

	if state.SongID != "song-a" {
		t.Fatalf("expected first song, got %q", state.SongID)
	}

	if !state.Playing {
		t.Fatal("first song should be playing")
	}

	p.Stop()

	song = q.Next()
	if song == nil {
		t.Fatal("expected second song")
	}

	p.Play(song.ID, 0)

	state = p.State()

	if state.SongID != "song-b" {
		t.Fatalf("expected second song, got %q", state.SongID)
	}

	if !state.Playing {
		t.Fatal("second song should be playing")
	}
}
