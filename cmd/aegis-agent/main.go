package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rahmanramsi/aegis/internal/daemon"
	"github.com/rahmanramsi/aegis/internal/daemon/config"
	"github.com/rahmanramsi/aegis/internal/daemon/harness"
)

func main() {
	cfg, err := config.Load()
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

	client := daemon.NewClient(cfg, reg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
