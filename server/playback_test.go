package main

import (
	"testing"
	"time"
)

func TestPlaybackPositionAdvances(t *testing.T) {
	p := NewPlayback()

	p.Play("test-song", 10)

	time.Sleep(100 * time.Millisecond)

	state := p.State()

	if !state.Playing {
		t.Fatal("playback should be playing")
	}

	if state.Position < 10.05 {
		t.Fatalf("position did not advance: got %.3f", state.Position)
	}
}

func TestPlaybackPauseFreezesPosition(t *testing.T) {
	p := NewPlayback()

	p.Play("test-song", 10)

	time.Sleep(100 * time.Millisecond)

	before := p.State().Position

	p.Pause(before)

	time.Sleep(100 * time.Millisecond)

	after := p.State().Position

	if after < before-0.01 || after > before+0.01 {
		t.Fatalf("paused position changed: before=%.3f after=%.3f", before, after)
	}
}

func TestPlaybackSeekWhilePlaying(t *testing.T) {
	p := NewPlayback()

	p.Play("test-song", 10)
	p.Seek(42.5)

	state := p.State()

	if state.Position < 42.5 || state.Position > 42.52 {
		t.Fatalf("expected position near 42.5, got %.3f", state.Position)
	}

	if !state.Playing {
		t.Fatal("seek should preserve playing state")
	}
}

func TestPlaybackSeekClampsNegative(t *testing.T) {
	p := NewPlayback()

	p.Play("test-song", 10)
	p.Seek(-5)

	state := p.State()

	if state.Position < 0 || state.Position > 0.02 {
		t.Fatalf("expected negative seek to clamp near 0, got %.3f", state.Position)
	}
}
