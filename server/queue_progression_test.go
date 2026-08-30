package main

import "testing"

func TestQueueProgression(t *testing.T) {
	q := NewQueue()

	songs := []*Song{
		{ID: "song-a", URL: "https://youtube.com/watch?v=a"},
		{ID: "song-b", URL: "https://youtube.com/watch?v=b"},
		{ID: "song-c", URL: "https://youtube.com/watch?v=c"},
	}

	for _, song := range songs {
		if err := q.Add(song); err != nil {
			t.Fatalf("add failed: %v", err)
		}
	}

	if q.Len() != 3 {
		t.Fatalf("expected queue length 3, got %d", q.Len())
	}

	for _, expected := range songs {
		current := q.Next()

		if current == nil {
			t.Fatal("expected next song, got nil")
		}

		if current.ID != expected.ID {
			t.Fatalf("expected %q, got %q", expected.ID, current.ID)
		}

		if q.CurrentSong() == nil {
			t.Fatal("current song should not be nil")
		}

		if q.CurrentSong().ID != expected.ID {
			t.Fatalf("current song mismatch: expected %q, got %q",
				expected.ID, q.CurrentSong().ID)
		}
	}

	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got %d", q.Len())
	}

	if q.Next() != nil {
		t.Fatal("expected nil after queue is exhausted")
	}

	if q.CurrentSong() != nil {
		t.Fatal("current song should be nil after queue exhaustion")
	}
}
