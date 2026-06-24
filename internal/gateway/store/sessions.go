package store

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           string `json:"id"`
	ConnectionID string `json:"connection_id"`
	UserName     string `json:"user_name"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (s *Store) GetOrCreateSession(connectionID, userName string) (*Session, error) {
	// Try to find the most recent session for this connection
	var sess Session
	err := s.DB.QueryRow(
		"SELECT id, connection_id, user_name, created_at, updated_at FROM sessions WHERE connection_id = ? ORDER BY updated_at DESC LIMIT 1",
		connectionID,
	).Scan(&sess.ID, &sess.ConnectionID, &sess.UserName, &sess.CreatedAt, &sess.UpdatedAt)
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
		ID:           uuid.NewString(),
		ConnectionID: connectionID,
		UserName:     userName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	_, err = s.DB.Exec(
		"INSERT INTO sessions (id, connection_id, user_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		sess.ID, sess.ConnectionID, sess.UserName, sess.CreatedAt, sess.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) ListSessions(connectionID string) ([]Session, error) {
	rows, err := s.DB.Query(
		"SELECT id, connection_id, user_name, created_at, updated_at FROM sessions WHERE connection_id = ? ORDER BY updated_at DESC",
		connectionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]Session, 0)
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.ConnectionID, &sess.UserName, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}
