package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestPlaybackSync(t *testing.T) {
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
			UserID: "playback-owner",
			Name:   "Playback Owner",
		},
	}); err != nil {
		t.Fatalf("create room failed: %v", err)
	}

	created, err := owner.Read(ctx)
	if err != nil {
		t.Fatalf("reading room_created failed: %v", err)
	}

	if created.Type != "room_created" {
		t.Fatalf("expected room_created, got %q", created.Type)
	}

	var room RoomResponse
	raw, err := json.Marshal(created.Data)
	if err != nil {
		t.Fatalf("marshal room data failed: %v", err)
	}

	if err := json.Unmarshal(raw, &room); err != nil {
		t.Fatalf("decode room data failed: %v", err)
	}

	if err := guest.Connect(ctx); err != nil {
		t.Fatalf("guest connect failed: %v", err)
	}

	if err := guest.Send(ClientMessage{
		Type: "join_room",
		Data: ClientMessageData{
			UserID:   "playback-guest",
			Name:     "Playback Guest",
			RoomCode: room.Code,
		},
	}); err != nil {
		t.Fatalf("join room failed: %v", err)
	}

	joined, err := guest.Read(ctx)
	if err != nil {
		t.Fatalf("reading room_joined failed: %v", err)
	}

	if joined.Type != "room_joined" {
		t.Fatalf("expected room_joined, got %q", joined.Type)
	}

	queue, err := guest.Read(ctx)
	if err != nil {
		t.Fatalf("reading queue_state failed: %v", err)
	}

	if queue.Type != "queue_state" {
		t.Fatalf("expected queue_state, got %q", queue.Type)
	}

	if err := owner.Send(ClientMessage{
		Type: "play",
		Data: ClientMessageData{
			URL: "test-song-sync",
		},
	}); err != nil {
		t.Fatalf("play failed: %v", err)
	}

	ownerPlay, err := owner.Read(ctx)
	if err != nil {
		t.Fatalf("owner play read failed: %v", err)
	}

	guestPlay, err := guest.Read(ctx)
	if err != nil {
		t.Fatalf("guest play read failed: %v", err)
	}

	if ownerPlay.Type != "playback_state" {
		t.Fatalf("owner expected playback_state, got %q", ownerPlay.Type)
	}

	if guestPlay.Type != "playback_state" {
		t.Fatalf("guest expected playback_state, got %q", guestPlay.Type)
	}

	var ownerState PlaybackState
	var guestState PlaybackState

	raw, err = json.Marshal(ownerPlay.Data)
	if err != nil {
		t.Fatalf("marshal owner playback failed: %v", err)
	}

	if err := json.Unmarshal(raw, &ownerState); err != nil {
		t.Fatalf("decode owner playback failed: %v", err)
	}

	raw, err = json.Marshal(guestPlay.Data)
	if err != nil {
		t.Fatalf("marshal guest playback failed: %v", err)
	}

	if err := json.Unmarshal(raw, &guestState); err != nil {
		t.Fatalf("decode guest playback failed: %v", err)
	}

	if ownerState.SongID != "test-song-sync" {
		t.Fatalf("expected song test-song-sync, got %q", ownerState.SongID)
	}

	if !ownerState.Playing {
		t.Fatal("owner playback should be playing")
	}

	if guestState.SongID != ownerState.SongID {
		t.Fatalf("song mismatch: owner=%q guest=%q", ownerState.SongID, guestState.SongID)
	}

	if guestState.Playing != ownerState.Playing {
		t.Fatalf("playing mismatch: owner=%v guest=%v", ownerState.Playing, guestState.Playing)
	}

	if err := owner.Send(ClientMessage{
		Type: "seek",
		Data: ClientMessageData{
			Position: 42.5,
		},
	}); err != nil {
		t.Fatalf("seek failed: %v", err)
	}

	ownerSeek, err := owner.Read(ctx)
	if err != nil {
		t.Fatalf("owner seek read failed: %v", err)
	}

	guestSeek, err := guest.Read(ctx)
	if err != nil {
		t.Fatalf("guest seek read failed: %v", err)
	}

	if ownerSeek.Type != "playback_state" {
		t.Fatalf("owner expected playback_state after seek, got %q", ownerSeek.Type)
	}

	if guestSeek.Type != "playback_state" {
		t.Fatalf("guest expected playback_state after seek, got %q", guestSeek.Type)
	}

	var ownerSeekState PlaybackState
	var guestSeekState PlaybackState

	raw, err = json.Marshal(ownerSeek.Data)
	if err != nil {
		t.Fatalf("marshal owner seek failed: %v", err)
	}

	if err := json.Unmarshal(raw, &ownerSeekState); err != nil {
		t.Fatalf("decode owner seek failed: %v", err)
	}

	raw, err = json.Marshal(guestSeek.Data)
	if err != nil {
		t.Fatalf("marshal guest seek failed: %v", err)
	}

	if err := json.Unmarshal(raw, &guestSeekState); err != nil {
		t.Fatalf("decode guest seek failed: %v", err)
	}

	const expectedSeekPosition = 42.5
	const seekTolerance = 0.01

	if abs(ownerSeekState.Position-expectedSeekPosition) > seekTolerance {
		t.Fatalf("owner expected position near %.2f, got %v", expectedSeekPosition, ownerSeekState.Position)
	}

	if abs(guestSeekState.Position-expectedSeekPosition) > seekTolerance {
		t.Fatalf("guest expected position near %.2f, got %v", expectedSeekPosition, guestSeekState.Position)
	}

	if ownerSeekState.Playing != guestSeekState.Playing {
		t.Fatalf("seek playing mismatch: owner=%v guest=%v", ownerSeekState.Playing, guestSeekState.Playing)
	}

	if err := owner.Send(ClientMessage{
		Type: "pause",
		Data: ClientMessageData{},
	}); err != nil {
		t.Fatalf("pause failed: %v", err)
	}

	ownerPause, err := owner.Read(ctx)
	if err != nil {
		t.Fatalf("owner pause read failed: %v", err)
	}

	guestPause, err := guest.Read(ctx)
	if err != nil {
		t.Fatalf("guest pause read failed: %v", err)
	}

	if ownerPause.Type != "playback_state" {
		t.Fatalf("owner expected playback_state after pause, got %q", ownerPause.Type)
	}

	if guestPause.Type != "playback_state" {
		t.Fatalf("guest expected playback_state after pause, got %q", guestPause.Type)
	}

	t.Log("Playback synchronization OK")
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
