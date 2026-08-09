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
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	if cfg.SearchGRPCAddress == "" {
		logger.Error("SEARCH_GRPC_ADDR is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect source database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	client, err := searchservice.NewClient(cfg.SearchGRPCAddress, cfg.SearchServiceToken)
	if err != nil {
		logger.Error("configure search client", "error", err)
		os.Exit(1)
	}
	defer client.Close()
	count, deleted, err := searchservice.Reindex(ctx, pool, client)
	if err != nil {
		logger.Error("reindex search", "error", err)
		os.Exit(1)
	}
	logger.Info("search reindex complete", "documents", count, "pruned", deleted)
}
