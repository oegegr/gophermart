package internal

import (
	"net/http"
	"time"

	"github.com/oegegr/gophermart/internal/config"
	"github.com/oegegr/gophermart/internal/config/db"
	"github.com/oegegr/gophermart/internal/handlers"
	"github.com/oegegr/gophermart/internal/middleware"
	"github.com/oegegr/gophermart/internal/models/accrual"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

type ApplicationBuilder struct {
	cfg *config.Config
}

func NewApplicationBuilder(cfg *config.Config) *ApplicationBuilder {
	return &ApplicationBuilder{cfg}
}

func (b *ApplicationBuilder) Build() (*Application, error) {
	dbConn, err := db.NewDB(b.cfg)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewPGStorage(dbConn)
	if err != nil {
		return nil, err
	}

	hashProvider := services.NewBcryptHashProvider()
	jwt := services.NewJWTParser(b.cfg.JWTSecret)

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
		Addr:         b.cfg.RunAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	accrualClient, err := accrual.NewClientWithResponses(b.cfg.AccrualSystemAddress)
	if err != nil {
		return nil, err
	}

	accr := services.NewAccrualProcessor(
		accrualClient,
		s,
		b.cfg.AccrualInterval,
		10,
		1000,
	)

	return &Application{
		cfg:    b.cfg,
		server: server,
		dbConn: dbConn,
		accr:   accr,
	}, nil
}
