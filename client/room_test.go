package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestCreateRoom(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := NewWSClient("ws://127.0.0.1:8765/ws")
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	err := client.Send(ClientMessage{
		Type: "create_room",
		Data: ClientMessageData{
			UserID: "client-test-user",
			Name:   "Client Test",
		},
	})
	if err != nil {
		t.Fatalf("send create_room failed: %v", err)
	}

	message, err := client.Read(ctx)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}

	if message.Type != "room_created" {
		t.Fatalf("expected room_created, got %q", message.Type)
	}

	data, err := json.Marshal(message.Data)
	if err != nil {
		t.Fatalf("marshal response data failed: %v", err)
	}

	var room RoomResponse
	if err := json.Unmarshal(data, &room); err != nil {
		t.Fatalf("decode room response failed: %v", err)
	}

	if room.ID == "" {
		t.Fatal("room id is empty")
	}

	if room.Code == "" {
		t.Fatal("room code is empty")
	}

	if room.OwnerID != "client-test-user" {
		t.Fatalf("expected owner client-test-user, got %q", room.OwnerID)
	}

	if room.UserCount != 1 {
		t.Fatalf("expected user count 1, got %d", room.UserCount)
	}
}
