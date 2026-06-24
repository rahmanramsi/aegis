package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GatewayURL     string
	DaemonID       string
	DaemonName     string
	Token          string
	WorkspacesRoot string
	MaxConcurrent  int
	AgentTimeout   time.Duration
	LogLevel       string
}

func Load() *Config {
	cfg := &Config{
		GatewayURL:     getEnv("AEGIS_GATEWAY_URL", "ws://localhost:8080/ws/daemon"),
		DaemonID:       os.Getenv("AEGIS_DAEMON_ID"),
		DaemonName:     getEnv("AEGIS_DAEMON_NAME", "aegis-agent"),
		Token:          os.Getenv("AEGIS_DAEMON_TOKEN"),
		WorkspacesRoot: getEnv("AEGIS_WORKSPACES_ROOT", "./workspaces"),
		MaxConcurrent:  getEnvInt("AEGIS_MAX_CONCURRENT", 5),
		AgentTimeout:   getEnvDuration("AEGIS_AGENT_TIMEOUT", 30*time.Minute),
		LogLevel:       getEnv("AEGIS_LOG_LEVEL", "info"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
