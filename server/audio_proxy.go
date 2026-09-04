package main

import (
	"io"
	"net/http"
	"os/exec"
)

func audioProxyHandler(w http.ResponseWriter, r *http.Request) {
	videoURL := r.URL.Query().Get("url")
	if videoURL == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(
		"yt-dlp",
		"--no-playlist",
		"--quiet",
		"--no-warnings",
		"-f", "bestaudio",
		"-o", "-",
		videoURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, "audio stream failed", http.StatusBadGateway)
		return
	}

	if err := cmd.Start(); err != nil {
		http.Error(w, "audio extraction failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "audio/webm")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Accept-Ranges", "none")

	_, _ = io.Copy(w, stdout)
	_ = cmd.Wait()
}
