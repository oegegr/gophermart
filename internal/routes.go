package internal 

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"

	"github.com/oegegr/gophermart/internal/handlers"
	"github.com/oegegr/gophermart/internal/middleware/auth"
	"github.com/oegegr/gophermart/internal/models/accrual"
)

func SetupRoutes(
	userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler,
	balanceHandler *handlers.BalanceHandler,
	accrualClient *accrual.Client,
	tokenAuth *jwtauth.JWTAuth,
) http.Handler {
	r := chi.NewRouter()

	// Базовые middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// Публичные маршруты
	r.Group(func(r chi.Router) {
		r.Post("/api/user/register", userHandler.Register)
		r.Post("/api/user/login", userHandler.Login)
	})

	// Защищенные маршруты (требуют аутентификации)
	r.Group(func(r chi.Router) {
		// JWT аутентификация
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(auth.Authenticator)

		// Маршруты для работы с заказами
		r.Route("/api/user/orders", func(r chi.Router) {
			r.Post("/", orderHandler.UploadOrder)
			r.Get("/", orderHandler.GetOrders)
		})

		// Маршруты для работы с балансом
		r.Route("/api/user/balance", func(r chi.Router) {
			r.Get("/", balanceHandler.GetBalance)
			r.Post("/withdraw", balanceHandler.Withdraw)
		})

		// Маршруты для истории списаний
		r.Get("/api/user/withdrawals", balanceHandler.GetWithdrawals)
	})

	return r
}