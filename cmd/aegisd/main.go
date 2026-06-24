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
	staticFS, _ := fs.Sub(aegis.EmbeddedStatic, "static")
	server := gateway.NewServer(s, hub, staticFS)

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

	// Start Telegram adapter if configured
	tg, err := msg.NewTelegramAdapter()
	if err != nil {
		slog.Info("telegram adapter not started", "reason", err)
	} else {
		r := router.NewRouter(s, hub)
		msgCh, err := tg.Start(ctx)
		if err != nil {
			slog.Error("start telegram adapter", "err", err)
		} else {
			go func() {
				for m := range msgCh {
					r.Handle(ctx, m, tg)
				}
			}()
		}
	}

	<-ctx.Done()
	slog.Info("gateway shutting down")
}
