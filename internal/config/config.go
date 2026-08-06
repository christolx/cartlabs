package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment          string
	APIAddress           string
	DatabaseURL          string
	RedisAddress         string
	RabbitMQURL          string
	AccessTokenSecret    string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	DemoMode             bool
	CookieSecure         bool
	PaymentProviderURL   string
	PaymentWebhookURL    string
	PaymentWebhookSecret string
	MockPaymentAPIKey    string
	ShutdownTimeout      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:          envOr("APP_ENV", "local"),
		APIAddress:           envOr("API_ADDR", ":8080"),
		DatabaseURL:          envOr("DATABASE_URL", "postgres://cartlabs:cartlabs@localhost:5432/cartlabs?sslmode=disable"),
		RedisAddress:         envOr("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL:          envOr("RABBITMQ_URL", "amqp://cartlabs:cartlabs@localhost:5672/"),
		AccessTokenSecret:    envOr("ACCESS_TOKEN_SECRET", "cartlabs-local-access-token-secret-change-me"),
		AccessTokenTTL:       15 * time.Minute,
		RefreshTokenTTL:      7 * 24 * time.Hour,
		PaymentProviderURL:   envOr("PAYMENT_PROVIDER_URL", "http://localhost:8081"),
		PaymentWebhookURL:    envOr("PAYMENT_WEBHOOK_URL", "http://localhost:8080/api/v1/payments/webhook"),
		PaymentWebhookSecret: envOr("PAYMENT_WEBHOOK_SECRET", "cartlabs-local-webhook-secret-change-me"),
		MockPaymentAPIKey:    envOr("MOCK_PAYMENT_API_KEY", "cartlabs-local-payment-api-key"),
		ShutdownTimeout:      10 * time.Second,
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
	if len(cfg.PaymentWebhookSecret) < 32 {
		return Config{}, fmt.Errorf("PAYMENT_WEBHOOK_SECRET must contain at least 32 bytes")
	}
	if len(cfg.MockPaymentAPIKey) < 24 {
		return Config{}, fmt.Errorf("MOCK_PAYMENT_API_KEY must contain at least 24 bytes")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" && cfg.AccessTokenSecret == "cartlabs-local-access-token-secret-change-me" {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_SECRET must be set outside local environment")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" &&
		(cfg.PaymentWebhookSecret == "cartlabs-local-webhook-secret-change-me" || cfg.MockPaymentAPIKey == "cartlabs-local-payment-api-key") {
		return Config{}, fmt.Errorf("payment secrets must be set outside local environment")
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
