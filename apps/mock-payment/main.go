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
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
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
