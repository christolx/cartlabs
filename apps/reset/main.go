package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/christolx/cartlabs/internal/cloudinary"
	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/database"
	searchservice "github.com/christolx/cartlabs/internal/search"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	cleanup, cleanupErr := cloudinary.DeleteDemoUploads(ctx, http.DefaultClient, cloudinary.CleanupConfig{
		CloudName: cfg.CloudinaryCloudName, APIKey: cfg.CloudinaryAPIKey, APISecret: cfg.CloudinaryAPISecret,
	})
	if cleanupErr != nil {
		logger.Error("Cloudinary cleanup failed; database reset continuing", "error", cleanupErr)
	} else {
		logger.Info("Cloudinary cleanup complete", "deleted", cleanup.Deleted, "pages", cleanup.Pages)
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
