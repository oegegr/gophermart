package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/oegegr/gophermart/internal/models"
	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("Invalid Credentials")
)

type UserService interface {
	CreateUser(ctx context.Context, user api.User) error
	LoginUser(ctx context.Context, user api.User) error
}

func NewUserServiceImpl(s storage.Storage, hashProvider HashProvider) (*UserServiceImpl, error) {
	return &UserServiceImpl{storage: s, hashProvider: hashProvider}, nil
}

type UserServiceImpl struct {
	storage      storage.Storage
	hashProvider HashProvider
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user api.User) error {
	hash, err := s.hashProvider.GetHash(user.Password)
	if err != nil {
		return fmt.Errorf("Failed to hash password %w", err)
	}

	u := models.User{
		Login:        user.Login,
		PasswordHash: hash,
	}

	return s.storage.CreateUser(ctx, u)
}

func (s *UserServiceImpl) LoginUser(ctx context.Context, user api.User) error {
	u, err := s.storage.FindUserByLogin(ctx, user.Login)
	if err != nil {
		return err
	}

	if !s.hashProvider.CompareHash(user.Password, u.PasswordHash) {
		return ErrInvalidCredentials
	}

	return nil
}
