package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/rahmanramsi/aegis/internal/config"
	"github.com/rahmanramsi/aegis/internal/gateway"
	"github.com/rahmanramsi/aegis/internal/gateway/daemonws"
	"github.com/rahmanramsi/aegis/internal/gateway/messaging"
	"github.com/rahmanramsi/aegis/internal/gateway/routing"
	"github.com/rahmanramsi/aegis/internal/gateway/store"
)

func Run(ctx context.Context, cfg *config.Gateway) error {
	s, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer s.Close()

	hub := daemonws.NewHub(s)
	r := routing.NewRouter(s, hub)
	bm := messaging.NewBotManager(func(ctx context.Context, tokenHash string, m messaging.Message, adapter messaging.Adapter) {
		agent, err := s.GetAgentByTelegramToken(tokenHash)
		if err != nil {
			adapter.Send(m.ChatID, "No agent configured for this bot.")
			return
		}
		r.HandleWithAgent(ctx, m, adapter, agent)
	})

	tokens, err := s.GetAllTelegramTokens()
	if err != nil {
		return err
	}
	for _, token := range tokens {
		if err := bm.AddBot(ctx, token); err != nil {
			slog.Warn("restore telegram bot", "err", err)
		}
	}

	handler := gateway.NewServer(s, hub, bm, gateway.Options{
		APIKey:  cfg.APIKey,
		Env:     cfg.Env,
		BaseURL: cfg.BaseURL,
	})
	httpServer := &http.Server{Addr: cfg.Addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	slog.Info("gateway starting", "addr", cfg.Addr)

	select {
	case <-ctx.Done():
		slog.Info("gateway shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
