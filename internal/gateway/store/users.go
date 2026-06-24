package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserContextKey contextKey = "aegis_user"

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	APIKeyHash   string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

type WorkspaceMember struct {
	WorkspaceID string `json:"workspace_id"`
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
}

var (
	ErrEmailTaken    = errors.New("email already registered")
	ErrInvalidLogin  = errors.New("invalid email or password")
	ErrInvalidAPIKey = errors.New("invalid API key")
)

func UserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(UserContextKey).(*User)
	return user
}

func generateAPIKey() (plaintext, hash string) {
	b := make([]byte, 32)
	rand.Read(b)
	plaintext = "aegis-sk-" + hex.EncodeToString(b)
	hash = sha256Hex(plaintext)
	return
}

func (s *Store) CreateUser(email, password string) (*User, string, error) {
	var exists int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&exists)
	if err != nil {
		return nil, "", err
	}
	if exists > 0 {
		return nil, "", ErrEmailTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	apiKey, apiKeyHash := generateAPIKey()
	now := time.Now().UTC().Format(time.RFC3339)

	u := &User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: string(passwordHash),
		APIKeyHash:   apiKeyHash,
		CreatedAt:    now,
	}

	_, err = s.DB.Exec(
		"INSERT INTO users (id, email, password_hash, api_key_hash, created_at) VALUES (?, ?, ?, ?, ?)",
		u.ID, u.Email, u.PasswordHash, u.APIKeyHash, u.CreatedAt,
	)
	if err != nil {
		return nil, "", err
	}

	return u, apiKey, nil
}

func (s *Store) LoginUser(email, password string) (*User, string, error) {
	var u User
	err := s.DB.QueryRow(
		"SELECT id, email, password_hash, api_key_hash, created_at FROM users WHERE email = ?",
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.APIKeyHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, "", ErrInvalidLogin
	}
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidLogin
	}

	return &u, "", nil
}

func (s *Store) GetUserByAPIKey(apiKeyHash string) (*User, error) {
	var u User
	err := s.DB.QueryRow(
		"SELECT id, email, password_hash, api_key_hash, created_at FROM users WHERE api_key_hash = ?",
		apiKeyHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.APIKeyHash, &u.CreatedAt)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}
	return &u, nil
}

func (s *Store) RotateAPIKey(userID string) (string, error) {
	apiKey, apiKeyHash := generateAPIKey()
	_, err := s.DB.Exec("UPDATE users SET api_key_hash = ? WHERE id = ?", apiKeyHash, userID)
	return apiKey, err
}

func (s *Store) AddMember(workspaceID, userID, role string) error {
	_, err := s.DB.Exec(
		"INSERT OR IGNORE INTO workspace_members (workspace_id, user_id, role) VALUES (?, ?, ?)",
		workspaceID, userID, role,
	)
	return err
}

func (s *Store) IsMember(workspaceID, userID string) (bool, error) {
	var count int
	err := s.DB.QueryRow(
		"SELECT COUNT(*) FROM workspace_members WHERE workspace_id = ? AND user_id = ?",
		workspaceID, userID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) ListUserWorkspaces(userID string) ([]Workspace, error) {
	rows, err := s.DB.Query(
		`SELECT w.id, w.name, w.slug, w.created_at
		 FROM workspaces w
		 JOIN workspace_members m ON w.id = m.workspace_id
		 WHERE m.user_id = ?
		 ORDER BY w.created_at DESC`,
		userID,
	)
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

func (s *Store) VerifyAPIKey(apiKey string) (*User, error) {
	if apiKey == "" {
		return nil, ErrInvalidAPIKey
	}
	hash := sha256Hex(apiKey)
	return s.GetUserByAPIKey(hash)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
