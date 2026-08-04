package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/platform"
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

	logger.Info("worker ready", "environment", cfg.Environment)
	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()
	logger.Info("worker stopped")
}
