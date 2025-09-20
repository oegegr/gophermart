package models

import "time"

// User представляет пользователя системы
type User struct {
	ID           string    `db:"id"`
	Login        string    `db:"login"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

// Order представляет заказ пользователя
type Order struct {
	ID         string    `db:"id"`
	Number     string    `db:"number"`
	UserID     string    `db:"user_id"`
	Status     string    `db:"status"`
	Accrual    float64   `db:"accrual"`
	UploadedAt time.Time `db:"uploaded_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// Balance представляет баланс пользователя
type Balance struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Current   float64   `db:"current"`
	Withdrawn float64   `db:"withdrawn"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Withdrawal представляет операцию списания средств
type Withdrawal struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	OrderNumber string    `db:"order_number"`
	Sum         float64   `db:"sum"`
	ProcessedAt time.Time `db:"processed_at"`
}

// AccrualResponse представляет ответ от системы начислений
type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}