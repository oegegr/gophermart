package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/storage"
)

func NewOrderHandler(
	service services.OrderService,
	loginProvider UserLoginProvider,
	jwt services.JWTParser,
	validator services.OrderValidator,
) (*OrderHandler, error) {
	return &OrderHandler{
		orderService:   service,
		loginProvider:  loginProvider,
		jwt:            jwt,
		orderValidator: validator,
	}, nil
}

type OrderHandler struct {
	orderService   services.OrderService
	loginProvider  UserLoginProvider
	jwt            services.JWTParser
	orderValidator services.OrderValidator
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login, err := h.loginProvider.Get(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	orderNumber, err := parseOrderNumber(r)
	if err != nil {
		http.Error(w, "Invalid order number: parsing failed", http.StatusUnprocessableEntity)
		return
	}

	ok := h.orderValidator.Validate(orderNumber)
	if !ok {
		http.Error(w, "Invalid order number: validation failed", http.StatusUnprocessableEntity)
		return
	}

	err = h.orderService.UploadOrder(ctx, login, orderNumber)
	if err != nil {
		if errors.Is(err, storage.ErrStorageOrderAlreadyUploadedByOtherUser) {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		if errors.Is(err, storage.ErrStorageOrderAlreadyUploadedByUser) {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login, err := h.loginProvider.Get(ctx)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	orders, err := h.orderService.GetUserOrders(ctx, login)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		log.Printf("error encoding orders: %v", err)
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

}

func parseOrderNumber(r *http.Request) (string, error) {
	body := make([]byte, r.ContentLength)
	if _, err := r.Body.Read(body); err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil

}
