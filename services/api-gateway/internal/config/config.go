package config

import (
	"os"
)

type Config struct {
	Port          string
	PostgresURL   string
	RedisURL      string
	NatsURL       string
	Environment   string
	LogLevel      string
	JWTSecret     string
	FinnhubAPIKey string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://portfolio_user:portfolio_pass@localhost:5432/portfolio_db?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		NatsURL:       getEnv("NATS_URL", "nats://localhost:4222"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		FinnhubAPIKey: getEnv("FINNHUB_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
