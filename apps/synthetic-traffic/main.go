package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/observability"
	"github.com/christolx/cartlabs/internal/synthetic"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	config, err := loadConfig()
	if err != nil {
		logger.Error("load synthetic traffic configuration", "error", err)
		os.Exit(1)
	}
	shutdown, err := observability.SetupTracing(context.Background(), observability.TraceConfig{
		ServiceName: "cartlabs-synthetic-traffic", Environment: envOr("APP_ENV", "local"),
		Endpoint: os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"), SampleRatio: 1,
	})
	if err != nil {
		logger.Error("configure synthetic traffic tracing", "error", err)
		os.Exit(1)
	}
	defer func() { _ = shutdown(context.Background()) }()

	metrics := synthetic.NewMetrics()
	runner, err := synthetic.New(config, &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}, metrics)
	if err != nil {
		logger.Error("configure synthetic traffic", "error", err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	server := &http.Server{Addr: envOr("SYNTHETIC_METRICS_ADDR", ":9094"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go runner.Run(ctx, func(name string, err error) {
		if err != nil {
			logger.Error("synthetic traffic cycle failed", "cycle", name, "error", err)
			return
		}
		logger.Info("synthetic traffic cycle complete", "cycle", name)
	})
	select {
	case <-ctx.Done():
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("synthetic metrics server stopped", "error", err)
		}
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func loadConfig() (synthetic.Config, error) {
	readInterval, err := durationEnv("SYNTHETIC_READ_INTERVAL", 10*time.Second)
	if err != nil {
		return synthetic.Config{}, err
	}
	journeyInterval, err := durationEnv("SYNTHETIC_JOURNEY_INTERVAL", 15*time.Minute)
	if err != nil {
		return synthetic.Config{}, err
	}
	requestTimeout, err := durationEnv("SYNTHETIC_REQUEST_TIMEOUT", 10*time.Second)
	if err != nil {
		return synthetic.Config{}, err
	}
	notificationWait, err := durationEnv("SYNTHETIC_NOTIFICATION_WAIT", 30*time.Second)
	if err != nil {
		return synthetic.Config{}, err
	}
	return synthetic.Config{
		BaseURL: envOr("SYNTHETIC_BASE_URL", "http://localhost:8080/api/v1"),
		Email:   envOr("SYNTHETIC_EMAIL", "buyer3@demo.cartlabs.local"), Password: envOr("SYNTHETIC_PASSWORD", "demo-pass-123"),
		VariantID:    envOr("SYNTHETIC_VARIANT_ID", "01989f00-0000-7000-8000-000000000401"),
		ReadInterval: readInterval, JourneyInterval: journeyInterval, RequestTimeout: requestTimeout, NotificationWait: notificationWait,
	}, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	return time.ParseDuration(raw)
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
