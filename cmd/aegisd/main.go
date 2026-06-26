package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rahmanramsi/aegis/internal/config"
	"github.com/rahmanramsi/aegis/internal/gateway/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.LoadGateway()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("run gateway", "err", err)
		os.Exit(1)
	}
}
