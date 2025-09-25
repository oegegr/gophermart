package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oegegr/gophermart/internal"
	"github.com/oegegr/gophermart/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application config: %v", err)
	}

	app, err := internal.NewApplicationBuilder(cfg).Build()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Start(ctx); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
