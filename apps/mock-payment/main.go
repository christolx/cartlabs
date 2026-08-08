package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/mockpayment"
	"github.com/christolx/cartlabs/internal/observability"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	traceShutdown, err := observability.SetupTracing(context.Background(), observability.TraceConfig{
		ServiceName: "cartlabs-mock-payment", Environment: envOr("APP_ENV", "local"), Endpoint: os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"), SampleRatio: 1,
	})
	if err != nil {
		logger.Error("configure tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := traceShutdown(shutdownCtx); err != nil {
			logger.Error("shutdown tracing", "error", err)
		}
	}()
	address := os.Getenv("MOCK_PAYMENT_ADDR")
	if address == "" {
		address = ":8081"
	}

	apiKey := os.Getenv("MOCK_PAYMENT_API_KEY")
	if apiKey == "" {
		apiKey = "cartlabs-local-payment-api-key"
	}
	webhookSecret := os.Getenv("PAYMENT_WEBHOOK_SECRET")
	if webhookSecret == "" {
		webhookSecret = "cartlabs-local-webhook-secret-change-me"
	}
	application, err := mockpayment.New(mockpayment.Config{APIKey: apiKey, WebhookSecret: []byte(webhookSecret)})
	if err != nil {
		logger.Error("configure mock payment", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              address,
		Handler:           otelhttp.NewHandler(application.Handler(), "mock-payment.http"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info("mock payment listening", "address", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("serve mock payment", "error", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown mock payment", "error", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
