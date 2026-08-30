package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func audioProxyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawURL := strings.TrimSpace(r.URL.Query().Get("url"))
	if rawURL == "" {
		http.Error(w, "missing youtube url", http.StatusBadRequest)
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil || !isYouTubeURL(u) {
		http.Error(w, "invalid youtube url", http.StatusBadRequest)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Accept-Ranges", "none")

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "audio/webm")
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	tempDir, err := os.MkdirTemp("", "nt-music-audio-*")
	if err != nil {
		http.Error(w, "failed to create temporary directory", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	inputFile := filepath.Join(tempDir, "audio.webm")

	ytdlp := exec.CommandContext(
		ctx,
		"/tmp/yt-dlp-venv/bin/yt-dlp",
		"--extractor-args",
		"youtube:player_client=visionos",
		"--js-runtimes",
		"deno",
		"--remote-components",
		"ejs:npm",
		"--no-playlist",
		"--no-warnings",
		"-f",
		"bestaudio",
		"-o",
		inputFile,
		rawURL,
	)

	output, err := ytdlp.CombinedOutput()
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("audio extraction failed: %v: %s", err, strings.TrimSpace(string(output))),
			http.StatusBadGateway,
		)
		return
	}

	if _, err := os.Stat(inputFile); err != nil {
		http.Error(w, "audio file was not created", http.StatusBadGateway)
		return
	}

	ffmpeg := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel",
		"error",
		"-i",
		inputFile,
		"-vn",
		"-c:a",
		"libopus",
		"-b:a",
		"128k",
		"-f",
		"webm",
		"pipe:1",
	)

	ffmpegOutput, err := ffmpeg.Output()
	if err != nil {
		http.Error(w, fmt.Sprintf("audio conversion failed: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "audio/webm")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(ffmpegOutput)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(ffmpegOutput)
}
