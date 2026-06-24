package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Daemon struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Name             string  `json:"name"`
	TokenHash        string  `json:"-"`
	Status           string  `json:"status"`
	LastSeen         *string `json:"last_seen"`
	HarnessModelsJSON string  `json:"-"`
	CreatedAt        string  `json:"created_at"`
}

func (s *Store) CreateDaemon(userID, name, tokenHash string) (*Daemon, error) {
	d := &Daemon{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		TokenHash:   tokenHash,
		Status:      "offline",
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.DB.Exec(
		"INSERT INTO daemons (id, user_id, name, token_hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		d.ID, d.UserID, d.Name, d.TokenHash, d.Status, d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}
func (s *Store) AuthenticateDaemon(daemonID, tokenHash string, harnesses []string, modelsJSON string) error {
	var storedHash string
	err := s.DB.QueryRow(
		"SELECT token_hash FROM daemons WHERE id = ?", daemonID,
	).Scan(&storedHash)
	if err != nil {
		return err
	}
	if storedHash != tokenHash {
		return ErrInvalidToken
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.DB.Exec(
		"UPDATE daemons SET status = 'online', last_seen = ?, harness_models = ? WHERE id = ?",
		now, modelsJSON, daemonID,
	)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec("DELETE FROM daemon_harnesses WHERE daemon_id = ?", daemonID)
	if err != nil {
		return err
	}
	for _, h := range harnesses {
		_, err = s.DB.Exec(
			"INSERT INTO daemon_harnesses (daemon_id, harness) VALUES (?, ?)",
			daemonID, h,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListDaemonsByUser(userID string) ([]Daemon, error) {
	rows, err := s.DB.Query(
		"SELECT id, user_id, name, token_hash, status, last_seen, harness_models, created_at FROM daemons WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	daemons := make([]Daemon, 0)
	for rows.Next() {
		var d Daemon
		if err := rows.Scan(&d.ID, &d.UserID, &d.Name, &d.TokenHash, &d.Status, &d.LastSeen, &d.HarnessModelsJSON, &d.CreatedAt); err != nil {
			return nil, err
		}
		daemons = append(daemons, d)
	}
	return daemons, rows.Err()
}

func (s *Store) GetDaemon(id string) (*Daemon, error) {
	var d Daemon
	err := s.DB.QueryRow(
		"SELECT id, user_id, name, token_hash, status, last_seen, harness_models, created_at FROM daemons WHERE id = ?", id,
	).Scan(&d.ID, &d.UserID, &d.Name, &d.TokenHash, &d.Status, &d.LastSeen, &d.HarnessModelsJSON, &d.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) DeleteDaemon(id string) error {
	result, err := s.DB.Exec("DELETE FROM daemons WHERE id = ?", id)
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

func (s *Store) SetDaemonOffline(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec(
		"UPDATE daemons SET status = 'offline', last_seen = ? WHERE id = ?",
		now, id,
	)
	return err
}

func (s *Store) SetAllDaemonsOffline() error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.Exec("UPDATE daemons SET status = 'offline', last_seen = ?", now)
	return err
}

func (s *Store) GetDaemonHarnesses(daemonID string) ([]string, error) {
	rows, err := s.DB.Query("SELECT harness FROM daemon_harnesses WHERE daemon_id = ?", daemonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
