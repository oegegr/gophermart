package services

import (
	"context"
	"fmt"
	"log"

	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/storage"
)

type WithdrawService interface {
	GetUserBalance(ctx context.Context, login string) (*api.Balance, error) 
}

func NewWithdrawServiceImpl(storage storage.Storage) *WithdrawServiceImpl {
	return &WithdrawServiceImpl{
		storage: storage,
	}
}

type WithdrawServiceImpl struct {
	storage storage.Storage
}


func (s *WithdrawServiceImpl) GetUserBalance(ctx context.Context, login string) (*api.Balance, error) {
	b, err := s.storage.GetUserBalance(ctx, login)
	if err != nil {
		log.Printf("failed to get user %s balance: %v", login, err)
		return nil, fmt.Errorf("failed to get user %s balance %v", login, err)
	}

	return &api.Balance{Current: &b.Current, Withdrawn: &b.Withdraw}, nil
}
