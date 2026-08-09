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
	OTLPTraceEndpoint    string
	TraceSampleRatio     float64
	WorkerMetricsAddress string
	SearchGRPCAddress    string
	SearchServerAddress  string
	SearchMetricsAddress string
	SearchDatabaseURL    string
	SearchServiceToken   string
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
		OTLPTraceEndpoint:    envOr("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", ""),
		WorkerMetricsAddress: envOr("WORKER_METRICS_ADDR", ":9091"),
		SearchGRPCAddress:    envOr("SEARCH_GRPC_ADDR", ""),
		SearchServerAddress:  envOr("SEARCH_GRPC_SERVER_ADDR", ":9092"),
		SearchMetricsAddress: envOr("SEARCH_METRICS_ADDR", ":9093"),
		SearchDatabaseURL:    envOr("SEARCH_DATABASE_URL", "postgres://cartlabs:cartlabs@localhost:5432/cartlabs_search?sslmode=disable"),
		SearchServiceToken:   envOr("SEARCH_SERVICE_TOKEN", "cartlabs-local-search-token-change-me"),
	}
	var err error
	if cfg.TraceSampleRatio, err = envFloat("OTEL_TRACES_SAMPLER_ARG", 1); err != nil {
		return Config{}, err
	}
	if cfg.DemoMode, err = envBool("DEMO_MODE", false); err != nil {
		return Config{}, err
	}
	if cfg.CookieSecure, err = envBool("COOKIE_SECURE", cfg.Environment != "local"); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" || cfg.RedisAddress == "" || cfg.RabbitMQURL == "" {
		return Config{}, fmt.Errorf("database, redis, and RabbitMQ configuration must be set")
	}
	if cfg.TraceSampleRatio <= 0 || cfg.TraceSampleRatio > 1 {
		return Config{}, fmt.Errorf("OTEL_TRACES_SAMPLER_ARG must be greater than 0 and at most 1")
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
	if len(cfg.SearchServiceToken) < 32 {
		return Config{}, fmt.Errorf("SEARCH_SERVICE_TOKEN must contain at least 32 bytes")
	}
	if cfg.SearchDatabaseURL == "" || cfg.SearchServerAddress == "" || cfg.SearchMetricsAddress == "" {
		return Config{}, fmt.Errorf("search database and listener configuration must be set")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" && cfg.AccessTokenSecret == "cartlabs-local-access-token-secret-change-me" {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_SECRET must be set outside local environment")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" &&
		(cfg.PaymentWebhookSecret == "cartlabs-local-webhook-secret-change-me" || cfg.MockPaymentAPIKey == "cartlabs-local-payment-api-key") {
		return Config{}, fmt.Errorf("payment secrets must be set outside local environment")
	}
	if cfg.Environment != "local" && cfg.Environment != "test" && cfg.SearchServiceToken == "cartlabs-local-search-token-change-me" {
		return Config{}, fmt.Errorf("SEARCH_SERVICE_TOKEN must be set outside local environment")
	}

	return cfg, nil
}

func envFloat(key string, fallback float64) (float64, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
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
