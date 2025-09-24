package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/oegegr/gophermart/internal/models"
	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/storage"
)

type WithdrawService interface {
	GetUserBalance(ctx context.Context, login string) (*api.Balance, error) 
	WithdrawBalance(ctx context.Context, login string, withdrawalRequest api.WithdrawalRequest) error
	GetUserWithdrawals(ctx context.Context, login string) (*[]api.Withdrawal, error) 
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

func (s *WithdrawServiceImpl) WithdrawBalance(ctx context.Context, login string, withdrawalRequest api.WithdrawalRequest) error {
	processedAt := time.Now()
	w := models.Withdraw{Order: withdrawalRequest.Order, Sum: withdrawalRequest.Sum, Login: login, ProcessedAt: processedAt}
	err := s.storage.WithdrawUserBalance(ctx, w)
	if err != nil {
		log.Printf("failed to get user %s balance: %v", login, err)
		return fmt.Errorf("failed to withdraw user %s balance %v", login, err)
	}
	return nil
}

func (s *WithdrawServiceImpl) GetUserWithdrawals(ctx context.Context, login string) (*[]api.Withdrawal, error) {
	withdrawals, err := s.storage.GetUserWithdrawals(ctx, login)
	if err != nil {
		log.Printf("failed to get user %s withdrawals: %v", login, err)
		return nil, fmt.Errorf("failed to get user %s withdrawals %v", login, err)
	}

	var apiWithdrawals []api.Withdrawal  
	for _, w := range *withdrawals {
		apiWithdrawals = append(apiWithdrawals, api.Withdrawal{Order: &w.Order, ProcessedAt: &w.ProcessedAt, Sum: &w.Sum})
	}

	return &apiWithdrawals, nil 
}
