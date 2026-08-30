package main

import (
	"context"
	"errors"
)

var ErrSongMetadata = errors.New("song metadata extraction failed")

type SongMetadataService struct {
	extractor *YouTubeExtractor
}

func NewSongMetadataService(extractor *YouTubeExtractor) *SongMetadataService {
	if extractor == nil {
		extractor = NewYouTubeExtractor()
	}

	return &SongMetadataService{
		extractor: extractor,
	}
}

func (s *SongMetadataService) Enrich(ctx context.Context, song *Song) error {
	if song == nil {
		return ErrSongMetadata
	}

	if s == nil || s.extractor == nil {
		return ErrSongMetadata
	}

	metadata, err := s.extractor.Extract(ctx, song.URL)
	if err != nil {
		return err
	}

	if metadata == nil {
		return ErrSongMetadata
	}

	song.Title = metadata.Title
	song.Duration = metadata.Duration
	song.Thumbnail = metadata.Thumbnail
	song.AudioURL = metadata.AudioURL
	println("AUDIO URL:", song.AudioURL)

	return nil
}
