package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := database.EnsureDatabase(ctx, cfg.SearchDatabaseURL); err != nil {
		logger.Error("ensure search database", "error", err)
		os.Exit(1)
	}
	pool, err := database.Connect(ctx, cfg.SearchDatabaseURL)
	if err != nil {
		logger.Error("connect search database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool, "migrations/search"); err != nil {
		logger.Error("migrate search database", "error", err)
		os.Exit(1)
	}
	logger.Info("search migrations complete")
}
