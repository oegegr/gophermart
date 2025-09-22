package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/oegegr/gophermart/internal/models"
)

var (
	ErrStorageUserAlreadyExists = errors.New("user already exists")
	ErrStorageUserNotFound      = errors.New("user not found")
	ErrStorageOrdersNotFound      = errors.New("orders not found")
)

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	FindUserByLogin(ctx context.Context, login string) (*models.User, error)
	CreateOrder(ctx context.Context, order models.Order) error
	FindOrdersByUser(ctx context.Context, login string) ([]models.Order, error)
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

func (s *PGStorage) CreateOrder(ctx context.Context, order models.Order) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO orders (order_number, accrual, status, uploaded_at, user_id) SELECT $1, $2, $3, $4, u.id FROM users u WHERE u.login = $5")
	if err != nil {
		log.Printf("sql request validation error: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(order.Number, order.Accrual, order.Status, order.UploadedAt, order.Login)

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

func (s *PGStorage) FindOrdersByUser(ctx context.Context, login string) ([]models.Order, error) {
	stmt, err := s.db.Prepare("SELECT order_number, accrual, status, uploaded_at FROM orders JOIN  users ON orders.user_id = users.id WHERE users.login = $1")
	if err != nil {
		log.Printf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var orders []models.Order
	rows, err := stmt.Query(login)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("orders not found for user %s", login)
			return nil, ErrStorageOrdersNotFound 
		}
		log.Printf("sql execution error: %v", err)
		return nil, err
	}

	for rows.Next() {
		var order models.Order
		rows.Scan(&order.Number, &order.Accrual, &order.Status, &order.UploadedAt)
		order.Login = login
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row deserialization error %w", err)
	}

	return orders, nil
}