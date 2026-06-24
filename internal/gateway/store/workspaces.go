package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	CreatedAt string `json:"created_at"`
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
	_, err := s.DB.Exec(
		"UPDATE workspaces SET name = ?, slug = ? WHERE id = ?",
		name, slug, id,
	)
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
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
