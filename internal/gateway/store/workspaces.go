package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Slug               string `json:"slug"`
	CreatedAt          string `json:"created_at"`
	EnrollmentKeyHash  string `json:"-"`
	HasEnrollmentKey   bool   `json:"has_enrollment_key"`
}

func (s *Store) CreateWorkspace(name, slug string) (*Workspace, error) {
	w := &Workspace{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.DB.Exec(
		"INSERT INTO workspaces (id, name, slug, created_at) VALUES (?, ?, ?, ?)",
		w.ID, w.Name, w.Slug, w.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (s *Store) scanWorkspace(row interface{ Scan(...interface{}) error }) (*Workspace, error) {
	var w Workspace
	err := row.Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) ListWorkspaces() ([]Workspace, error) {
	rows, err := s.DB.Query("SELECT id, name, slug, created_at FROM workspaces ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	workspaces := make([]Workspace, 0)
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, w)
	}
	return workspaces, rows.Err()
}

func (s *Store) GetWorkspace(id string) (*Workspace, error) {
	var w Workspace
	err := s.DB.QueryRow(
		"SELECT id, name, slug, created_at FROM workspaces WHERE id = ?", id,
	).Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (s *Store) UpdateWorkspace(id, name, slug string) (*Workspace, error) {
	_, err := s.DB.Exec("UPDATE workspaces SET name = ?, slug = ? WHERE id = ?", name, slug, id)
	if err != nil {
		return nil, err
	}
	return s.GetWorkspace(id)
}

func (s *Store) DeleteWorkspace(id string) error {
	result, err := s.DB.Exec("DELETE FROM workspaces WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) GenerateEnrollmentKey(workspaceID string) (string, error) {
	key, hash := generateAPIKey()
	_, err := s.DB.Exec("UPDATE workspaces SET enrollment_key_hash = ? WHERE id = ?", hash, workspaceID)
	return key, err
}

func (s *Store) GetWorkspaceByEnrollmentKey(keyHash string) (*Workspace, error) {
	var w Workspace
	err := s.DB.QueryRow(
		"SELECT id, name, slug, created_at FROM workspaces WHERE enrollment_key_hash = ?", keyHash,
	).Scan(&w.ID, &w.Name, &w.Slug, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
