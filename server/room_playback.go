package main

import (
	"errors"
	"sync"
	"time"
)

var ErrQueueEmpty = errors.New("queue is empty")

type RoomPlaybackController struct {
	room     *Room
	mu       sync.Mutex
	onChange func(*Song)
}

func NewRoomPlaybackController(room *Room) *RoomPlaybackController {
	return &RoomPlaybackController{
		room: room,
	}
}

func (c *RoomPlaybackController) SetOnChange(fn func(*Song)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChange = fn
}

func (c *RoomPlaybackController) StartNext() (*Song, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.room == nil {
		return nil, errors.New("room is required")
	}

	song := c.room.Queue.Next()
	if song == nil {
		c.room.Playback.Stop()
		return nil, ErrQueueEmpty
	}

	c.room.Playback.Play(song.ID, 0)

	if c.room.History != nil {
		_ = c.room.History.Add(song)
	}

	return song, nil
}

func (c *RoomPlaybackController) Resume() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.room == nil {
		return errors.New("room is required")
	}

	current := c.room.Queue.CurrentSong()
	if current == nil {
		return ErrQueueEmpty
	}

	state := c.room.Playback.State()
	c.room.Playback.Play(current.ID, state.Position)

	return nil
}

func (c *RoomPlaybackController) StartNextWhenFinished() {
	c.mu.Lock()

	if c.room == nil {
		c.mu.Unlock()
		return
	}

	state := c.room.Playback.State()
	current := c.room.Queue.CurrentSong()

	if current != nil {
		if !state.Playing {
			c.mu.Unlock()
			return
		}

		if current.Duration > 0 && state.Position < float64(current.Duration) {
			c.mu.Unlock()
			return
		}
	}

	next := c.room.Queue.Next()
	if next == nil {
		c.room.Playback.Stop()
		onChange := c.onChange
		c.mu.Unlock()

		if onChange != nil {
			onChange(nil)
		}
		return
	}

	c.room.Playback.Play(next.ID, 0)

	if c.room.History != nil {
		_ = c.room.History.Add(next)
	}

	onChange := c.onChange
	c.mu.Unlock()

	if onChange != nil {
		onChange(next)
	}
}

func (c *RoomPlaybackController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.room == nil {
		return
	}

	c.room.Playback.Stop()
}

func (c *RoomPlaybackController) Run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		c.StartNextWhenFinished()
	}
}
