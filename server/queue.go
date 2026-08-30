package main

import (
	"errors"
	"sync"
)

type Song struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Duration  int64  `json:"duration"`
	Thumbnail string `json:"thumbnail"`
	AudioURL  string `json:"audio_url"`
	AddedBy   string `json:"added_by"`
}

type Queue struct {
	Songs   []*Song
	Current *Song
	mu      sync.RWMutex
}

func NewQueue() *Queue {
	return &Queue{
		Songs: make([]*Song, 0),
	}
}

func (q *Queue) Add(song *Song) error {
	if song == nil {
		return errors.New("song is required")
	}

	if song.ID == "" {
		return errors.New("song id is required")
	}

	if song.URL == "" {
		return errors.New("song url is required")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	q.Songs = append(q.Songs, song)
	return nil
}

func (q *Queue) Remove(songID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Current != nil && q.Current.ID == songID {
		q.Current = nil
		return nil
	}

	for i, song := range q.Songs {
		if song.ID != songID {
			continue
		}

		q.Songs = append(q.Songs[:i], q.Songs[i+1:]...)
		return nil
	}

	return errors.New("song not found")
}

func (q *Queue) Next() *Song {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.Songs) == 0 {
		q.Current = nil
		return nil
	}

	song := q.Songs[0]
	q.Songs = q.Songs[1:]
	q.Current = song

	return song
}

func (q *Queue) RemoveCurrent() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.Current == nil {
		return errors.New("no current song")
	}

	q.Current = nil
	return nil
}

func (q *Queue) CurrentSong() *Song {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.Current
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.Songs = make([]*Song, 0)
	q.Current = nil
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return len(q.Songs)
}

func (q *Queue) List() []*Song {
	q.mu.RLock()
	defer q.mu.RUnlock()

	songs := make([]*Song, len(q.Songs))
	copy(songs, q.Songs)

	return songs
}

func (q *Queue) Snapshot() []*Song {
	q.mu.RLock()
	defer q.mu.RUnlock()

	songs := make([]*Song, 0, len(q.Songs)+1)

	if q.Current != nil {
		songs = append(songs, q.Current)
	}

	songs = append(songs, q.Songs...)
	return songs
}
