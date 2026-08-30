package main

import "testing"

func TestRemoveCurrentSong(t *testing.T) {
	owner := &User{
		ID:   "owner",
		Name: "Owner",
	}

	room := NewRoom("room-1", "ABC123", owner)

	first := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=test1",
		Title:    "First",
		Duration: 180,
		AddedBy:  owner.ID,
	}

	second := &Song{
		ID:       "song-2",
		URL:      "https://youtube.com/watch?v=test2",
		Title:    "Second",
		Duration: 180,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(first); err != nil {
		t.Fatalf("add first song failed: %v", err)
	}

	if err := room.Queue.Add(second); err != nil {
		t.Fatalf("add second song failed: %v", err)
	}

	if got := room.Queue.Next(); got == nil || got.ID != first.ID {
		t.Fatalf("expected first song to be current")
	}

	if err := room.Queue.RemoveCurrent(); err != nil {
		t.Fatalf("remove current song failed: %v", err)
	}

	if current := room.Queue.CurrentSong(); current != nil {
		t.Fatalf("expected no current song, got %q", current.ID)
	}
}
