package services

import (
	"database/sql"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/portfolio-management/api-gateway/internal/config"
)

type Services struct {
	DB      *sql.DB
	Redis   *redis.Client
	NATS    *nats.Conn
	Finnhub *FinnhubClient
	Logger  *zap.Logger
}

func NewServices(cfg *config.Config, logger *zap.Logger) (*Services, error) {
	services := &Services{
		Logger: logger,
	}

	// Initialize PostgreSQL
	db, err := sql.Open("postgres", cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	services.DB = db

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
		DB:   0,
	})
	services.Redis = rdb

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	services.NATS = nc

	// Initialize Finnhub client
	if cfg.FinnhubAPIKey != "" {
		services.Finnhub = NewFinnhubClient(cfg.FinnhubAPIKey)
		logger.Info("Finnhub client initialized")
	} else {
		logger.Warn("Finnhub API key not provided, market data features will be limited")
	}

	logger.Info("All services initialized successfully")
	return services, nil
}

func (s *Services) Close() error {
	var errs []error

	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.NATS != nil {
		s.NATS.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing services: %v", errs)
	}

	return nil
}
