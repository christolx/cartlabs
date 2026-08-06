package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/httpapi"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/christolx/cartlabs/internal/platform"
	"github.com/christolx/cartlabs/internal/purchase"
	marketstore "github.com/christolx/cartlabs/internal/store"
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
	identityService, err := identity.NewService(identity.NewPostgresRepository(dependencies.Postgres), identity.Config{
		AccessSecret: []byte(cfg.AccessTokenSecret),
		AccessTTL:    cfg.AccessTokenTTL,
		RefreshTTL:   cfg.RefreshTokenTTL,
		DemoMode:     cfg.DemoMode,
	})
	if err != nil {
		logger.Error("configure identity", "error", err)
		os.Exit(1)
	}
	storeService := marketstore.NewService(marketstore.NewPostgresRepository(dependencies.Postgres))
	catalogService := catalog.NewService(catalog.NewPostgresRepository(dependencies.Postgres))
	purchaseService, err := purchase.NewService(
		purchase.NewPostgresRepository(dependencies.Postgres),
		purchase.NewHTTPPaymentProvider(cfg.PaymentProviderURL, cfg.MockPaymentAPIKey),
		purchase.Config{WebhookSecret: []byte(cfg.PaymentWebhookSecret), WebhookURL: cfg.PaymentWebhookURL,
			ReservationTTL: 15 * time.Minute, DemoMode: cfg.DemoMode},
	)
	if err != nil {
		logger.Error("configure purchases", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr: cfg.APIAddress,
		Handler: httpapi.New(dependencies, logger,
			httpapi.WithServices(identityService, storeService, catalogService),
			httpapi.WithPurchaseService(purchaseService),
			httpapi.WithAuthRateLimiter(identity.NewRedisRateLimiter(dependencies.Redis)),
			httpapi.WithRefreshCookie(cfg.CookieSecure, cfg.RefreshTokenTTL),
		).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", cfg.APIAddress, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case <-stopCtx.Done():
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("serve API", "error", err)
		}
		return
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown API", "error", err)
	}
}
