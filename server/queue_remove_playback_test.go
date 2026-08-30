package main

import "testing"

func TestRemoveNextSong(t *testing.T) {
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

	third := &Song{
		ID:       "song-3",
		URL:      "https://youtube.com/watch?v=test3",
		Title:    "Third",
		Duration: 180,
		AddedBy:  owner.ID,
	}

	if err := room.Queue.Add(first); err != nil {
		t.Fatalf("add first song failed: %v", err)
	}

	if err := room.Queue.Add(second); err != nil {
		t.Fatalf("add second song failed: %v", err)
	}

	if err := room.Queue.Add(third); err != nil {
		t.Fatalf("add third song failed: %v", err)
	}

	started, err := room.Controller.StartNext()
	if err != nil {
		t.Fatalf("start next failed: %v", err)
	}

	if started.ID != first.ID {
		t.Fatalf("expected first song, got %q", started.ID)
	}

	if err := room.Queue.Remove(second.ID); err != nil {
		t.Fatalf("remove second song failed: %v", err)
	}

	next := room.Queue.Next()
	if next == nil {
		t.Fatal("expected next song")
	}

	if next.ID != third.ID {
		t.Fatalf("expected third song after removing second, got %q", next.ID)
	}
}
