package main

import (
	"testing"
)

func TestQueue(t *testing.T) {
	queue := NewQueue()

	first := &Song{
		ID:       "song-1",
		URL:      "https://youtube.com/watch?v=one",
		Title:    "First Song",
		Duration: 180,
		AddedBy:  "user-1",
	}

	second := &Song{
		ID:       "song-2",
		URL:      "https://youtube.com/watch?v=two",
		Title:    "Second Song",
		Duration: 200,
		AddedBy:  "user-1",
	}

	if err := queue.Add(first); err != nil {
		t.Fatalf("failed to add first song: %v", err)
	}

	if err := queue.Add(second); err != nil {
		t.Fatalf("failed to add second song: %v", err)
	}

	if queue.Len() != 2 {
		t.Fatalf("expected queue length 2, got %d", queue.Len())
	}

	songs := queue.List()

	if len(songs) != 2 {
		t.Fatalf("expected 2 songs, got %d", len(songs))
	}

	if songs[0].ID != "song-1" {
		t.Fatalf("expected first song, got %q", songs[0].ID)
	}

	if songs[1].ID != "song-2" {
		t.Fatalf("expected second song, got %q", songs[1].ID)
	}

	current := queue.Next()

	if current == nil {
		t.Fatal("expected current song")
	}

	if current.ID != "song-1" {
		t.Fatalf("expected song-1, got %q", current.ID)
	}

	if queue.Len() != 1 {
		t.Fatalf("expected queue length 1 after Next, got %d", queue.Len())
	}

	current = queue.CurrentSong()

	if current == nil {
		t.Fatal("expected current song")
	}

	if current.ID != "song-1" {
		t.Fatalf("expected current song-1, got %q", current.ID)
	}

	if err := queue.Remove("song-2"); err != nil {
		t.Fatalf("failed to remove song: %v", err)
	}

	if queue.Len() != 0 {
		t.Fatalf("expected empty queue, got %d", queue.Len())
	}

	if err := queue.Remove("missing"); err == nil {
		t.Fatal("expected error when removing missing song")
	}

	queue.Clear()

	if queue.Len() != 0 {
		t.Fatalf("expected empty queue after clear, got %d", queue.Len())
	}

	if queue.CurrentSong() != nil {
		t.Fatal("expected no current song after clear")
	}
}

func TestQueueValidation(t *testing.T) {
	queue := NewQueue()

	if err := queue.Add(nil); err == nil {
		t.Fatal("expected error for nil song")
	}

	if err := queue.Add(&Song{}); err == nil {
		t.Fatal("expected error for missing song id")
	}

	if err := queue.Add(&Song{
		ID: "song-1",
	}); err == nil {
		t.Fatal("expected error for missing song url")
	}

	if song := queue.Next(); song != nil {
		t.Fatal("expected nil from empty queue")
	}
}
