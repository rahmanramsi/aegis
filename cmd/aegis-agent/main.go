package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rahmanramsi/aegis/internal/config"
	"github.com/rahmanramsi/aegis/internal/daemon"
	"github.com/rahmanramsi/aegis/internal/daemon/harness"
)

func main() {
	cfg, err := config.LoadDaemon()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	level := slog.LevelInfo
	if cfg.LogLevel == "debug" {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	runners := []harness.Runner{
		harness.NewEchoRunner(),
		harness.NewPiRunner(getPath("pi"), getModel("PI")),
		harness.NewClaudeRunner(getPath("claude"), getModel("CLAUDE")),
		harness.NewCodexRunner(getPath("codex"), getModel("CODEX")),
		harness.NewOpenCodeRunner(getPath("opencode"), getModel("OPENCODE")),
		harness.NewCopilotRunner(getPath("copilot")),
		harness.NewGeminiRunner(getPath("gemini"), getModel("GEMINI")),
	}

	reg := harness.Discover(runners)
	if len(reg) == 0 {
		slog.Warn("no harnesses discovered")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// Use API key as daemon token
	if cfg.DaemonToken == "" {
		cfg.DaemonToken = cfg.APIKey
	}
	if cfg.DaemonToken == "" {
		slog.Error("no API key — set AEGIS_API_KEY")
		os.Exit(1)
	}

	// Generate daemon ID if not set
	if cfg.DaemonID == "" {
		id, err := newUUID()
		if err != nil {
			slog.Error("generate uuid", "err", err)
			os.Exit(1)
		}
		cfg.DaemonID = id
		slog.Info("generated daemon id", "id", id)
	}

	client := daemon.NewClient(cfg, reg)
	if err := client.Run(ctx); err != nil {
		slog.Error("run", "err", err)
	}
	client.Close()
}

func getPath(name string) string {
	envKey := "AEGIS_" + strings.ToUpper(name) + "_PATH"
	if p := os.Getenv(envKey); p != "" {
		return p
	}
	p, _ := exec.LookPath(name)
	return p
}

func getModel(name string) string {
	return os.Getenv("AEGIS_" + name + "_MODEL")
}

func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
