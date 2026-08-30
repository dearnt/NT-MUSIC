package main

import "errors"

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrUserNotFound = errors.New("user not found")
	ErrNotOwner     = errors.New("user is not the room owner")
	ErrNotDJ        = ErrNotOwner
)

func requireRoom(manager *RoomManager, roomCode string) (*Room, error) {
	if roomCode == "" {
		return nil, ErrRoomNotFound
	}

	room, exists := manager.GetRoomByCode(roomCode)
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

func requireUser(room *Room, userID string) (*User, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	user, exists := room.Users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func requireOwner(room *Room, userID string) error {
	if userID == "" {
		return ErrNotOwner
	}

	if !room.IsOwner(userID) {
		return ErrNotOwner
	}

	return nil
}

func requireClientRoom(client *Client) (*Room, *User, error) {
	if client == nil || client.Room == nil || client.User == nil {
		return nil, nil, errors.New("must be in a room")
	}

	user, err := requireUser(client.Room, client.User.ID)
	if err != nil {
		return nil, nil, err
	}

	return client.Room, user, nil
}

func requireOwnerOrDJ(room *Room, userID string) error {
	if userID == "" {
		return ErrNotDJ
	}

	if room.IsOwner(userID) || room.IsDJ(userID) {
		return nil
	}

	return ErrNotDJ
}

func requireOwnerCanManageDJ(room *Room, userID string) error {
	if userID == "" || !room.IsOwner(userID) {
		return ErrNotOwner
	}

	return nil
}
