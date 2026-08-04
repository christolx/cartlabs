package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Environment     string
	APIAddress      string
	DatabaseURL     string
	RedisAddress    string
	RabbitMQURL     string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:     envOr("APP_ENV", "local"),
		APIAddress:      envOr("API_ADDR", ":8080"),
		DatabaseURL:     envOr("DATABASE_URL", "postgres://cartlabs:cartlabs@localhost:5432/cartlabs?sslmode=disable"),
		RedisAddress:    envOr("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL:     envOr("RABBITMQ_URL", "amqp://cartlabs:cartlabs@localhost:5672/"),
		ShutdownTimeout: 10 * time.Second,
	}

	if cfg.DatabaseURL == "" || cfg.RedisAddress == "" || cfg.RabbitMQURL == "" {
		return Config{}, fmt.Errorf("database, redis, and RabbitMQ configuration must be set")
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
