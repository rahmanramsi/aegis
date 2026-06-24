package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	GatewayURL     string        `env:"AEGIS_GATEWAY_URL" envDefault:"ws://localhost:8080/ws/daemon"`
	DaemonID       string        `env:"AEGIS_DAEMON_ID"`
	DaemonName     string        `env:"AEGIS_DAEMON_NAME" envDefault:"aegis-agent"`
	Token          string        `env:"AEGIS_DAEMON_TOKEN"`
	WorkspacesRoot string        `env:"AEGIS_WORKSPACES_ROOT" envDefault:"./workspaces"`
	MaxConcurrent  int           `env:"AEGIS_MAX_CONCURRENT" envDefault:"5"`
	AgentTimeout   time.Duration `env:"AEGIS_AGENT_TIMEOUT" envDefault:"30m"`
	LogLevel       string        `env:"AEGIS_LOG_LEVEL" envDefault:"info"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
