package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func readServerMessage(t *testing.T, client *WSClient) ServerMessage {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := make(chan struct {
		message ServerMessage
		err     error
	}, 1)

	go func() {
		message, err := client.Read(ctx)
		result <- struct {
			message ServerMessage
			err     error
		}{message, err}
	}()

	select {
	case result := <-result:
		if result.err != nil {
			t.Fatalf("receive failed: %v", result.err)
		}
		return result.message
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server message")
		return ServerMessage{}
	}
}

func decodeRoomResponse(t *testing.T, data any) RoomResponse {
	t.Helper()

	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal room data: %v", err)
	}

	var room RoomResponse
	if err := json.Unmarshal(raw, &room); err != nil {
		t.Fatalf("failed to decode room response: %v", err)
	}

	return room
}

func TestCreateAndJoinRoom(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	owner := NewWSClient("ws://127.0.0.1:8765/ws")
	defer owner.Close()

	guest := NewWSClient("ws://127.0.0.1:8765/ws")
	defer guest.Close()

	if err := owner.Connect(ctx); err != nil {
		t.Fatalf("owner connect failed: %v", err)
	}

	if err := owner.Send(ClientMessage{
		Type: "create_room",
		Data: ClientMessageData{
			UserID: "room-owner-test",
			Name:   "Room Owner",
		},
	}); err != nil {
		t.Fatalf("create room send failed: %v", err)
	}

	createdMessage := readServerMessage(t, owner)

	if createdMessage.Type != "room_created" {
		t.Fatalf("expected room_created, got %q", createdMessage.Type)
	}

	created := decodeRoomResponse(t, createdMessage.Data)

	if created.ID == "" {
		t.Fatal("room id is empty")
	}

	if created.Code == "" {
		t.Fatal("room code is empty")
	}

	if created.OwnerID != "room-owner-test" {
		t.Fatalf("expected owner room-owner-test, got %q", created.OwnerID)
	}

	if created.UserCount != 1 {
		t.Fatalf("expected user count 1, got %d", created.UserCount)
	}

	if err := guest.Connect(ctx); err != nil {
		t.Fatalf("guest connect failed: %v", err)
	}

	if err := guest.Send(ClientMessage{
		Type: "join_room",
		Data: ClientMessageData{
			UserID:   "room-guest-test",
			Name:     "Room Guest",
			RoomCode: created.Code,
		},
	}); err != nil {
		t.Fatalf("join room send failed: %v", err)
	}

	joinedMessage := readServerMessage(t, guest)

	if joinedMessage.Type != "room_joined" {
		t.Fatalf("expected room_joined, got %q", joinedMessage.Type)
	}

	joined := decodeRoomResponse(t, joinedMessage.Data)

	if joined.ID != created.ID {
		t.Fatalf("room id mismatch: created=%q joined=%q", created.ID, joined.ID)
	}

	if joined.Code != created.Code {
		t.Fatalf("room code mismatch: created=%q joined=%q", created.Code, joined.Code)
	}

	if joined.OwnerID != "room-owner-test" {
		t.Fatalf("owner changed: got %q", joined.OwnerID)
	}

	if joined.UserCount != 2 {
		t.Fatalf("expected user count 2, got %d", joined.UserCount)
	}

	queueMessage := readServerMessage(t, guest)

	if queueMessage.Type != "queue_state" {
		t.Fatalf("expected queue_state after room_joined, got %q", queueMessage.Type)
	}
}
