package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/christolx/cartlabs/internal/config"
	"github.com/christolx/cartlabs/internal/messaging"
	"github.com/christolx/cartlabs/internal/platform"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	limit := 100
	if raw := os.Getenv("REPLAY_LIMIT"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 1000 {
			logger.Error("REPLAY_LIMIT must be between 1 and 1000")
			os.Exit(1)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dependencies, err := platform.Connect(ctx, cfg)
	if err != nil {
		logger.Error("connect dependencies", "error", err)
		os.Exit(1)
	}
	defer dependencies.Close()
	outboxCount, err := messaging.NewOutbox(dependencies.Postgres).ReplayDeadLetters(ctx, limit, time.Now().UTC())
	if err != nil {
		logger.Error("replay outbox dead letters", "error", err)
		os.Exit(1)
	}
	replayer, err := messaging.NewDeadLetterReplayer(dependencies.RabbitMQ, messaging.DefaultExchange)
	if err != nil {
		logger.Error("configure notification replay", "error", err)
		os.Exit(1)
	}
	defer replayer.Close()
	notificationCount, err := replayer.Replay(ctx, limit)
	if err != nil {
		logger.Error("replay notification dead letters", "error", err)
		os.Exit(1)
	}
	logger.Info("dead letters replayed", "outbox", outboxCount, "notifications", notificationCount)
}
