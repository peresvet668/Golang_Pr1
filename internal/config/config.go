package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Address            string
	DatabaseURL        string
	MaxOpenConnections int
}

func Load() (Config, error) {
	address := envOrDefault("APP_ADDR", ":8080")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	maxOpenConnections := 50
	if raw := os.Getenv("DB_MAX_OPEN_CONNS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return Config{}, fmt.Errorf("DB_MAX_OPEN_CONNS must be a positive integer")
		}
		maxOpenConnections = parsed
	}

	return Config{
		Address:            address,
		DatabaseURL:        databaseURL,
		MaxOpenConnections: maxOpenConnections,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
