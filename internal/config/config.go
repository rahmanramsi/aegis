package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Gateway struct {
	Addr        string `env:"AEGIS_ADDR" envDefault:":8080"`
	DatabaseURL string `env:"AEGIS_DATABASE_URL" envDefault:"./data/gateway.db"`
	APIKey      string `env:"AEGIS_API_KEY"`
	Env         string `env:"AEGIS_ENV" envDefault:"development"`
	BaseURL     string `env:"AEGIS_BASE_URL"`
}

type Daemon struct {
	APIKey         string        `env:"AEGIS_API_KEY"`
	LogLevel       string        `env:"AEGIS_LOG_LEVEL" envDefault:"info"`
	DaemonID       string        `env:"AEGIS_DAEMON_ID"`
	DaemonName     string        `env:"AEGIS_DAEMON_NAME" envDefault:"aegis-agent"`
	DaemonToken    string        `env:"AEGIS_DAEMON_TOKEN"`
	GatewayURL     string        `env:"AEGIS_GATEWAY_URL" envDefault:"ws://localhost:8080/ws/daemon"`
	WorkspacesRoot string        `env:"AEGIS_WORKSPACES_ROOT" envDefault:"./workspaces"`
	AgentTimeout   time.Duration `env:"AEGIS_AGENT_TIMEOUT" envDefault:"30m"`
}

func LoadGateway() (*Gateway, error) {
	cfg := &Gateway{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadDaemon() (*Daemon, error) {
	cfg := &Daemon{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
