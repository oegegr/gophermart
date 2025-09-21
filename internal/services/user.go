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
	CreateUser(ctx context.Context, user api.User) (string, error)
	LoginUser(ctx context.Context, user api.User) (string, error)
}

func NewUserServiceImpl(s storage.Storage, hashProvider HashProvider, jwt JWTParser) (*UserServiceImpl, error) {
	return &UserServiceImpl{storage: s, hashProvider: hashProvider, jwt: jwt}, nil
}

type UserServiceImpl struct {
	storage      storage.Storage
	hashProvider HashProvider
	jwt          JWTParser
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, user api.User) (string, error) {
	hash, err := s.hashProvider.GetHash(user.Password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password %w", err)
	}

	u := models.User{
		Login:        user.Login,
		PasswordHash: hash,
	}

	err = s.storage.CreateUser(ctx, u)
	if err != nil {
		return "", err
	} 

	return s.jwt.CreateNewJWTToken(u.Login)
}

func (s *UserServiceImpl) LoginUser(ctx context.Context, user api.User) (string, error) {
	u, err := s.storage.FindUserByLogin(ctx, user.Login)
	if err != nil {
		return "", err
	}

	if !s.hashProvider.CompareHash(user.Password, u.PasswordHash) {
		return "", ErrInvalidCredentials
	}

	return s.jwt.CreateNewJWTToken(u.Login)
}
