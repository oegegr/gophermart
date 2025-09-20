package handlers

import (
	"github.com/go-chi/jwtauth/v5"

	"github.com/oegegr/gophermart/internal/services"
	"github.com/oegegr/gophermart/internal/models/accrual"
)

type UserHandler struct {
	userService *services.UserService
	tokenAuth   *jwtauth.JWTAuth
}

type OrderHandler struct {
	orderService   *services.OrderService
	balanceService *services.BalanceService
	accrualClient  *accrual.ClientWithResponses
}

type BalanceHandler struct {
	balanceService *services.BalanceService
}

func NewUserHandler(userService *services.UserService, tokenAuth *jwtauth.JWTAuth) *UserHandler {
	return &UserHandler{
		userService: userService,
		tokenAuth:   tokenAuth,
	}
}

func NewOrderHandler(
	orderService *services.OrderService,
	balanceService *services.BalanceService,
	accrualClient *accrual.ClientWithResponses,
) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		balanceService: balanceService,
		accrualClient:  accrualClient,
	}
}

func NewBalanceHandler(balanceService *services.BalanceService) *BalanceHandler {
	return &BalanceHandler{balanceService: balanceService}
}