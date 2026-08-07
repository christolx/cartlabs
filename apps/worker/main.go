package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/messaging"
	"github.com/christolx/cartlabs/internal/platform"
	"github.com/christolx/cartlabs/internal/purchase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()
	dependencies, err := platform.Connect(startupCtx, cfg)
	if err != nil {
		logger.Error("connect dependencies", "error", err)
		os.Exit(1)
	}
	defer dependencies.Close()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	publisher, err := messaging.NewRabbitPublisher(dependencies.RabbitMQ, messaging.DefaultExchange)
	if err != nil {
		logger.Error("configure outbox publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()
	consumer, err := messaging.NewNotificationConsumer(dependencies.RabbitMQ, messaging.DefaultExchange, messaging.DefaultQueue)
	if err != nil {
		logger.Error("configure notification consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()
	outbox := messaging.NewOutbox(dependencies.Postgres)
	notifications := messaging.NewNotificationHandler(dependencies.Postgres)
	purchases, err := purchase.NewService(purchase.NewPostgresRepository(dependencies.Postgres), nil, purchase.Config{
		WebhookSecret: []byte(cfg.PaymentWebhookSecret), WebhookURL: cfg.PaymentWebhookURL,
	})
	if err != nil {
		logger.Error("configure reservation expiry", "error", err)
		os.Exit(1)
	}

	consumerErrors := make(chan error, 1)
	go func() { consumerErrors <- consumer.Run(stopCtx, notifications, logger) }()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	logger.Info("worker ready", "environment", cfg.Environment)
	for {
		select {
		case <-stopCtx.Done():
			logger.Info("worker stopped")
			return
		case err := <-consumerErrors:
			if err != nil {
				logger.Error("notification consumer stopped", "error", err)
			}
			return
		case tick := <-ticker.C:
			if expired, err := purchases.ExpireReservations(stopCtx); err != nil {
				logger.Error("expire reservations", "error", err)
			} else if expired > 0 {
				logger.Info("expired reservations", "count", expired)
			}
			for range 25 {
				relayed, err := outbox.RelayOne(stopCtx, publisher, tick.UTC())
				if err != nil {
					logger.Error("relay outbox", "error", err)
					break
				}
				if !relayed {
					break
				}
			}
		}
	}
}
