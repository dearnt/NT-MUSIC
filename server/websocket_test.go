package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocketServer(t *testing.T) {
	roomManager := NewRoomManager()
	server := NewWebSocketServer(roomManager)

	httpServer := httptest.NewServer(http.HandlerFunc(server.handler))
	defer httpServer.Close()

	url := "ws" + httpServer.URL[4:]

	ownerConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect owner: %v", err)
	}

	createRequest := ClientMessage{
		Type: "create_room",
		Data: ClientMessageData{
			UserID: "user-1",
			Name:   "Owner",
		},
	}

	if err := ownerConn.WriteJSON(createRequest); err != nil {
		ownerConn.Close()
		t.Fatalf("failed to send create request: %v", err)
	}

	var createResponse ServerMessage

	if err := ownerConn.ReadJSON(&createResponse); err != nil {
		ownerConn.Close()
		t.Fatalf("failed to read create response: %v", err)
	}

	if createResponse.Type != "room_created" {
		ownerConn.Close()
		t.Fatalf("expected room_created, got %q", createResponse.Type)
	}

	roomData, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		ownerConn.Close()
		t.Fatalf("unexpected room response type: %T", createResponse.Data)
	}

	roomCode, ok := roomData["code"].(string)
	if !ok || roomCode == "" {
		ownerConn.Close()
		t.Fatal("expected a valid room code")
	}

	guestConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		ownerConn.Close()
		t.Fatalf("failed to connect guest: %v", err)
	}

	joinRequest := ClientMessage{
		Type: "join_room",
		Data: ClientMessageData{
			RoomCode: roomCode,
			UserID:   "user-2",
			Name:     "Guest",
		},
	}

	if err := guestConn.WriteJSON(joinRequest); err != nil {
		guestConn.Close()
		ownerConn.Close()
		t.Fatalf("failed to send join request: %v", err)
	}

	var joinResponse ServerMessage

	if err := guestConn.ReadJSON(&joinResponse); err != nil {
		guestConn.Close()
		ownerConn.Close()
		t.Fatalf("failed to read join response: %v", err)
	}

	if joinResponse.Type != "room_joined" {
		guestConn.Close()
		ownerConn.Close()
		t.Fatalf("expected room_joined, got %q", joinResponse.Type)
	}

	room, exists := roomManager.GetRoomByCode(roomCode)
	if !exists {
		guestConn.Close()
		ownerConn.Close()
		t.Fatal("expected room to exist")
	}

	if room.UserCount() != 2 {
		guestConn.Close()
		ownerConn.Close()
		t.Fatalf("expected 2 users, got %d", room.UserCount())
	}

	if room.ClientCount() != 2 {
		guestConn.Close()
		ownerConn.Close()
		t.Fatalf("expected 2 clients, got %d", room.ClientCount())
	}

	guestConn.Close()

	waitForCondition(t, func() bool {
		room, exists := roomManager.GetRoomByCode(roomCode)
		return exists && room.UserCount() == 1 && room.ClientCount() == 1
	})

	room, exists = roomManager.GetRoomByCode(roomCode)
	if !exists {
		ownerConn.Close()
		t.Fatal("expected room to remain after guest disconnects")
	}

	if room.UserCount() != 1 {
		ownerConn.Close()
		t.Fatalf("expected 1 user after guest disconnect, got %d", room.UserCount())
	}

	if room.ClientCount() != 1 {
		ownerConn.Close()
		t.Fatalf("expected 1 client after guest disconnect, got %d", room.ClientCount())
	}

	ownerConn.Close()

	waitForCondition(t, func() bool {
		_, exists := roomManager.GetRoomByCode(roomCode)
		return !exists
	})

	if _, exists := roomManager.GetRoomByCode(roomCode); exists {
		t.Fatal("expected room to be removed after owner disconnects")
	}
}

func waitForCondition(t *testing.T, condition func() bool) {
	t.Helper()

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return
		}

		select {
		case <-timeout.C:
			t.Fatal("timed out waiting for server state")
		case <-ticker.C:
		}
	}
}

func TestWebSocketGuestCannotControlPlayback(t *testing.T) {
	roomManager := NewRoomManager()
	server := NewWebSocketServer(roomManager)

	httpServer := httptest.NewServer(http.HandlerFunc(server.handler))
	defer httpServer.Close()

	url := "ws" + httpServer.URL[4:]

	ownerConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect owner: %v", err)
	}
	defer ownerConn.Close()

	if err := ownerConn.WriteJSON(ClientMessage{
		Type: "create_room",
		Data: ClientMessageData{
			UserID: "owner-1",
			Name:   "Owner",
		},
	}); err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	var createResponse ServerMessage
	if err := ownerConn.ReadJSON(&createResponse); err != nil {
		t.Fatalf("failed to read create response: %v", err)
	}

	if createResponse.Type != "room_created" {
		t.Fatalf("expected room_created, got %q", createResponse.Type)
	}

	roomData, ok := createResponse.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected room response type: %T", createResponse.Data)
	}

	roomCode, ok := roomData["code"].(string)
	if !ok || roomCode == "" {
		t.Fatal("expected a valid room code")
	}

	guestConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect guest: %v", err)
	}
	defer guestConn.Close()

	if err := guestConn.WriteJSON(ClientMessage{
		Type: "join_room",
		Data: ClientMessageData{
			UserID:   "guest-1",
			Name:     "Guest",
			RoomCode: roomCode,
		},
	}); err != nil {
		t.Fatalf("failed to join room: %v", err)
	}

	var joinResponse ServerMessage
	if err := guestConn.ReadJSON(&joinResponse); err != nil {
		t.Fatalf("failed to read join response: %v", err)
	}

	if joinResponse.Type != "room_joined" {
		t.Fatalf("expected room_joined, got %q", joinResponse.Type)
	}

	var queueResponse ServerMessage
	if err := guestConn.ReadJSON(&queueResponse); err != nil {
		t.Fatalf("failed to read initial queue state: %v", err)
	}

	if queueResponse.Type != "queue_state" {
		t.Fatalf("expected queue_state, got %q", queueResponse.Type)
	}

	var playbackResponse ServerMessage
	if err := guestConn.ReadJSON(&playbackResponse); err != nil {
		t.Fatalf("failed to read initial playback state: %v", err)
	}

	if playbackResponse.Type != "playback_state" {
		t.Fatalf("expected playback_state, got %q", playbackResponse.Type)
	}

	var historyResponse ServerMessage
	if err := guestConn.ReadJSON(&historyResponse); err != nil {
		t.Fatalf("failed to read initial history state: %v", err)
	}

	if historyResponse.Type != "history_state" {
		t.Fatalf("expected history_state, got %q", historyResponse.Type)
	}

	tests := []ClientMessage{
		{
			Type: "play",
			Data: ClientMessageData{
				URL:      "https://youtube.com/watch?v=test",
				Position: 0,
			},
		},
		{
			Type: "pause",
			Data: ClientMessageData{},
		},
		{
			Type: "seek",
			Data: ClientMessageData{
				Position: 30,
			},
		},
		{
			Type: "stop",
			Data: ClientMessageData{},
		},
	}

	for _, request := range tests {
		if err := guestConn.WriteJSON(request); err != nil {
			t.Fatalf("failed to send %s request: %v", request.Type, err)
		}

		var response ServerMessage
		if err := guestConn.ReadJSON(&response); err != nil {
			t.Fatalf("failed to read %s response: %v", request.Type, err)
		}

		if response.Type != "error" {
			t.Fatalf("expected error for guest %s, got %q", request.Type, response.Type)
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("unexpected error response type for %s: %T", request.Type, response.Data)
		}

		message, ok := data["message"].(string)
		if !ok {
			t.Fatalf("expected error message for %s", request.Type)
		}

		if message != ErrNotOwner.Error() {
			t.Fatalf("expected %q for %s, got %q", ErrNotOwner.Error(), request.Type, message)
		}
	}
}
