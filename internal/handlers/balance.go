package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

func NewBalanceHandler(
	service services.WithdrawService,
	loginProvider UserLoginProvider,
	jwt services.JWTParser,
	validator services.OrderValidator,
) *BalanceHandler {
	return &BalanceHandler{
		withdrawService: service,
		loginProvider:   loginProvider,
		jwt:             jwt,
		orderValidator:  validator,
	}
}

type BalanceHandler struct {
	withdrawService services.WithdrawService
	loginProvider   UserLoginProvider
	jwt             services.JWTParser
	orderValidator  services.OrderValidator
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
		http.Error(w, "Failed to get user balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		log.Printf("error encoding balance: %v", err)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}
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
		if errors.Is(err, services.ErrServiceInsufficientUserBalance) {
			http.Error(w, "insufficient balance", http.StatusPaymentRequired)
			return
		}
		http.Error(w, "Failed to withdraw balance", http.StatusInternalServerError)
		return
	}

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
		if errors.Is(err, storage.ErrStorageWithdrawalsNotFound) {
			http.Error(w, "Withdrawals not found", http.StatusNoContent)
			return
		}
		http.Error(w, "failed to get user withdrawals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		log.Printf("error encoding withdrawals: %v", err)
		http.Error(w, "failed to get user withdrawals", http.StatusInternalServerError)
		return
	}
}
