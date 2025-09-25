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
	"github.com/oegegr/gophermart/internal/middleware"
	"github.com/oegegr/gophermart/internal/models/accrual"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

type Application struct {
	cfg    *config.Config
	server *http.Server
	accr   *services.AccrualProcessor
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
	jwt := services.NewJWTParser(cfg.JWTSecret)

	userService, err := services.NewUserServiceImpl(s, hashProvider)
	if err != nil {
		return nil, err
	}

	userHandler, err := handlers.NewUserHandler(userService, jwt)
	if err != nil {
		return nil, err
	}

	userLoginProvider := &middleware.AuthContextUserIDPovider{}

	orderService := services.NewOrderServiceImpl(s)

	orderValidator := &services.LunhOrderValidator{}

	orderHandler, err := handlers.NewOrderHandler(orderService, userLoginProvider, jwt, orderValidator)
	if err != nil {
		return nil, err
	}

	withdrawService := services.NewWithdrawServiceImpl(s)

	balanceHandler := handlers.NewBalanceHandler(withdrawService, userLoginProvider, jwt, orderValidator)

	router := NewRouter(userHandler, orderHandler, balanceHandler, jwt)

	server := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	accrualClient, err := accrual.NewClientWithResponses(cfg.AccrualSystemAddress)
	if err != nil {
		return nil, err
	}

	accr := services.NewAccrualProcessor(
		accrualClient,
		s,
		cfg.AccrualInterval,
		10,
		1000,
	)

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
