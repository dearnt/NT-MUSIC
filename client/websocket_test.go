package main

import (
	"context"
	"testing"
	"time"
)

func TestWSClientConnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client := NewWSClient("ws://127.0.0.1:8765/ws")
	defer client.Close()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Fatal("client should be connected")
	}
}
