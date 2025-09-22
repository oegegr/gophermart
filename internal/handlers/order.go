package handlers

import (
	"encoding/json"
	"net/http"
	"io"
	"strings"

	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/middleware"
)

func NewOrderHandler(
	service services.OrderService, 
	loginProvider UserLoginProvider,
	jwt services.JWTParser,
	validator services.OrderValidator,
	) (*OrderHandler, error) {
	return &OrderHandler{
		orderService: service,
		loginProvider: loginProvider,
		jwt: jwt,
		orderValidator: validator,
	}, nil
}

type OrderHandler struct {
	orderService services.OrderService
	loginProvider UserLoginProvider
	jwt          services.JWTParser
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
        http.Error(w, "Invalid order number: parsing failed", http.StatusBadRequest)
		return
	}

	ok := h.orderValidator.Validate(orderNumber)
	if !ok {
        http.Error(w, "Invalid order number: validation failed", http.StatusBadRequest)
        return
	}

    err = h.orderService.UploadOrder(ctx, login, orderNumber)
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

	token, err := h.jwt.CreateNewJWTToken(login)

	middleware.SetAuthCookie(w, token)
	middleware.SetAuthorizationHeader(w, token)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)

}


func parseOrderNumber(r *http.Request) (string, error) {
	body := make([]byte, r.ContentLength)
    if _, err := r.Body.Read(body); err != nil && err != io.EOF {
        return "", err
    }
    return strings.TrimSpace(string(body)), nil

}