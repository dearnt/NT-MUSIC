package main

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"nt-music/server/security"
)

type Client struct {
	Conn *websocket.Conn
	Room *Room
	User *User
	mu   sync.Mutex
}

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type WebSocketServer struct {
	Rooms *RoomManager
	Abuse *security.AbuseGuard
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebSocketServer(rooms *RoomManager) *WebSocketServer {
	return &WebSocketServer{
		Rooms: rooms,
		Abuse: security.NewAbuseGuard(),
	}
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn: conn,
	}
}

func (c *Client) send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Conn.WriteJSON(message)
}

func (c *Client) read() (Message, error) {
	var message Message

	if err := c.Conn.ReadJSON(&message); err != nil {
		return Message{}, err
	}

	return message, nil
}

func (c *Client) close() error {
	return c.Conn.Close()
}

func (s *WebSocketServer) handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := newClient(conn)
	println("WEBSOCKET CLIENT CONNECTED")
	defer s.disconnectClient(client)

	for {
		var message ClientMessage

		if err := conn.ReadJSON(&message); err != nil {
			return
		}

		switch message.Type {
		case "create_room":
			s.handleCreateRoom(client, message)
		case "join_room":
			s.handleJoinRoom(client, message)
		case "add_song":
			s.handleAddSong(client, message)
		case "play":
			s.handlePlay(client, message)
		case "seek":
			s.handleSeek(client, message)
		case "pause":
			s.handlePause(client, message)
		case "stop":
			s.handleStop(client, message)
		case "next_song":
			s.handleNextSong(client, message)
		case "play_song":
			s.handlePlaySong(client, message)
		case "set_dj":
			s.handleSetDJ(client, message)
		case "remove_dj":
			s.handleRemoveDJ(client, message)
		case "history":
			s.handleHistory(client)
		case "play_history":
			s.handlePlayHistory(client, message)
		case "set_volume":
			s.handleSetVolume(client, message)
		case "set_global_volume":
			s.handleSetGlobalVolume(client, message)
		case "set_global_mute":
			s.handleSetGlobalMute(client, message)
		case "sync_state":
			s.handleSyncState(client)
		default:
			s.writeError(client, "unknown message type")
		}
	}
}

func (s *WebSocketServer) handleCreateRoom(client *Client, message ClientMessage) {
	if message.Data.UserID == "" {
		s.writeError(client, "user_id is required")
		return
	}

	name, err := security.CleanText(message.Data.Name, security.MaxNameLength)
	if err != nil {
		s.writeError(client, "valid name is required")
		return
	}

	user := &User{
		ID:   message.Data.UserID,
		Name: name,
	}

	room, err := s.Rooms.CreateRoom(user)
	if err != nil {
		s.writeError(client, "failed to create room")
		return
	}

	client.User = user
	client.Room = room
	s.attachRoomController(room)

	if err := room.AddClient(client); err != nil {
		client.Room = nil
		client.User = nil
		s.Rooms.RemoveRoom(room.ID)
		s.writeError(client, "failed to register client")
		return
	}

	response := RoomResponse{
		ID:        room.ID,
		Code:      room.Code,
		OwnerID:   room.OwnerID,
		UserCount: room.UserCount(),
	}

	_ = client.send(Message{
		Type: "room_created",
		Data: response,
	})
}

func (s *WebSocketServer) handleJoinRoom(client *Client, message ClientMessage) {
	if message.Data.UserID == "" {
		s.writeError(client, "user_id is required")
		return
	}

	name, err := security.CleanText(message.Data.Name, security.MaxNameLength)
	if err != nil {
		s.writeError(client, "valid name is required")
		return
	}

	roomCode, err := security.CleanText(message.Data.RoomCode, security.MaxRoomCodeLength)
	if err != nil {
		s.writeError(client, "valid room_code is required")
		return
	}

	if s.Abuse == nil {
		s.Abuse = security.NewAbuseGuard()
	}

	if err := s.Abuse.CheckJoin(message.Data.UserID); err != nil {
		s.writeError(client, "too many join attempts")
		return
	}

	room, exists := s.Rooms.GetRoomByCode(roomCode)
	if !exists {
		s.writeError(client, "room not found")
		return
	}

	if err := security.CanJoinRoom(roomCode, room.UserCount()); err != nil {
		s.writeError(client, err.Error())
		return
	}

	user := &User{
		ID:   message.Data.UserID,
		Name: name,
	}

	if err := room.AddUser(user); err != nil {
		s.writeError(client, err.Error())
		return
	}

	client.User = user
	client.Room = room
	s.attachRoomController(room)

	if err := room.AddClient(client); err != nil {
		room.RemoveUser(user.ID)
		client.Room = nil
		client.User = nil
		s.writeError(client, "failed to register client")
		return
	}

	response := RoomResponse{
		ID:        room.ID,
		Code:      room.Code,
		OwnerID:   room.OwnerID,
		UserCount: room.UserCount(),
	}

	_ = client.send(Message{
		Type: "room_joined",
		Data: response,
	})

	_ = client.send(Message{
		Type: "queue_state",
		Data: client.Room.Queue.Snapshot(),
	})

	_ = client.send(Message{
		Type: "playback_state",
		Data: client.Room.Playback.State(),
	})

	_ = client.send(Message{
		Type: "history_state",
		Data: client.Room.History.List(),
	})

	s.broadcastPermissionsExcept(room, client)
	s.broadcastAudioStateExcept(room, client)
}

func (s *WebSocketServer) attachRoomController(room *Room) {
	if room == nil || room.Controller == nil {
		return
	}

	room.Controller.SetOnChange(func(song *Song) {
		s.broadcastPlayback(room)
		s.broadcastQueue(room)
		s.broadcastHistory(room)
	})
}

func (s *WebSocketServer) handleAddSong(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.URL == "" {
		s.writeError(client, "url is required")
		return
	}

	songService := NewSongService()

	song, err := songService.CreateSong(message.Data.URL, client.User.ID)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	metadataService := NewSongMetadataService(nil)

	if err := metadataService.Enrich(context.Background(), song); err != nil {
		s.writeError(client, "failed to extract song metadata")
		return
	}

	if err := room.Queue.Add(song); err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastToRoom(room, Message{
		Type: "song_added",
		Data: song,
	})

	s.broadcastQueue(room)

	if !room.Playback.State().Playing && room.Queue.CurrentSong() == nil {
		if _, err := room.Controller.StartNext(); err == nil {
			s.broadcastPlayback(room)
			s.broadcastQueue(room)
			s.broadcastHistory(room)
		}
	}
}

func (s *WebSocketServer) broadcastToRoom(room *Room, message Message) {
	room.mu.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for _, roomClient := range room.Clients {
		clients = append(clients, roomClient)
	}
	room.mu.RUnlock()

	for _, roomClient := range clients {
		_ = roomClient.send(message)
	}
}

func (s *WebSocketServer) handlePlay(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.URL != "" {
		room.Playback.Play(message.Data.URL, message.Data.Position)
		s.broadcastPlayback(room)
		return
	}

	if room.Controller == nil {
		s.writeError(client, "playback controller is not initialized")
		return
	}

	var song *Song

	if room.Queue.CurrentSong() != nil {
		song = room.Queue.CurrentSong()
		err = room.Controller.Resume()
	} else {
		song, err = room.Controller.StartNext()
	}

	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastToRoom(room, Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})

	_ = song
}

func (s *WebSocketServer) handleNextSong(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	song, err := room.Controller.StartNext()
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastToRoom(room, Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})

	s.broadcastQueue(room)
	s.broadcastHistory(room)

	_ = song
}

func (s *WebSocketServer) handlePlaySong(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.SongID == "" {
		s.writeError(client, "song_id is required")
		return
	}

	var selected *Song

	for _, song := range room.Queue.List() {
		if song.ID == message.Data.SongID {
			selected = song
			break
		}
	}

	if selected == nil {
		s.writeError(client, "song not found in queue")
		return
	}

	room.Queue.mu.Lock()
	remaining := make([]*Song, 0, len(room.Queue.Songs))
	for _, song := range room.Queue.Songs {
		if song.ID != selected.ID {
			remaining = append(remaining, song)
		}
	}
	room.Queue.Songs = remaining
	room.Queue.Current = selected
	room.Queue.mu.Unlock()

	room.Playback.Play(selected.ID, 0)

	if room.History != nil {
		_ = room.History.Add(selected)
	}

	s.broadcastToRoom(room, Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})

	s.broadcastQueue(room)
	s.broadcastHistory(room)
}

func (s *WebSocketServer) broadcastQueue(room *Room) {
	s.broadcastToRoom(room, Message{
		Type: "queue_state",
		Data: room.Queue.Snapshot(),
	})
}

func (s *WebSocketServer) writeError(client *Client, message string) {
	_ = client.send(Message{
		Type: "error",
		Data: map[string]string{
			"message": message,
		},
	})
}

func (s *WebSocketServer) disconnectClient(client *Client) {
	if client == nil {
		return
	}

	if client.Room != nil && client.User != nil {
		room := client.Room
		userID := client.User.ID

		room.RemoveClient(userID)

		if room.ClientCount() == 0 {
			removeAudioState(room)
			s.Rooms.RemoveRoom(room.ID)
		} else {
			s.broadcastPermissions(room)
			s.broadcastAudioState(room)
		}

		client.Room = nil
		client.User = nil
	}

	_ = client.close()
}

func (s *WebSocketServer) handlePause(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	room.Playback.Pause(room.Playback.State().Position)

	s.broadcastPlayback(room)
}

func (s *WebSocketServer) handleStop(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	room.Controller.Stop()

	s.broadcastPlayback(room)
}

func (s *WebSocketServer) handleSeek(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	room.Playback.Seek(message.Data.Position)

	s.broadcastPlayback(room)
}

func (s *WebSocketServer) broadcastHistory(room *Room) {
	if room == nil || room.History == nil {
		return
	}

	s.broadcastToRoom(room, Message{
		Type: "history_state",
		Data: room.History.List(),
	})
}

func (s *WebSocketServer) handleSyncState(client *Client) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	_ = client.send(Message{
		Type: "queue_state",
		Data: room.Queue.Snapshot(),
	})

	_ = client.send(Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})

	_ = client.send(Message{
		Type: "history_state",
		Data: room.History.List(),
	})

	audio := audioStateSnapshot(room)
	audio["user_volume"] = userVolume(room, user.ID)
	audio["effective_volume"] = effectiveVolume(room, user.ID)

	_ = client.send(Message{
		Type: "audio_state",
		Data: audio,
	})

	room.mu.RLock()

	members := make([]map[string]any, 0, len(room.Users))

	for _, member := range room.Users {
		role := "member"

		if member.ID == room.OwnerID {
			role = "owner"
		} else if room.DJs[member.ID] {
			role = "dj"
		}

		members = append(members, map[string]any{
			"user_id": member.ID,
			"name":    member.Name,
			"role":    role,
		})
	}

	room.mu.RUnlock()

	_ = client.send(Message{
		Type: "permissions_state",
		Data: members,
	})
}

func (s *WebSocketServer) handleHistory(client *Client) {
	room, _, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	_ = client.send(Message{
		Type: "history_state",
		Data: room.History.List(),
	})
}

func (s *WebSocketServer) handlePlayHistory(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerOrDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.SongID == "" {
		s.writeError(client, "song_id is required")
		return
	}

	song, err := room.History.Get(message.Data.SongID)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	room.Queue.mu.Lock()
	room.Queue.Current = song
	room.Queue.mu.Unlock()

	room.Playback.Play(song.ID, 0)

	if room.History != nil {
		_ = room.History.Add(song)
	}

	s.broadcastToRoom(room, Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})

	s.broadcastQueue(room)
	s.broadcastHistory(room)
}

func (s *WebSocketServer) broadcastPlayback(room *Room) {
	s.broadcastToRoom(room, Message{
		Type: "playback_state",
		Data: room.Playback.State(),
	})
}

func (s *WebSocketServer) handleSetDJ(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerCanManageDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.TargetUserID == "" {
		s.writeError(client, "target_user_id is required")
		return
	}

	if err := room.SetDJ(message.Data.TargetUserID, true); err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastPermissions(room)
}

func (s *WebSocketServer) handleRemoveDJ(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwnerCanManageDJ(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.TargetUserID == "" {
		s.writeError(client, "target_user_id is required")
		return
	}

	if err := room.SetDJ(message.Data.TargetUserID, false); err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastPermissions(room)
}

func (s *WebSocketServer) handleSetVolume(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.Volume < 0 || message.Data.Volume > 1 {
		s.writeError(client, "volume must be between 0 and 1")
		return
	}

	if err := setUserVolume(room, user.ID, message.Data.Volume); err != nil {
		s.writeError(client, err.Error())
		return
	}

	s.broadcastAudioState(room)
}

func (s *WebSocketServer) handleSetGlobalVolume(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwner(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	if message.Data.GlobalVolume < 0 || message.Data.GlobalVolume > 1 {
		s.writeError(client, "global volume must be between 0 and 1")
		return
	}

	setGlobalVolume(room, message.Data.GlobalVolume)
	s.broadcastAudioState(room)
}

func (s *WebSocketServer) handleSetGlobalMute(client *Client, message ClientMessage) {
	room, user, err := requireClientRoom(client)
	if err != nil {
		s.writeError(client, err.Error())
		return
	}

	if err := requireOwner(room, user.ID); err != nil {
		s.writeError(client, err.Error())
		return
	}

	setGlobalMute(room, message.Data.Enabled)
	s.broadcastAudioState(room)
}

func (s *WebSocketServer) sendAudioState(client *Client) {
	if client == nil || client.Room == nil || client.User == nil {
		return
	}

	data := audioStateSnapshot(client.Room)
	data["user_volume"] = userVolume(client.Room, client.User.ID)
	data["effective_volume"] = effectiveVolume(client.Room, client.User.ID)

	_ = client.send(Message{
		Type: "audio_state",
		Data: data,
	})
}

func (s *WebSocketServer) broadcastAudioState(room *Room) {
	if room == nil {
		return
	}

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client)
	}
	room.mu.RUnlock()

	for _, client := range clients {
		data := audioStateSnapshot(room)
		data["user_volume"] = userVolume(room, client.User.ID)
		data["effective_volume"] = effectiveVolume(room, client.User.ID)

		_ = client.send(Message{
			Type: "audio_state",
			Data: data,
		})
	}
}

func (s *WebSocketServer) broadcastPermissionsExcept(room *Room, excluded *Client) {
	if room == nil {
		return
	}

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		if client != excluded {
			clients = append(clients, client)
		}
	}
	users := make([]map[string]any, 0, len(room.Users))

	for _, user := range room.Users {
		role := "member"

		if user.ID == room.OwnerID {
			role = "owner"
		} else if room.DJs[user.ID] {
			role = "dj"
		}

		users = append(users, map[string]any{
			"user_id": user.ID,
			"name":    user.Name,
			"role":    role,
		})
	}
	room.mu.RUnlock()

	message := Message{
		Type: "permissions_state",
		Data: users,
	}

	for _, client := range clients {
		_ = client.send(message)
	}
}

func (s *WebSocketServer) broadcastAudioStateExcept(room *Room, excluded *Client) {
	if room == nil {
		return
	}

	room.mu.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		if client != excluded {
			clients = append(clients, client)
		}
	}
	room.mu.RUnlock()

	for _, client := range clients {
		data := audioStateSnapshot(room)
		data["user_volume"] = userVolume(room, client.User.ID)
		data["effective_volume"] = effectiveVolume(room, client.User.ID)

		_ = client.send(Message{
			Type: "audio_state",
			Data: data,
		})
	}
}

func (s *WebSocketServer) broadcastPermissions(room *Room) {
	if room == nil {
		return
	}

	room.mu.RLock()
	users := make([]map[string]any, 0, len(room.Users))

	for _, user := range room.Users {
		role := "member"
		if user.ID == room.OwnerID {
			role = "owner"
		} else if room.DJs[user.ID] {
			role = "dj"
		}

		users = append(users, map[string]any{
			"user_id": user.ID,
			"name":    user.Name,
			"role":    role,
		})
	}

	room.mu.RUnlock()

	s.broadcastToRoom(room, Message{
		Type: "permissions_state",
		Data: users,
	})
}
