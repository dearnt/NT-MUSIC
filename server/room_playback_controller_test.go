package main

import "testing"

func TestRoomPlaybackControllerStartNext(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)
	controller := NewRoomPlaybackController(room)

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

	started, err := controller.StartNext()
	if err != nil {
		t.Fatalf("start next failed: %v", err)
	}

	if started == nil {
		t.Fatal("expected started song")
	}

	if started.ID != song.ID {
		t.Fatalf("expected song %q, got %q", song.ID, started.ID)
	}

	state := room.Playback.State()

	if state.SongID != song.ID {
		t.Fatalf("expected playback song %q, got %q", song.ID, state.SongID)
	}

	if !state.Playing {
		t.Fatal("expected playback to be playing")
	}
}

func TestRoomPlaybackControllerEmptyQueue(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)
	controller := NewRoomPlaybackController(room)

	song, err := controller.StartNext()

	if err != ErrQueueEmpty {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}

	if song != nil {
		t.Fatalf("expected nil song, got %v", song)
	}

	state := room.Playback.State()

	if state.Playing {
		t.Fatal("playback should not be playing")
	}
}

func TestRoomPlaybackControllerStop(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)
	controller := NewRoomPlaybackController(room)

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

	if _, err := controller.StartNext(); err != nil {
		t.Fatalf("start next failed: %v", err)
	}

	controller.Stop()

	state := room.Playback.State()

	if state.Playing {
		t.Fatal("playback should be stopped")
	}
}
