package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Agent struct {
	ID                string `json:"id"`
	WorkspaceID       string `json:"workspace_id"`
	DaemonID          string `json:"daemon_id"`
	Name              string `json:"name"`
	Harness           string `json:"harness"`
	Model             string `json:"model"`
	ExtraArgs         string `json:"extra_args"`
	Enabled           bool   `json:"enabled"`
	Personality       string `json:"personality"`
	HasTelegramToken  bool   `json:"has_telegram_token"`
	TelegramTokenHash string `json:"-"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func (s *Store) CreateAgent(workspaceID, daemonID, name, harness, model, extraArgs, personality, telegramTokenHash, telegramTokenRaw string, enabled bool) (*Agent, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	a := &Agent{
		ID:                uuid.NewString(),
		WorkspaceID:       workspaceID,
		DaemonID:          daemonID,
		Name:              name,
		Harness:           harness,
		Model:             model,
		ExtraArgs:         extraArgs,
		Enabled:           enabled,
		Personality:       personality,
		HasTelegramToken:  telegramTokenHash != "",
		TelegramTokenHash: telegramTokenHash,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	_, err := s.DB.Exec(
		"INSERT INTO agents (id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, personality, telegram_token_hash, telegram_token_raw, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		a.ID, a.WorkspaceID, a.DaemonID, a.Name, a.Harness, a.Model, a.ExtraArgs, enabledInt, a.Personality, a.TelegramTokenHash, telegramTokenRaw, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func scanAgent(row interface{ Scan(...interface{}) error }) (*Agent, error) {
	var a Agent
	err := row.Scan(&a.ID, &a.WorkspaceID, &a.DaemonID, &a.Name, &a.Harness, &a.Model, &a.ExtraArgs, &a.Enabled, &a.Personality, &a.TelegramTokenHash, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.HasTelegramToken = a.TelegramTokenHash != ""
	return &a, nil
}

func (s *Store) ListAgents(workspaceID string) ([]Agent, error) {
	rows, err := s.DB.Query(
		`SELECT id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, 
		        COALESCE(personality, '') as personality,
		        COALESCE(telegram_token_hash, '') as telegram_token_hash, created_at, updated_at 
		 FROM agents WHERE workspace_id = ? ORDER BY created_at DESC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := make([]Agent, 0)
	for rows.Next() {
		a, err := scanAgent(rows)
		if err != nil {
			return nil, err
		}
		agents = append(agents, *a)
	}
	return agents, rows.Err()
}

func (s *Store) GetAgent(id string) (*Agent, error) {
	return scanAgent(s.DB.QueryRow(
		`SELECT id, workspace_id, daemon_id, name, harness, model, extra_args, enabled,
		        COALESCE(personality, '') as personality,
		        COALESCE(telegram_token_hash, '') as telegram_token_hash, created_at, updated_at
		 FROM agents WHERE id = ?`, id,
	))
}

func (s *Store) UpdateAgent(id, name, harness, model, extraArgs, personality string, enabled bool) (*Agent, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := s.DB.Exec(
		"UPDATE agents SET name = ?, harness = ?, model = ?, extra_args = ?, personality = ?, enabled = ?, updated_at = ? WHERE id = ?",
		name, harness, model, extraArgs, personality, enabledInt, now, id,
	)
	if err != nil {
		return nil, err
	}
	return s.GetAgent(id)
}

func (s *Store) DeleteAgent(id string) error {
	result, err := s.DB.Exec("DELETE FROM agents WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GetAgentByTelegramToken(tokenHash string) (*Agent, error) {
	return scanAgent(s.DB.QueryRow(
		`SELECT id, workspace_id, daemon_id, name, harness, model, extra_args, enabled,
		        COALESCE(personality, '') as personality,
		        COALESCE(telegram_token_hash, '') as telegram_token_hash, created_at, updated_at
		 FROM agents WHERE telegram_token_hash = ? AND enabled = 1`, tokenHash,
	))
}

func (s *Store) GetAllTelegramTokens() ([]string, error) {
	rows, err := s.DB.Query("SELECT telegram_token_raw FROM agents WHERE telegram_token_raw != '' AND enabled = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return tokens, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}
