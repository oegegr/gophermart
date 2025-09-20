package internal 

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"

	"github.com/oegegr/gophermart/internal/config"
	"github.com/oegegr/gophermart/internal/handlers"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
	"github.com/oegegr/gophermart/internal/generated/accrual"
)

type Application struct {
	cfg    *config.Config
	router http.Handler
}

func NewApplication(cfg *config.Config) (*Application, error) {
	store, err := storage.NewPostgresStorage(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w",err)
	}

	userService := services.NewUserService(store)
	orderService := services.NewOrderService(store)
	balanceService := services.NewBalanceService(store)

	accrualClient, err := accrual.NewClientWithResponses(cfg.AccrualSystemAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize accrual client: %w", err)
	}

	// Инициализация JWT аутентификации
	tokenAuth := jwtauth.New("HS256", []byte(cfg.JWTSecret), nil)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService)
	orderHandler := handlers.NewOrderHandler(orderService, accrualClient)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	// Настройка маршрутов
	router := SetupRoutes(userHandler, orderHandler, balanceHandler, accrualClient, tokenAuth)

	return &Application{
		cfg:    cfg,
		router: router,
	}, nil
}

func (s *Application) Start() error {
	server := &http.Server{
		Addr:         s.cfg.RunAddress,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Application starting on %s", s.cfg.RunAddress)
	return server.ListenAndServe()
}