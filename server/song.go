package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
)

var (
	ErrInvalidSongURL = errors.New("invalid song url")
	ErrEmptySongURL   = errors.New("song url is required")
	ErrEmptyUserID    = errors.New("user id is required")
)

type SongService struct{}

func NewSongService() *SongService {
	return &SongService{}
}

func (s *SongService) CreateSong(songURL string, userID string) (*Song, error) {
	if strings.TrimSpace(songURL) == "" {
		return nil, ErrEmptySongURL
	}

	if strings.TrimSpace(userID) == "" {
		return nil, ErrEmptyUserID
	}

	normalizedURL, err := normalizeYouTubeURL(songURL)
	if err != nil {
		return nil, ErrInvalidSongURL
	}

	id, err := generateSongID()
	if err != nil {
		return nil, err
	}

	return &Song{
		ID:      id,
		URL:     normalizedURL,
		AddedBy: userID,
	}, nil
}

func normalizeYouTubeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", ErrInvalidSongURL
	}

	host := strings.ToLower(parsed.Hostname())

	switch host {
	case "youtu.be", "www.youtu.be":
		videoID := strings.Trim(parsed.Path, "/")
		if videoID == "" {
			return "", ErrInvalidSongURL
		}

		return "https://www.youtube.com/watch?v=" + url.QueryEscape(videoID), nil

	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com":
		videoID := parsed.Query().Get("v")
		if videoID == "" {
			return "", ErrInvalidSongURL
		}

		return "https://www.youtube.com/watch?v=" + url.QueryEscape(videoID), nil

	default:
		return "", ErrInvalidSongURL
	}
}

func isYouTubeURL(parsed *url.URL) bool {
	host := strings.ToLower(parsed.Hostname())

	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com", "youtu.be", "www.youtu.be":
		return true
	default:
		return false
	}
}

func generateSongID() (string, error) {
	data := make([]byte, 16)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}
