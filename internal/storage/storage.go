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
	ErrStorageUserAlreadyExists               = errors.New("user already exists")
	ErrStorageUserNotFound                    = errors.New("user not found")
	ErrStorageOrdersNotFound                  = errors.New("orders not found")
	ErrStorageOrderAlreadyUploadedByUser      = errors.New("order already uploaded by user")
	ErrStorageOrderAlreadyUploadedByOtherUser = errors.New("order already uploaded by other user")
	ErrStorageWithdrawalsNotFound             = errors.New("withdrawals not found")
)

type Storage interface {
	CreateUser(ctx context.Context, user models.User) error
	FindUserByLogin(ctx context.Context, login string) (*models.User, error)
	CreateOrder(ctx context.Context, order models.Order) error
	FindOrdersByUser(ctx context.Context, login string) ([]models.Order, error)
	FindOrdersByStatus(ctx context.Context, status string) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, status string, number string, accrual float32) error
	UpdateUserBalance(ctx context.Context, login string, accrual float32) error
	GetUserBalance(ctx context.Context, login string) (*models.Balance, error)
	WithdrawUserBalance(ctx context.Context, withdraw models.Withdraw) error
	GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdraw, error)
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

	_, err = stmt.ExecContext(ctx, user.Login, user.PasswordHash)

	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return ErrStorageUserAlreadyExists
		}

		log.Printf("sql request execution error: %v", err)
		rollbackTransaction(tx)
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

	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM orders WHERE order_number = $1 AND user_id = (SELECT id FROM users WHERE login = $2))", order.Number, order.Login).Scan(&exists)
	if err != nil {
		log.Printf("sql request execution error: %v", err)
		return err
	}
	if exists {
		return ErrStorageOrderAlreadyUploadedByUser
	}

	err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM orders WHERE order_number = $1)", order.Number).Scan(&exists)
	if err != nil {
		log.Printf("sql request execution error: %v", err)
		return err
	}
	if exists {
		return ErrStorageOrderAlreadyUploadedByOtherUser
	}

	stmt, err := tx.Prepare("INSERT INTO orders (order_number, accrual, status, uploaded_at, user_id) SELECT $1, $2, $3, $4, u.id FROM users u WHERE u.login = $5")
	if err != nil {
		log.Printf("sql request validation error: %v", err)
		rollbackTransaction(tx)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(order.Number, order.Accrual, order.Status, order.UploadedAt, order.Login)
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return ErrStorageUserAlreadyExists
		}
		log.Printf("sql request execution error: %v", err)
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
		if err := rows.Scan(&order.Number, &order.Accrual, &order.Status, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("row deserialization error %w", err)
		}
		order.Login = login
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row deserialization error %w", err)
	}

	return orders, nil
}

func (s *PGStorage) FindOrdersByStatus(ctx context.Context, status string) ([]models.Order, error) {
	stmt, err := s.db.Prepare("SELECT order_number, accrual, status, uploaded_at, users.login FROM orders JOIN  users ON orders.user_id = users.id WHERE status = $1")
	if err != nil {
		log.Printf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var orders []models.Order
	rows, err := stmt.Query(status)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("orders with status %s not found", status)
			return nil, ErrStorageOrdersNotFound
		}
		log.Printf("sql execution error: %v", err)
		return nil, err
	}

	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.Accrual, &order.Status, &order.UploadedAt, &order.Login); err != nil {
			return nil, fmt.Errorf("row deserialization error %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row deserialization error %w", err)
	}

	return orders, nil
}

func (s *PGStorage) UpdateOrderStatus(ctx context.Context, status string, number string, accrual float32) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE order_number = $3")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, status, accrual, number)
	if err != nil {
		log.Printf("sql execution error: %v", err)
		rollbackTransaction(tx)
		return err
	}

	return tx.Commit()
}

func (s *PGStorage) UpdateUserBalance(ctx context.Context, login string, accrual float32) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, "UPDATE users SET balance = balance + $1 WHERE login = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, accrual, login)
	if err != nil {
		log.Printf("sql execution error: %v", err)
		rollbackTransaction(tx)
		return err
	}

	return tx.Commit()
}

func (s *PGStorage) GetUserBalance(ctx context.Context, login string) (*models.Balance, error) {
	stmt, err := s.db.Prepare("SELECT u.balance, COALESCE(SUM(w.withdraw), 0) AS total_withdrawn FROM users u LEFT JOIN withdrawals w ON w.user_id = u.id WHERE  u.login = $1 GROUP BY u.balance")
	if err != nil {
		log.Printf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var balance models.Balance
	err = stmt.QueryRow(login).Scan(&balance.Current, &balance.Withdraw)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("balance for user %s not found", login)
			return nil, ErrStorageUserNotFound
		}
		log.Printf("sql execution error: %v", err)
		return nil, err
	}
	return &balance, nil
}

func (s *PGStorage) WithdrawUserBalance(ctx context.Context, withdraw models.Withdraw) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer rollbackTransaction(tx)

	updateBalance, err := tx.Prepare("UPDATE users SET balance = balance - $1 WHERE login = $2 AND balance >= $1")
	if err != nil {
		log.Printf("sql request validation error: %v", err)
		return err
	}
	defer updateBalance.Close()

	insertWithdraw, err := tx.Prepare("INSERT INTO withdrawals (order_number, withdraw, processed_at, user_id) SELECT $1, $2, $3, u.id FROM users u WHERE u.login = $4")
	if err != nil {
		log.Printf("sql request validation error: %v", err)
		return err
	}
	defer insertWithdraw.Close()

	result, err := updateBalance.ExecContext(ctx, withdraw.Sum, withdraw.Login)

	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to withdraw balance: %w", err)
	}

	_, err = insertWithdraw.ExecContext(ctx, withdraw.Order, withdraw.Sum, withdraw.ProcessedAt, withdraw.Login)
	if err != nil {
		return fmt.Errorf("failed to insert withdrawal record: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PGStorage) GetUserWithdrawals(ctx context.Context, login string) (*[]models.Withdraw, error) {
	stmt, err := s.db.Prepare("SELECT order_number, withdraw, processed_at FROM withdrawals JOIN  users ON withdrawals.user_id = users.id WHERE login = $1")
	if err != nil {
		log.Printf("sql validation error: %v", err)
		return nil, err
	}
	defer stmt.Close()

	var withdrawals []models.Withdraw
	rows, err := stmt.Query(login)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("user %s withdrawals not found", login)
			return nil, ErrStorageWithdrawalsNotFound
		}
		log.Printf("sql execution error: %v", err)
		return nil, err
	}

	for rows.Next() {
		var withdraw models.Withdraw
		if err := rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			return nil, fmt.Errorf("row deserialization error %w", err)
		}
		withdraw.Login = login
		withdrawals = append(withdrawals, withdraw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row deserialization error %w", err)
	}

	return &withdrawals, nil
}

func rollbackTransaction(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		log.Printf("error rolling back transaction: %v", err)
	}
}
