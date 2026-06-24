package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	aegis "github.com/rahmanramsi/aegis"
	"github.com/rahmanramsi/aegis/internal/gateway"
	"github.com/rahmanramsi/aegis/internal/gateway/msg"
	"github.com/rahmanramsi/aegis/internal/gateway/router"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
	"github.com/rahmanramsi/aegis/internal/gateway/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	dbPath := os.Getenv("AEGIS_DB_PATH")
	if dbPath == "" {
		dbPath = "./data/gateway.db"
	}

	s, err := store.Open(dbPath)
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

	staticFS, _ := fs.Sub(aegis.EmbeddedStatic, "static")
	server := gateway.NewServer(s, hub, bm, staticFS)

	addr := os.Getenv("AEGIS_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	slog.Info("gateway starting", "addr", addr)

	go func() {
		if err := http.ListenAndServe(addr, server); err != nil {
			slog.Error("server", "err", err)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	slog.Info("gateway shutting down")
}
