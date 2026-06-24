package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rahmanramsi/aegis/internal/config"
	"github.com/rahmanramsi/aegis/internal/gateway"
	"github.com/rahmanramsi/aegis/internal/gateway/msg"
	"github.com/rahmanramsi/aegis/internal/gateway/router"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	s, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	defer s.Close()

	hub := ws.NewHub(s)
	r := router.NewRouter(s, hub)

	bm := msg.NewBotManager(func(ctx context.Context, tokenHash string, m msg.Message, adapter msg.Adapter) {
		agent, err := s.GetAgentByTelegramToken(tokenHash)
		if err != nil {
			adapter.Send(m.ChatID, "No agent configured for this bot.")
			return
		}
		r.HandleWithAgent(ctx, m, adapter, agent)
	})

	// Restore bots from env
	for _, tok := range cfg.TelegramTokens {
		bm.AddBot(context.Background(), tok)
	}
	if cfg.TelegramToken != "" {
		bm.AddBot(context.Background(), cfg.TelegramToken)
	}

	// Restore agents' Telegram bots from DB
	tokens, _ := s.GetAllTelegramTokens()
	for _, tok := range tokens {
		bm.AddBot(context.Background(), tok)
	}

	server := gateway.NewServer(s, hub, bm)

	slog.Info("gateway starting", "addr", cfg.Addr)

	go func() {
		if err := http.ListenAndServe(cfg.Addr, server); err != nil {
			slog.Error("server", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	slog.Info("gateway shutting down")
}
