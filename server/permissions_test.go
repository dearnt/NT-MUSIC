package main

import (
	"errors"
	"testing"
)

func TestRequireRoom(t *testing.T) {
	manager := NewRoomManager()

	owner := &User{
		ID:   "owner-1",
		Name: "Owner",
	}

	room, err := manager.CreateRoom(owner)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	found, err := requireRoom(manager, room.Code)
	if err != nil {
		t.Fatalf("expected room to be found: %v", err)
	}

	if found != room {
		t.Fatal("expected the same room")
	}

	if _, err := requireRoom(manager, ""); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}

	if _, err := requireRoom(manager, "INVALID"); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestRequireUser(t *testing.T) {
	manager := NewRoomManager()

	owner := &User{
		ID:   "owner-1",
		Name: "Owner",
	}

	room, err := manager.CreateRoom(owner)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	found, err := requireUser(room, owner.ID)
	if err != nil {
		t.Fatalf("expected user to be found: %v", err)
	}

	if found != owner {
		t.Fatal("expected the same user")
	}

	if _, err := requireUser(room, ""); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	if _, err := requireUser(room, "missing-user"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRequireOwner(t *testing.T) {
	manager := NewRoomManager()

	owner := &User{
		ID:   "owner-1",
		Name: "Owner",
	}

	guest := &User{
		ID:   "guest-1",
		Name: "Guest",
	}

	room, err := manager.CreateRoom(owner)
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	if err := room.AddUser(guest); err != nil {
		t.Fatalf("failed to add guest: %v", err)
	}

	if err := requireOwner(room, owner.ID); err != nil {
		t.Fatalf("expected owner to pass: %v", err)
	}

	if err := requireOwner(room, guest.ID); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner for guest, got %v", err)
	}

	if err := requireOwner(room, ""); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner for empty user ID, got %v", err)
	}

	if err := requireOwner(room, "missing-user"); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ErrNotOwner for missing user, got %v", err)
	}
}
