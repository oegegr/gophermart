package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oegegr/gophermart/internal/handlers"
	app_middleware "github.com/oegegr/gophermart/internal/middleware"
	"github.com/oegegr/gophermart/internal/services"
)

func NewRouter(
	userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler,
	balanceHandler *handlers.BalanceHandler,
	jwt services.JWTParser,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", userHandler.Register)
		r.Post("/api/user/login", userHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(app_middleware.AuthMiddleware(jwt))

		r.Post("/api/user/orders", orderHandler.UploadOrder)
		r.Get("/api/user/orders", orderHandler.GetUserOrders)

		r.Get("/api/user/balance", balanceHandler.GetUserBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.WithdrawUserBalance)
		r.Get("/api/user/withdrawals", balanceHandler.GetUserWithdrawals)
	})

	return r
}
