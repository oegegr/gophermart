package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	JWTSecret            string
	AccrualSystemAddress string
	AccrualInterval      time.Duration
}

func LoadConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "", "Address and port to run server")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Database connection string")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://127.0.0.1:8081", "Accrual system address")
	flag.StringVar(&cfg.JWTSecret, "s", "jwt-secret-key", "jwt secret key")
	flag.DurationVar(&cfg.AccrualInterval, "i", 1*time.Second, "interval accrual processing time")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.RunAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		cfg.AccrualSystemAddress = envAccrualAddr
	}

	if envAccrualInterval := os.Getenv("ACCRUAL_INTERVAL"); envAccrualInterval != "" {
		cfg.AccrualInterval, _ = time.ParseDuration(envAccrualInterval)
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	}

	return cfg
}
