package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"learning.local/workspace/internal/blob"
	"learning.local/workspace/internal/delivery"
	"learning.local/workspace/internal/scholar"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	objects, err := blob.FromEnv()
	if err != nil {
		slog.Error("invalid object configuration")
		os.Exit(1)
	}
	if objects != nil {
		go func() {
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					task, cancel := context.WithTimeout(ctx, 10*time.Second)
					err := objects.CleanupOnce(task, pool)
					cancel()
					if err != nil && ctx.Err() == nil {
						slog.Error("profile lifecycle retry required")
					}
				}
			}
		}()
	}
	academic, err := scholar.New(pool, objects)
	if err != nil {
		slog.Error("academic worker configuration failed", "error", err)
		os.Exit(1)
	}
	queue, err := academic.WorkerClient()
	if err != nil {
		slog.Error("academic queue configuration failed", "error", err)
		os.Exit(1)
	}
	if err = queue.Start(ctx); err != nil {
		slog.Error("academic worker start failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := queue.Stop(stopCtx); err != nil {
			slog.Error("academic worker shutdown failed", "error", err)
		}
	}()
	sender := delivery.SMTP{Address: os.Getenv("SMTP_ADDR"), From: os.Getenv("SMTP_FROM"), Username: os.Getenv("SMTP_USERNAME"), Password: os.Getenv("SMTP_PASSWORD"), Local: os.Getenv("APP_ENV") == "development" || os.Getenv("APP_ENV") == "test"}
	if err = (delivery.Worker{DB: pool, Sender: sender}).Run(ctx); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}
