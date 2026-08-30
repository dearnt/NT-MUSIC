package main

import "testing"

func TestRoomInitializesPlaybackController(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	if room.Controller == nil {
		t.Fatal("room playback controller was not initialized")
	}

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

	started, err := room.Controller.StartNext()
	if err != nil {
		t.Fatalf("start next failed: %v", err)
	}

	if started.ID != song.ID {
		t.Fatalf("expected %q, got %q", song.ID, started.ID)
	}
}
