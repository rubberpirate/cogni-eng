package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/delivery"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("worker configuration failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	sender := delivery.SMTP{Address: os.Getenv("SMTP_ADDR"), From: os.Getenv("SMTP_FROM"), Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"), Local: os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "test"}
	if err = (delivery.Worker{DB: pool, Sender: sender}).Run(ctx); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
