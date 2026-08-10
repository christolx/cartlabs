package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/database"
	searchservice "github.com/christolx/cartlabs/internal/search"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.Reset(ctx, pool, cfg.Environment, "migrations", "seeds/demo.sql"); err != nil {
		logger.Error("reset database", "error", err)
		os.Exit(1)
	}
	if cfg.SearchGRPCAddress != "" {
		client, err := searchservice.NewClient(cfg.SearchGRPCAddress, cfg.SearchServiceToken)
		if err != nil {
			logger.Error("configure reset search client", "error", err)
			os.Exit(1)
		}
		defer client.Close()
		count, deleted, err := searchservice.Reindex(ctx, pool, client)
		if err != nil {
			logger.Error("reindex reset search", "error", err)
			os.Exit(1)
		}
		logger.Info("search reindex complete", "documents", count, "pruned", deleted)
	}
	logger.Info("database reset complete", "environment", cfg.Environment)
}
