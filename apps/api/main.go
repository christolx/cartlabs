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
	"github.com/christolx/cartlabs/internal/observability"
	"github.com/christolx/cartlabs/internal/platform"
	"github.com/christolx/cartlabs/internal/purchase"
	searchservice "github.com/christolx/cartlabs/internal/search"
	marketstore "github.com/christolx/cartlabs/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	traceShutdown, err := observability.SetupTracing(context.Background(), observability.TraceConfig{
		ServiceName: "cartlabs-api", Environment: cfg.Environment, Endpoint: cfg.OTLPTraceEndpoint, SampleRatio: cfg.TraceSampleRatio,
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
	metrics := observability.NewHTTPMetrics("api")
	catalogOptions := []catalog.Option{catalog.WithSearchObserver(func(result string) {
		metrics.CatalogSearchResults.WithLabelValues(result).Inc()
		if result == "fallback" {
			logger.Warn("catalog search compatibility fallback")
		}
	})}
	var searchClient *searchservice.Client
	if cfg.SearchGRPCAddress != "" {
		searchClient, err = searchservice.NewClient(cfg.SearchGRPCAddress, cfg.SearchServiceToken)
		if err != nil {
			logger.Error("configure search client", "error", err)
			os.Exit(1)
		}
		defer searchClient.Close()
		catalogOptions = append(catalogOptions, catalog.WithCandidateSearcher(searchClient))
	}
	catalogService := catalog.NewService(catalog.NewPostgresRepository(dependencies.Postgres), catalogOptions...)
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
			httpapi.WithMetrics(metrics),
		).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
