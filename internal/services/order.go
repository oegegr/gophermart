package services

import (
	"context"
	"time"

	"github.com/oegegr/gophermart/internal/models"
	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/storage"
)

type OrderService interface {
	UploadOrder(ctx context.Context, login string, orderNumber string) error
	GetUserOrders(ctx context.Context, login string) ([]api.Order, error)
}

func NewOrderServiceImpl(storage storage.Storage) *OrderServiceImpl {
	return &OrderServiceImpl{
		storage: storage,
	}
}

type OrderServiceImpl struct {
	storage storage.Storage
}

func (s *OrderServiceImpl) UploadOrder(ctx context.Context, login string, orderNumber string) error {
	accrual := float32(0.0)
	uploadedAt := time.Now()
	order := models.Order{
		Number: orderNumber,
		Login: login,
		Accrual: accrual,
		Status: string(api.REGISTERED),
		UploadedAt: uploadedAt,
	}

	return s.storage.CreateOrder(ctx, order)
}

func (s *OrderServiceImpl) GetUserOrders(ctx context.Context, login string) ([]api.Order, error) {
	orders, err := s.storage.FindOrdersByUser(ctx, login)
	if err != nil {
		return nil, err
	}

	var apiOrders []api.Order
	for _, o := range orders {
		apiOrders = append(apiOrders, o.ToApi(o))
	} 
	return apiOrders, nil
}
