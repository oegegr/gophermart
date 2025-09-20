package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/oegegr/gophermart/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(c *config.Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", c.DatabaseURI)

	if err != nil {
		log.Printf("failed to create db connection: %v", err.Error())
		return nil, err
	}

	m, err := migrate.New("file://migrations", c.DatabaseURI)
	if err != nil {
		log.Printf("failed to configure db migrations: %v", err.Error())
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("failed to apply db migrations: %v", err.Error())
		return nil, err
	}

	return db, nil
}
