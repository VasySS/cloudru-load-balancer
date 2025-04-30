// Package main is the entry point for the application.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/VasySS/cloudru-load-balancer/internal/app"
	"github.com/VasySS/cloudru-load-balancer/internal/config"
)

func main() {
	setupLogger()

	cfg := config.MustInit()

	slog.Info("starting the application")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("error running app", slog.Any("error", err))
	}
}

func setupLogger() {
	var slogLogger *slog.Logger

	slogLogger = slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	slog.SetDefault(slogLogger)
}
