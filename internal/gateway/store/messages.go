package store

import "time"

type Message struct {
	ID        int64  `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	AgentID   *string
	CreatedAt string `json:"created_at"`
}

func (s *Store) CreateMessage(sessionID, role, content, agentID string) (*Message, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.DB.Exec(
		"INSERT INTO messages (session_id, role, content, agent_id, created_at) VALUES (?, ?, ?, ?, ?)",
		sessionID, role, content, nilIfEmpty(agentID), now,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	var agentIDPtr *string
	if agentID != "" {
		agentIDPtr = &agentID
	}
	return &Message{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		AgentID:   agentIDPtr,
		CreatedAt: now,
	}, nil
}

func (s *Store) ListMessages(sessionID string, limit, offset int) ([]Message, error) {
	rows, err := s.DB.Query(
		"SELECT id, session_id, role, content, agent_id, created_at FROM messages WHERE session_id = ? ORDER BY id ASC LIMIT ? OFFSET ?",
		sessionID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.AgentID, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
