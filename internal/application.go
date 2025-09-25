package internal

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/oegegr/gophermart/internal/config"
	"github.com/oegegr/gophermart/internal/services"
)

type Application struct {
	cfg    *config.Config
	server *http.Server
	accr   *services.AccrualProcessor
	dbConn *sql.DB
}

func NewApplication(
	cfg *config.Config,
	server *http.Server,
	accr *services.AccrualProcessor,
	dbConn *sql.DB) (*Application, error) {

	return &Application{
		cfg:    cfg,
		server: server,
		dbConn: dbConn,
		accr:   accr,
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Accrual processing starting on %s", app.cfg.RunAddress)
		app.accr.Start(ctx)
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
