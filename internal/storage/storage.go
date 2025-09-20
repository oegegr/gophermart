package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/oegegr/gophermart/internal/models"
)

var (
	ErrStorageUserAlreadyExists = errors.New("user already exists")
	ErrStorageUserNotFound      = errors.New("user not found")
)

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	FindUserByLogin(ctx context.Context, login string) (*models.User, error)
}

func NewPGStorage(dbConn *sql.DB) (*PGStorage, error) {
	return &PGStorage{db: dbConn}, nil
}

type PGStorage struct {
	db *sql.DB
}

func (s *PGStorage) CreateUser(ctx context.Context, user models.User) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO users (login, hash) VALUES ($1, $2)")
	if err != nil {
		log.Printf("sql request validation error: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Login, user.PasswordHash)

	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return ErrStorageUserAlreadyExists
		}

		log.Printf("sql request execution error: %v", err)
		err := tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *PGStorage) FindUserByLogin(ctx context.Context, login string) (*models.User, error) {
	stmt, err := s.db.Prepare("SELECT login, hash FROM users WHERE login = $1")
	if err != nil {
		log.Printf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var user models.User
	err = stmt.QueryRow(login).Scan(&user.Login, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user not found %s", user.Login)
			return nil, ErrStorageUserNotFound
		}
		log.Printf("sql execution error: %v", err)
		return nil, err
	}

	return &user, nil
}
