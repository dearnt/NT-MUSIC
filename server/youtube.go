package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	ErrYouTubeExtractor = errors.New("youtube extractor failed")
	ErrYouTubeTimeout   = errors.New("youtube extractor timed out")
)

type YouTubeMetadata struct {
	ID        string
	Title     string
	Duration  int64
	URL       string
	Thumbnail string
	AudioURL  string
}

type YouTubeExtractor struct {
	Binary  string
	Timeout time.Duration
}

func NewYouTubeExtractor() *YouTubeExtractor {
	binary := strings.TrimSpace(os.Getenv("YTDLP_BIN"))
	if binary == "" {
		binary = "yt-dlp"
	}

	return &YouTubeExtractor{
		Binary:  binary,
		Timeout: 30 * time.Second,
	}
}

func (e *YouTubeExtractor) command(ctx context.Context, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, e.Binary, args...)
}

func (e *YouTubeExtractor) Extract(ctx context.Context, videoURL string) (*YouTubeMetadata, error) {
	if strings.TrimSpace(videoURL) == "" {
		return nil, ErrInvalidSongURL
	}

	if e == nil {
		return nil, ErrYouTubeExtractor
	}

	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	commonArgs := []string{
		"--extractor-args",
		"youtube:player_client=android_vr",
		"--js-runtimes",
		"deno",
		"--remote-components",
		"ejs:npm",
		"--no-playlist",
		"--no-warnings",
	}

	metadataArgs := append(append([]string{}, commonArgs...),
		"--dump-single-json",
		"--no-download",
		videoURL,
	)

	metadataCmd := e.command(requestCtx, metadataArgs...)

	output, err := metadataCmd.CombinedOutput()
	if err != nil {
		if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrYouTubeTimeout
		}
		log.Printf("yt-dlp metadata error: %s", strings.TrimSpace(string(output)))
		return nil, ErrYouTubeExtractor
	}

	var data struct {
		ID        string  `json:"id"`
		Title     string  `json:"title"`
		Duration  float64 `json:"duration"`
		Webpage   string  `json:"webpage_url"`
		Thumbnail string  `json:"thumbnail"`
	}

	if err := json.Unmarshal(output, &data); err != nil {
		return nil, ErrYouTubeExtractor
	}

	if data.ID == "" || data.Title == "" {
		return nil, ErrYouTubeExtractor
	}

	audioArgs := append(append([]string{}, commonArgs...),
		"-f",
		"bestaudio",
		"-g",
		videoURL,
	)

	audioCmd := e.command(requestCtx, audioArgs...)

	audioOutput, err := audioCmd.CombinedOutput()
	if err != nil {
		if errors.Is(requestCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrYouTubeTimeout
		}
		log.Printf("yt-dlp audio error: %s", strings.TrimSpace(string(audioOutput)))
		return nil, ErrYouTubeExtractor
	}

	audioURL := strings.TrimSpace(strings.Split(string(audioOutput), "\n")[0])
	if audioURL == "" {
		return nil, ErrYouTubeExtractor
	}

	duration := int64(data.Duration)
	if duration < 0 {
		duration = 0
	}

	resultURL := data.Webpage
	if resultURL == "" {
		resultURL = videoURL
	}

	return &YouTubeMetadata{
		ID:        data.ID,
		Title:     data.Title,
		Duration:  duration,
		URL:       resultURL,
		Thumbnail: data.Thumbnail,
		AudioURL:  audioURL,
	}, nil
}

func parseYouTubeDuration(value string) int64 {
	parts := strings.Split(value, ":")

	if len(parts) == 0 {
		return 0
	}

	var total int64

	for _, part := range parts {
		seconds, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0
		}

		total = total*60 + seconds
	}

	return total
}
