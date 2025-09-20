package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "Accrual system address")
	flag.Parse()

	// Переопределение флагов переменными окружения
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.RunAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualSystemAddress = envAccrualAddr
	}

	// Секрет для JWT (лучше хранить в переменных окружения)
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	} else {
		cfg.JWTSecret = "default-secret-key-change-in-production"
	}

	return cfg
}
