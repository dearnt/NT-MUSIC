package main

import "testing"

func TestRoomPlaybackStartsQueuedSong(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	song := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=test1",
		Title:    "Test Song",
		Duration: 180,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(song); err != nil {
		t.Fatalf("add song failed: %v", err)
	}

	next := room.Queue.Next()
	if next == nil {
		t.Fatal("expected queued song")
	}

	room.Playback.Play(next.ID, 0)

	state := room.Playback.State()

	if state.SongID != song.ID {
		t.Fatalf("expected song %q, got %q", song.ID, state.SongID)
	}

	if !state.Playing {
		t.Fatal("expected playback to be playing")
	}

	if room.Queue.Len() != 0 {
		t.Fatalf("expected empty queue, got %d", room.Queue.Len())
	}
}

func TestRoomPlaybackStartsNextSong(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	first := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=test1",
		Title:    "First Song",
		Duration: 180,
		AddedBy:  owner.ID,
	}

	second := &Song{
		ID:       "song-2",
		URL:      "https://youtube.com/watch?v=test2",
		Title:    "Second Song",
		Duration: 200,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(first); err != nil {
		t.Fatalf("add first song failed: %v", err)
	}

	if err := room.Queue.Add(second); err != nil {
		t.Fatalf("add second song failed: %v", err)
	}

	current := room.Queue.Next()
	if current == nil {
		t.Fatal("expected first song")
	}

	room.Playback.Play(current.ID, 0)

	state := room.Playback.State()

	if state.SongID != first.ID {
		t.Fatalf("expected first song, got %q", state.SongID)
	}

	room.Playback.Stop()

	current = room.Queue.Next()
	if current == nil {
		t.Fatal("expected second song")
	}

	room.Playback.Play(current.ID, 0)

	state = room.Playback.State()

	if state.SongID != second.ID {
		t.Fatalf("expected second song, got %q", state.SongID)
	}

	if !state.Playing {
		t.Fatal("expected second song to be playing")
	}
}
