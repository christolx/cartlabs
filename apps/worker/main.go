package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/messaging"
	"github.com/christolx/cartlabs/internal/observability"
	"github.com/christolx/cartlabs/internal/platform"
	"github.com/christolx/cartlabs/internal/purchase"
	searchservice "github.com/christolx/cartlabs/internal/search"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	traceShutdown, err := observability.SetupTracing(context.Background(), observability.TraceConfig{
		ServiceName: "cartlabs-worker", Environment: cfg.Environment, Endpoint: cfg.OTLPTraceEndpoint, SampleRatio: cfg.TraceSampleRatio,
	})
	if err != nil {
		logger.Error("configure tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := traceShutdown(shutdownCtx); err != nil {
			logger.Error("shutdown tracing", "error", err)
		}
	}()

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
	var searchConsumer *messaging.Consumer
	var searchClient *searchservice.Client
	if cfg.SearchGRPCAddress != "" {
		searchClient, err = searchservice.NewClient(cfg.SearchGRPCAddress, cfg.SearchServiceToken)
		if err != nil {
			logger.Error("configure search client", "error", err)
			os.Exit(1)
		}
		defer searchClient.Close()
		searchConsumer, err = messaging.NewSearchConsumer(dependencies.RabbitMQ, messaging.DefaultExchange)
		if err != nil {
			logger.Error("configure search consumer", "error", err)
			os.Exit(1)
		}
		defer searchConsumer.Close()
	}
	outbox := messaging.NewOutbox(dependencies.Postgres)
	notifications := messaging.NewNotificationHandler(dependencies.Postgres)
	metrics := observability.NewWorkerMetrics()
	metrics.RegisterOutboxGauges(dependencies.Postgres)
	metricsServer := &http.Server{Addr: cfg.WorkerMetricsAddress, Handler: metrics.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	metricsErrors := make(chan error, 1)
	go func() { metricsErrors <- metricsServer.ListenAndServe() }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown worker metrics", "error", err)
		}
	}()
	purchases, err := purchase.NewService(purchase.NewPostgresRepository(dependencies.Postgres), nil, purchase.Config{
		WebhookSecret: []byte(cfg.PaymentWebhookSecret), WebhookURL: cfg.PaymentWebhookURL,
	})
	if err != nil {
		logger.Error("configure reservation expiry", "error", err)
		os.Exit(1)
	}

	consumerErrors := make(chan error, 2)
	go func() {
		consumerErrors <- consumer.Run(stopCtx, notifications, logger, func(result string) {
			metrics.NotificationResults.WithLabelValues(result).Inc()
		})
	}()
	if searchConsumer != nil {
		handler := searchservice.NewSyncHandler(searchClient)
		go func() {
			consumerErrors <- searchConsumer.Run(stopCtx, handler, logger, func(result string) {
				metrics.SearchResults.WithLabelValues(result).Inc()
			})
		}()
	}
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
		case err := <-metricsErrors:
			if err != nil && err != http.ErrServerClosed {
				logger.Error("worker metrics stopped", "error", err)
			}
			return
		case tick := <-ticker.C:
			if expired, err := purchases.ExpireReservations(stopCtx); err != nil {
				logger.Error("expire reservations", "error", err)
			} else if expired > 0 {
				metrics.ExpiredReservations.Add(float64(expired))
				logger.Info("expired reservations", "count", expired)
			}
			for range 25 {
				relayed, err := outbox.RelayOne(stopCtx, publisher, tick.UTC())
				if err != nil {
					metrics.OutboxRelays.WithLabelValues("failed").Inc()
					logger.Error("relay outbox", "error", err)
					break
				}
				if !relayed {
					break
				}
				metrics.OutboxRelays.WithLabelValues("published").Inc()
			}
		}
	}
}
