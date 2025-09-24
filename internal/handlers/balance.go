package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/oegegr/gophermart/internal/middleware"
	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/services"
)

func NewBalanceHandler(
	service services.WithdrawService, 
	loginProvider UserLoginProvider,
	jwt services.JWTParser,
	validator services.OrderValidator,
	) (*BalanceHandler, error) {
	return &BalanceHandler{
		withdrawService: service,
		loginProvider: loginProvider,
		jwt: jwt,
		orderValidator: validator,
	}, nil
}

type BalanceHandler struct {
	withdrawService services.WithdrawService
	loginProvider UserLoginProvider
	jwt          services.JWTParser
	orderValidator services.OrderValidator
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login, err := h.loginProvider.Get(ctx) 
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

    balance, err := h.withdrawService.GetUserBalance(ctx, login)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.jwt.CreateNewJWTToken(login)

	middleware.SetAuthCookie(w, token)
	middleware.SetAuthorizationHeader(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) WithdrawUserBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login, err := h.loginProvider.Get(ctx) 
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var withdrawRequest api.WithdrawalRequest

	if err := json.NewDecoder(r.Body).Decode(&withdrawRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

    err = h.withdrawService.WithdrawBalance(ctx, login, withdrawRequest)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.jwt.CreateNewJWTToken(login)

	middleware.SetAuthCookie(w, token)
	middleware.SetAuthorizationHeader(w, token)

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login, err := h.loginProvider.Get(ctx) 
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

    withdrawals, err := h.withdrawService.GetUserWithdrawals(ctx, login)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.jwt.CreateNewJWTToken(login)

	middleware.SetAuthCookie(w, token)
	middleware.SetAuthorizationHeader(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}