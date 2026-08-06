package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment       string
	APIAddress        string
	DatabaseURL       string
	RedisAddress      string
	RabbitMQURL       string
	AccessTokenSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	DemoMode          bool
	CookieSecure      bool
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:       envOr("APP_ENV", "local"),
		APIAddress:        envOr("API_ADDR", ":8080"),
		DatabaseURL:       envOr("DATABASE_URL", "postgres://cartlabs:cartlabs@localhost:5432/cartlabs?sslmode=disable"),
		RedisAddress:      envOr("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL:       envOr("RABBITMQ_URL", "amqp://cartlabs:cartlabs@localhost:5672/"),
		AccessTokenSecret: envOr("ACCESS_TOKEN_SECRET", "cartlabs-local-access-token-secret-change-me"),
		AccessTokenTTL:    15 * time.Minute,
		RefreshTokenTTL:   7 * 24 * time.Hour,
		ShutdownTimeout:   10 * time.Second,
	}
	var err error
	if cfg.DemoMode, err = envBool("DEMO_MODE", false); err != nil {
		return Config{}, err
	}
	if cfg.CookieSecure, err = envBool("COOKIE_SECURE", cfg.Environment != "local"); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" || cfg.RedisAddress == "" || cfg.RabbitMQURL == "" {
		return Config{}, fmt.Errorf("database, redis, and RabbitMQ configuration must be set")
	}
	if len(cfg.AccessTokenSecret) < 32 {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_SECRET must contain at least 32 bytes")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" && cfg.AccessTokenSecret == "cartlabs-local-access-token-secret-change-me" {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_SECRET must be set outside local environment")
	}

	return cfg, nil
}

func envBool(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func envOr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
