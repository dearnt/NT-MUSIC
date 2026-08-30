package main

import (
	"errors"
	"sync"
)

const maxHistorySongs = 100

var ErrHistorySongNotFound = errors.New("song not found in history")

type SongHistory struct {
	Songs []*Song
	mu    sync.RWMutex
}

func NewSongHistory() *SongHistory {
	return &SongHistory{
		Songs: make([]*Song, 0),
	}
}

func (h *SongHistory) Add(song *Song) error {
	if song == nil {
		return errors.New("song is required")
	}

	if song.ID == "" {
		return errors.New("song id is required")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	copySong := *song

	for i, existing := range h.Songs {
		if existing.ID == song.ID {
			h.Songs = append(h.Songs[:i], h.Songs[i+1:]...)
			break
		}
	}

	h.Songs = append([]*Song{&copySong}, h.Songs...)

	if len(h.Songs) > maxHistorySongs {
		h.Songs = h.Songs[:maxHistorySongs]
	}

	return nil
}

func (h *SongHistory) Get(songID string) (*Song, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, song := range h.Songs {
		if song.ID == songID {
			copySong := *song
			return &copySong, nil
		}
	}

	return nil, ErrHistorySongNotFound
}

func (h *SongHistory) List() []*Song {
	h.mu.RLock()
	defer h.mu.RUnlock()

	songs := make([]*Song, 0, len(h.Songs))

	for _, song := range h.Songs {
		copySong := *song
		songs = append(songs, &copySong)
	}

	return songs
}

func (h *SongHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Songs = make([]*Song, 0)
}

func (h *SongHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.Songs)
}
