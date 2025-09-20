package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/oegegr/gophermart/internal/middlewares/auth"
	"github.com/oegegr/gophermart/internal/models"
)

// UploadOrder обрабатывает загрузку номера заказа
func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	// Проверяем валидность номера заказа (алгоритм Луна)
	if !h.isValidOrderNumber(orderNumber) {
		http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		return
	}

	// Проверяем, не был ли заказ уже загружен этим пользователем
	existingOrder, err := h.orderService.GetOrderByNumber(r.Context(), orderNumber)
	if err == nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
			return
		}
	}

	// Создаем заказ
	order := &models.Order{
		Number:     orderNumber,
		UserID:     userID,
		Status:     "NEW",
		UploadedAt: time.Now(),
	}

	if err := h.orderService.CreateOrder(r.Context(), order); err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetOrders возвращает список заказов пользователя
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

// isValidOrderNumber проверяет номер заказа с помощью алгоритма Луна
func (h *OrderHandler) isValidOrderNumber(number string) bool {
	sum := 0
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}