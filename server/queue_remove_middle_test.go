package main

import "testing"

func TestRemoveSongKeepsQueueOrder(t *testing.T) {
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

	if err := room.Queue.Remove(second.ID); err != nil {
		t.Fatalf("remove middle song failed: %v", err)
	}

	next := room.Queue.Next()
	if next == nil {
		t.Fatal("expected first song")
	}

	if next.ID != first.ID {
		t.Fatalf("expected %q, got %q", first.ID, next.ID)
	}

	next = room.Queue.Next()
	if next == nil {
		t.Fatal("expected third song")
	}

	if next.ID != third.ID {
		t.Fatalf("expected %q after removing middle song, got %q", third.ID, next.ID)
	}
}
