package platform

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Dependencies struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
	RabbitMQ *amqp.Connection
}

func Connect(ctx context.Context, cfg config.Config) (*Dependencies, error) {
	postgres, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("configure PostgreSQL: %w", err)
	}
	if err := postgres.Ping(ctx); err != nil {
		postgres.Close()
		return nil, fmt.Errorf("connect PostgreSQL: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddress})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		postgres.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("connect Redis: %w", err)
	}

	dialer := net.Dialer{Timeout: 5 * time.Second}
	rabbit, err := amqp.DialConfig(cfg.RabbitMQURL, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
		Dial: func(network, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
	})
	if err != nil {
		postgres.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("connect RabbitMQ: %w", err)
	}

	return &Dependencies{Postgres: postgres, Redis: redisClient, RabbitMQ: rabbit}, nil
}

func (d *Dependencies) Close() {
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
	if d.RabbitMQ.IsClosed() {
		states["rabbitmq"] = "unavailable"
		ready = false
	}

	return states, ready
}
