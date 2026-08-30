package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	root := os.Getenv("NT_MUSIC_FRONTEND")
	if root == "" {
		root = "./frontend"
	}

	port := os.Getenv("CLIENT_PORT")
	if port == "" {
		port = "3000"
	}

	http.Handle("/", http.FileServer(http.Dir(root)))

	log.Printf("NT-MUSIC client starting")
	log.Printf("NT-MUSIC frontend listening on :%s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
