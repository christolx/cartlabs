package platform

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/observability"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	RabbitMQ *amqp.Connection

	rabbitMQURL string
	rabbitMu    sync.Mutex
}

func Connect(ctx context.Context, cfg config.Config) (*Dependencies, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse PostgreSQL configuration: %w", err)
	}
	poolConfig.ConnConfig.Tracer = observability.NewPGXTracer()
	postgres, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL: %w", err)
	}
	if err := postgres.Ping(ctx); err != nil {
		postgres.Close()
		return nil, fmt.Errorf("connect PostgreSQL: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:            cfg.RedisAddress,
		DialTimeout:     2 * time.Second,
		ReadTimeout:     2 * time.Second,
		WriteTimeout:    2 * time.Second,
		MaxRetries:      1,
		MinRetryBackoff: 50 * time.Millisecond,
		MaxRetryBackoff: 100 * time.Millisecond,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		postgres.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("connect Redis: %w", err)
	}

	rabbit, err := dialRabbit(ctx, cfg.RabbitMQURL)
	if err != nil {
		postgres.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("connect RabbitMQ: %w", err)
	}

	return &Dependencies{Postgres: postgres, Redis: redisClient, RabbitMQ: rabbit, rabbitMQURL: cfg.RabbitMQURL}, nil
}

func dialRabbit(ctx context.Context, address string) (*amqp.Connection, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	return amqp.DialConfig(address, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial: func(network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
	})
}

func (d *Dependencies) Close() {
	d.rabbitMu.Lock()
	defer d.rabbitMu.Unlock()
	if d.RabbitMQ != nil {
		_ = d.RabbitMQ.Close()
	}
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
	if d.Postgres != nil {
		d.Postgres.Close()
	}
}

func (d *Dependencies) Check(ctx context.Context) (map[string]string, bool) {
	states := map[string]string{
		"postgres": "ok",
		"redis":    "ok",
		"rabbitmq": "ok",
	}
	ready := true

	if err := d.Postgres.Ping(ctx); err != nil {
		states["postgres"] = "unavailable"
		ready = false
	}
	if err := d.Redis.Ping(ctx).Err(); err != nil {
		states["redis"] = "unavailable"
		ready = false
	}
	if !d.rabbitReady(ctx) {
		states["rabbitmq"] = "unavailable"
		ready = false
	}

	return states, ready
}

func (d *Dependencies) rabbitReady(ctx context.Context) bool {
	d.rabbitMu.Lock()
	defer d.rabbitMu.Unlock()
	if d.RabbitMQ != nil && !d.RabbitMQ.IsClosed() {
		return true
	}
	connection, err := dialRabbit(ctx, d.rabbitMQURL)
	if err != nil {
		return false
	}
	d.RabbitMQ = connection
	return true
}
