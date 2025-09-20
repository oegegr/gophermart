package internal

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/oegegr/gophermart/internal/handlers"
)

func NewRouter(
	userHandler *handlers.UserHandler) http.Handler {
	r := chi.NewRouter()

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

	return r
}
