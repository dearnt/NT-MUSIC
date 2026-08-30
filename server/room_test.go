package main

import (
	"sync"
	"testing"
)

func TestRoomManager(t *testing.T) {
	manager := NewRoomManager()

	owner := &User{
		ID:   "owner-1",
		Name: "Owner",
	}

	room, err := manager.CreateRoom(owner)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	if room.ID == "" {
		t.Fatal("expected room ID")
	}

	if room.Code == "" {
		t.Fatal("expected room code")
	}

	if room.OwnerID != owner.ID {
		t.Fatalf("expected owner %q, got %q", owner.ID, room.OwnerID)
	}

	if room.UserCount() != 1 {
		t.Fatalf("expected user count 1, got %d", room.UserCount())
	}

	if room.ClientCount() != 0 {
		t.Fatalf("expected client count 0, got %d", room.ClientCount())
	}

	if room.Queue == nil {
		t.Fatal("expected room queue")
	}

	foundByID, exists := manager.GetRoom(room.ID)
	if !exists {
		t.Fatal("expected room to exist by ID")
	}

	if foundByID != room {
		t.Fatal("expected the same room instance")
	}

	foundByCode, exists := manager.GetRoomByCode(room.Code)
	if !exists {
		t.Fatal("expected room to exist by code")
	}

	if foundByCode != room {
		t.Fatal("expected the same room instance")
	}

	guest := &User{
		ID:   "guest-1",
		Name: "Guest",
	}

	if err := room.AddUser(guest); err != nil {
		t.Fatalf("failed to add guest: %v", err)
	}

	if room.UserCount() != 2 {
		t.Fatalf("expected user count 2, got %d", room.UserCount())
	}

	if err := room.AddUser(guest); err == nil {
		t.Fatal("expected duplicate user error")
	}

	if !room.IsOwner(owner.ID) {
		t.Fatal("expected owner check to succeed")
	}

	if room.IsOwner(guest.ID) {
		t.Fatal("expected guest not to be owner")
	}

	room.RemoveUser(guest.ID)

	if room.UserCount() != 1 {
		t.Fatalf("expected user count 1 after removal, got %d", room.UserCount())
	}

	manager.RemoveRoom(room.ID)

	if _, exists := manager.GetRoom(room.ID); exists {
		t.Fatal("expected room to be removed")
	}

	if _, exists := manager.GetRoomByCode(room.Code); exists {
		t.Fatal("expected room to be removed by code")
	}
}

func TestRoomManagerConcurrentCreation(t *testing.T) {
	manager := NewRoomManager()

	const count = 100

	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(index int) {
			defer wg.Done()

			owner := &User{
				ID:   "owner-" + string(rune(index)),
				Name: "Owner",
			}

			if _, err := manager.CreateRoom(owner); err != nil {
				t.Errorf("failed to create room: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if len(manager.Rooms) != count {
		t.Fatalf("expected %d rooms, got %d", count, len(manager.Rooms))
	}
}

func TestRoomQueueIsolation(t *testing.T) {
	manager := NewRoomManager()

	ownerA := &User{
		ID:   "owner-a",
		Name: "Owner A",
	}

	ownerB := &User{
		ID:   "owner-b",
		Name: "Owner B",
	}

	roomA, err := manager.CreateRoom(ownerA)
	if err != nil {
		t.Fatalf("failed to create room A: %v", err)
	}

	roomB, err := manager.CreateRoom(ownerB)
	if err != nil {
		t.Fatalf("failed to create room B: %v", err)
	}

	song := &Song{
		ID:       "song-a",
		URL:      "https://youtube.com/watch?v=test",
		Title:    "Room A Song",
		Duration: 180,
		AddedBy:  ownerA.ID,
	}

	if err := roomA.Queue.Add(song); err != nil {
		t.Fatalf("failed to add song to room A: %v", err)
	}

	if roomA.Queue.Len() != 1 {
		t.Fatalf("expected room A queue length 1, got %d", roomA.Queue.Len())
	}

	if roomB.Queue.Len() != 0 {
		t.Fatalf("expected room B queue length 0, got %d", roomB.Queue.Len())
	}

	if roomA.Queue == roomB.Queue {
		t.Fatal("room queues must be independent")
	}
}
