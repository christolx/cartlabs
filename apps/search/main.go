package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	searchv1 "github.com/christolx/cartlabs/internal/contract/searchv1"
	"github.com/christolx/cartlabs/internal/observability"
	searchservice "github.com/christolx/cartlabs/internal/search"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	traceShutdown, err := observability.SetupTracing(context.Background(), observability.TraceConfig{
		ServiceName: "cartlabs-search", Environment: cfg.Environment, Endpoint: cfg.OTLPTraceEndpoint, SampleRatio: cfg.TraceSampleRatio,
	})
	if err != nil {
		logger.Error("configure tracing", "error", err)
		os.Exit(1)
	}
	defer func() { _ = traceShutdown(context.Background()) }()

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()
	poolConfig, err := pgxpool.ParseConfig(cfg.SearchDatabaseURL)
	if err != nil {
		logger.Error("parse search database", "error", err)
		os.Exit(1)
	}
	poolConfig.ConnConfig.Tracer = observability.NewPGXTracer()
	pool, err := pgxpool.NewWithConfig(startupCtx, poolConfig)
	if err == nil {
		err = pool.Ping(startupCtx)
	}
	if err != nil {
		logger.Error("connect search database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	metrics := searchservice.NewMetrics(registry)
	repository := searchservice.NewPostgresRepository(pool)
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "cartlabs_search_documents", Help: "Documents owned by search service."}, func() float64 {
		count, err := repository.Count(context.Background())
		if err != nil {
			return -1
		}
		return float64(count)
	}))

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpcServer := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.UnaryInterceptor(searchservice.UnaryAuthInterceptor(cfg.SearchServiceToken)))
	searchv1.RegisterSearchServiceServer(grpcServer, searchservice.NewServer(repository, metrics))
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	listener, err := net.Listen("tcp", cfg.SearchServerAddress)
	if err != nil {
		logger.Error("listen for search RPC", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{EnableOpenMetrics: true}))
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "search database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	metricsServer := &http.Server{Addr: cfg.SearchMetricsAddress, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second}
	errors := make(chan error, 2)
	go func() { errors <- grpcServer.Serve(listener) }()
	go func() { errors <- metricsServer.ListenAndServe() }()
	logger.Info("search service ready", "grpc", cfg.SearchServerAddress, "metrics", cfg.SearchMetricsAddress)

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case <-stopCtx.Done():
	case err := <-errors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("search server stopped", "error", err)
		}
	}
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpcServer.GracefulStop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown search metrics", "error", err)
	}
}
