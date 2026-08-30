package main

import (
	"encoding/json"
	"testing"
)

func TestClientMessageProtocol(t *testing.T) {
	raw := []byte(`{
	"type":"join_room",
	"data":{
	"user_id":"user-2",
	"name":"Guest",
	"room_code":"ABC123"
}
}`)

	var message ClientMessage

	if err := json.Unmarshal(raw, &message); err != nil {
		t.Fatalf("failed to decode client message: %v", err)
	}

	if message.Type != "join_room" {
		t.Fatalf("expected join_room, got %q", message.Type)
	}

	if message.Data.UserID != "user-2" {
		t.Fatalf("expected user-2, got %q", message.Data.UserID)
	}

	if message.Data.RoomCode != "ABC123" {
		t.Fatalf("expected ABC123, got %q", message.Data.RoomCode)
	}
}

func TestServerMessageProtocol(t *testing.T) {
	message := ServerMessage{
		Type: "room_joined",
		Data: RoomResponse{
			ID:        "room-id",
			Code:      "ABC123",
			OwnerID:   "owner",
			UserCount: 2,
		},
	}

	raw, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to encode server message: %v", err)
	}

	var decoded map[string]interface{}

	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("failed to decode server message: %v", err)
	}

	if decoded["type"] != "room_joined" {
		t.Fatalf("expected room_joined, got %v", decoded["type"])
	}

	data, ok := decoded["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", decoded["data"])
	}

	if data["code"] != "ABC123" {
		t.Fatalf("expected ABC123, got %v", data["code"])
	}

	if data["user_count"] != float64(2) {
		t.Fatalf("expected user count 2, got %v", data["user_count"])
	}
}

func TestClientMessageJSONTags(t *testing.T) {
	message := ClientMessage{
		Type: "add_song",
		Data: ClientMessageData{
			UserID: "user-1",
			Name:   "Owner",
			URL:    "https://youtube.com/watch?v=test",
		},
	}

	raw, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("failed to encode message: %v", err)
	}

	expected := `{"type":"add_song","data":{"user_id":"user-1","name":"Owner","url":"https://youtube.com/watch?v=test"}}`

	if string(raw) != expected {
		t.Fatalf("unexpected JSON: %s", raw)
	}
}
