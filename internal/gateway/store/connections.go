package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Connection struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	Platform  string `json:"platform"`
	ChatID    string `json:"chat_id"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateConnection(agentID, platform, chatID string) (*Connection, error) {
	c := &Connection{
		ID:        uuid.NewString(),
		AgentID:   agentID,
		Platform:  platform,
		ChatID:    chatID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.DB.Exec(
		"INSERT INTO connections (id, agent_id, platform, chat_id, created_at) VALUES (?, ?, ?, ?, ?)",
		c.ID, c.AgentID, c.Platform, c.ChatID, c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) ListConnections(agentID string) ([]Connection, error) {
	rows, err := s.DB.Query(
		"SELECT id, agent_id, platform, chat_id, created_at FROM connections WHERE agent_id = ? ORDER BY created_at DESC",
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	connections := make([]Connection, 0)
	for rows.Next() {
		var c Connection
		if err := rows.Scan(&c.ID, &c.AgentID, &c.Platform, &c.ChatID, &c.CreatedAt); err != nil {
			return nil, err
		}
		connections = append(connections, c)
	}
	return connections, rows.Err()
}

func (s *Store) DeleteConnection(id string) error {
	result, err := s.DB.Exec("DELETE FROM connections WHERE id = ?", id)
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

func (s *Store) FindConnection(platform, chatID string) (*Connection, error) {
	var c Connection
	err := s.DB.QueryRow(
		"SELECT id, agent_id, platform, chat_id, created_at FROM connections WHERE platform = ? AND chat_id = ?",
		platform, chatID,
	).Scan(&c.ID, &c.AgentID, &c.Platform, &c.ChatID, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}
