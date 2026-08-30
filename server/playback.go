package main

import (
	"sync"
	"time"
)

type PlaybackState struct {
	SongID   string  `json:"song_id"`
	Playing  bool    `json:"playing"`
	Position float64 `json:"position"`
	Updated  int64   `json:"updated"`
}

type Playback struct {
	mu    sync.RWMutex
	state PlaybackState
}

func NewPlayback() *Playback {
	return &Playback{}
}

func (p *Playback) Play(songID string, position float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if position < 0 {
		position = 0
	}

	p.state = PlaybackState{
		SongID:   songID,
		Playing:  true,
		Position: position,
		Updated:  time.Now().UnixMilli(),
	}
}

func (p *Playback) Pause(position float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state.Playing {
		elapsed := time.Since(time.UnixMilli(p.state.Updated)).Seconds()
		position = p.state.Position + elapsed
	}

	if position < 0 {
		position = 0
	}

	p.state.Playing = false
	p.state.Position = position
	p.state.Updated = time.Now().UnixMilli()
}

func (p *Playback) Seek(position float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if position < 0 {
		position = 0
	}

	p.state.Position = position
	p.state.Updated = time.Now().UnixMilli()
}

func (p *Playback) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.state = PlaybackState{
		Playing: false,
	}
}

func (p *Playback) State() PlaybackState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state := p.state

	if state.Playing {
		elapsed := time.Since(time.UnixMilli(state.Updated)).Seconds()
		state.Position += elapsed
	}

	return state
}
