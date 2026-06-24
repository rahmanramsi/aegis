package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type DaemonState struct {
	DaemonID string `json:"daemon_id"`
	Token    string `json:"token"`
}

func StatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aegis", "daemon.json")
}

func LoadState() (*DaemonState, error) {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		return nil, err
	}
	var s DaemonState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func SaveState(s *DaemonState) error {
	dir := filepath.Dir(StatePath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(), data, 0600)
}
