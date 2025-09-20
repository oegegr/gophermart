package internal

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/oegegr/gophermart/internal/config"
	"github.com/oegegr/gophermart/internal/config/db"
	"github.com/oegegr/gophermart/internal/handlers"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

type Application struct {
	cfg    *config.Config
	server *http.Server
	dbConn *sql.DB
}

func NewApplication(cfg *config.Config) (*Application, error) {
	dbConn, err := db.NewDB(cfg)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewPGStorage(dbConn)
	if err != nil {
		return nil, err
	}

	hashProvider := services.NewBcryptHashProvider()

	userService, err := services.NewUserServiceImpl(s, hashProvider)
	if err != nil {
		return nil, err
	}

	userHandler, err := handlers.NewUserHandler(userService)
	if err != nil {
		return nil, err
	}

	router := NewRouter(userHandler)

	server := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Application{
		cfg:    cfg,
		server: server,
		dbConn: dbConn,
	}, nil
}

func (app *Application) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Printf("HTTP server starting on %s", app.cfg.RunAddress)
		if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := app.dbConn.Close(); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	wg.Wait()

	return nil
}
