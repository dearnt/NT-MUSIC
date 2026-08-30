package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

const roomCodeLength = 6

const roomCodeCharacters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type User struct {
	ID   string
	Name string
}

type Room struct {
	ID         string
	Code       string
	OwnerID    string
	Users      map[string]*User
	Clients    map[string]*Client
	DJs        map[string]bool
	Queue      *Queue
	History    *SongHistory
	Playback   *Playback
	Controller *RoomPlaybackController
	mu         sync.RWMutex
}

type RoomManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

func NewRoom(id string, code string, owner *User) *Room {
	room := &Room{
		ID:      id,
		Code:    code,
		OwnerID: owner.ID,
		Users: map[string]*User{
			owner.ID: owner,
		},
		Clients:  make(map[string]*Client),
		DJs:      make(map[string]bool),
		Queue:    NewQueue(),
		History:  NewSongHistory(),
		Playback: NewPlayback(),
	}

	room.Controller = NewRoomPlaybackController(room)
	go room.Controller.Run()

	return room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room),
	}
}

func (m *RoomManager) CreateRoom(owner *User) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, err := generateRoomID()
	if err != nil {
		return nil, err
	}

	code, err := generateRoomCode()
	if err != nil {
		return nil, err
	}

	room := NewRoom(id, code, owner)
	m.Rooms[id] = room

	return room, nil
}

func (m *RoomManager) GetRoom(id string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	room, exists := m.Rooms[id]
	return room, exists
}

func (m *RoomManager) GetRoomByCode(code string) (*Room, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, room := range m.Rooms {
		if room.Code == code {
			return room, true
		}
	}

	return nil, false
}

func (m *RoomManager) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Rooms, id)
}

func (r *Room) AddUser(user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Users[user.ID]; exists {
		return errors.New("user already exists")
	}

	r.Users[user.ID] = user
	return nil
}

func (r *Room) RemoveUser(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Users, userID)
}

func (r *Room) AddClient(client *Client) error {
	if client == nil || client.User == nil {
		return errors.New("client user is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Clients[client.User.ID]; exists {
		return errors.New("client already exists")
	}

	r.Clients[client.User.ID] = client
	return nil
}

func (r *Room) RemoveClient(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Clients, userID)
	delete(r.Users, userID)
}

func (r *Room) IsOwner(userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.OwnerID == userID
}

func (r *Room) IsDJ(userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.DJs[userID]
}

func (r *Room) SetDJ(userID string, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if userID == "" {
		return ErrUserNotFound
	}

	if _, exists := r.Users[userID]; !exists {
		return ErrUserNotFound
	}

	if userID == r.OwnerID {
		return nil
	}

	if enabled {
		r.DJs[userID] = true
	} else {
		delete(r.DJs, userID)
	}

	return nil
}

func (r *Room) Role(userID string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if userID == r.OwnerID {
		return "owner"
	}

	if r.DJs[userID] {
		return "dj"
	}

	return "member"
}

func (r *Room) UserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Users)
}

func (r *Room) ClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Clients)
}

func generateRoomID() (string, error) {
	data := make([]byte, 16)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}

func generateRoomCode() (string, error) {
	code := make([]byte, roomCodeLength)

	for i := range code {
		value := make([]byte, 1)
		if _, err := rand.Read(value); err != nil {
			return "", err
		}
		code[i] = roomCodeCharacters[int(value[0])%len(roomCodeCharacters)]
	}

	return string(code), nil
}
