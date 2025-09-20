package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

func NewUserHandler(service services.UserService) (*UserHandler, error) {
	return &UserHandler{
		userService: service,
	}, nil
}

type UserHandler struct {
	userService services.UserService
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user api.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Login == "" || user.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	if err := h.userService.CreateUser(r.Context(), user); err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var user api.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Login == "" || user.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	if err := h.userService.LoginUser(r.Context(), user); err != nil {
		if errors.Is(err, storage.ErrStorageUserNotFound) {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
