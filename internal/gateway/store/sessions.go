package store

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	ChatID    string `json:"chat_id"`
	UserName  string `json:"user_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *Store) GetOrCreateSessionByAgent(agentID, chatID, userName string) (*Session, error) {
	// Try to find the most recent session for this agent+chat
	var sess Session
	err := s.DB.QueryRow(
		"SELECT id, agent_id, chat_id, user_name, created_at, updated_at FROM sessions WHERE agent_id = ? AND chat_id = ? ORDER BY updated_at DESC LIMIT 1",
		agentID, chatID,
	).Scan(&sess.ID, &sess.AgentID, &sess.ChatID, &sess.UserName, &sess.CreatedAt, &sess.UpdatedAt)
	if err == nil {
		// Update username if changed
		if userName != "" && sess.UserName != userName {
			sess.UserName = userName
			now := time.Now().UTC().Format(time.RFC3339)
			sess.UpdatedAt = now
			_, err = s.DB.Exec(
				"UPDATE sessions SET user_name = ?, updated_at = ? WHERE id = ?",
				userName, now, sess.ID,
			)
			if err != nil {
				return nil, err
			}
		}
		return &sess, nil
	}

	// Create new session
	now := time.Now().UTC().Format(time.RFC3339)
	sess = Session{
		ID:        uuid.NewString(),
		AgentID:   agentID,
		ChatID:    chatID,
		UserName:  userName,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.DB.Exec(
		"INSERT INTO sessions (id, agent_id, chat_id, user_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		sess.ID, sess.AgentID, sess.ChatID, sess.UserName, sess.CreatedAt, sess.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) ListSessions(agentID, chatID string) ([]Session, error) {
	rows, err := s.DB.Query(
		"SELECT id, agent_id, chat_id, user_name, created_at, updated_at FROM sessions WHERE agent_id = ? AND chat_id = ? ORDER BY updated_at DESC",
		agentID, chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.AgentID, &sess.ChatID, &sess.UserName, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}
