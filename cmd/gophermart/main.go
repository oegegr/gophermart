package main

import (
	"log"

	"github.com/oegegr/gophermart/internal/config"
	"github.com/oegegr/gophermart/internal"
)

func main() {
	cfg := config.LoadConfig()
	srv, err := internal.NewApplication(cfg)

	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}