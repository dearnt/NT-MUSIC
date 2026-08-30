package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	roomManager := NewRoomManager()
	websocketServer := NewWebSocketServer(roomManager)

	http.Handle("/", http.FileServer(http.Dir("../client/frontend")))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/audio", audioProxyHandler)
	http.HandleFunc("/ws", websocketServer.handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8765"
	}
	addr := ":" + port
	log.Printf("NT-MUSIC server listening on %s", addr)
	log.Printf("Rooms initialized: %d", len(roomManager.Rooms))

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "nt-music",
	})
}
