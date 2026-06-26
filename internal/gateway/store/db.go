package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/tursodatabase/go-libsql"
)

type Store struct {
	DB *sql.DB
}

// Open opens a database connection. Supports:
//   - Local file:   Open("./data/gateway.db") or Open("file:/abs/path")
//   - Turso cloud:  Open("libsql://db-org.turso.io?authToken=...")
func Open(dsn string) (*Store, error) {
	if !strings.HasPrefix(dsn, "libsql://") && !strings.HasPrefix(dsn, "file:") {
		// Local path — create parent dir and add file: prefix
		if err := os.MkdirAll(filepath.Dir(dsn), 0755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
		dsn = "file:" + dsn
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{DB: db}, nil
}

func migrate(db *sql.DB) error {
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS workspaces (id TEXT PRIMARY KEY, name TEXT NOT NULL, slug TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS users (id TEXT PRIMARY KEY, email TEXT NOT NULL UNIQUE, password_hash TEXT NOT NULL, api_key_hash TEXT NOT NULL, created_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS workspace_members (workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, role TEXT NOT NULL DEFAULT 'admin', PRIMARY KEY (workspace_id, user_id))`,
		`CREATE TABLE IF NOT EXISTS daemons (id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, name TEXT NOT NULL, token_hash TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'offline', last_seen TEXT, created_at TEXT NOT NULL, UNIQUE(user_id, name))`,
		`CREATE TABLE IF NOT EXISTS daemon_harnesses (daemon_id TEXT NOT NULL REFERENCES daemons(id) ON DELETE CASCADE, harness TEXT NOT NULL, PRIMARY KEY (daemon_id, harness))`,
		`CREATE TABLE IF NOT EXISTS agents (id TEXT PRIMARY KEY, workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, daemon_id TEXT NOT NULL REFERENCES daemons(id), name TEXT NOT NULL, harness TEXT NOT NULL, model TEXT DEFAULT '', extra_args TEXT DEFAULT '', enabled INTEGER DEFAULT 1, personality TEXT DEFAULT '', telegram_token_hash TEXT DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS sessions (id TEXT PRIMARY KEY, agent_id TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE, chat_id TEXT NOT NULL, user_name TEXT DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w\nSQL: %s", err, stmt)
		}
	}

	// Schema evolution for existing databases — ignore errors if columns already exist
	db.Exec("ALTER TABLE daemons ADD COLUMN user_id TEXT DEFAULT ''")
	db.Exec("ALTER TABLE daemons ADD COLUMN harness_models TEXT DEFAULT ''")
	db.Exec("ALTER TABLE users ADD COLUMN api_key_raw TEXT DEFAULT ''")
	db.Exec("ALTER TABLE agents ADD COLUMN telegram_token_raw TEXT DEFAULT ''")
	db.Exec("DROP TABLE IF EXISTS sessions_old")
	db.Exec("ALTER TABLE sessions RENAME TO sessions_old")
	db.Exec("CREATE TABLE IF NOT EXISTS sessions (id TEXT PRIMARY KEY, agent_id TEXT NOT NULL REFERENCES agents(id) ON DELETE CASCADE, chat_id TEXT NOT NULL, user_name TEXT DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)")
	if err := migrateAgentsTelegramTokenConstraint(db); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_agents_telegram_token ON agents(telegram_token_hash) WHERE telegram_token_hash != ''`); err != nil {
		return fmt.Errorf("create agents telegram token index: %w", err)
	}
	return nil
}

func migrateAgentsTelegramTokenConstraint(db *sql.DB) error {
	var schema string
	if err := db.QueryRow("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'agents'").Scan(&schema); err != nil {
		return fmt.Errorf("read agents schema: %w", err)
	}
	normalized := strings.ToLower(schema)
	for _, old := range []string{" ", "\n", "\r", "\t"} {
		normalized = strings.ReplaceAll(normalized, old, "")
	}
	if !strings.Contains(normalized, "unique(telegram_token_hash)") {
		return nil
	}

	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}
	defer db.Exec("PRAGMA foreign_keys = ON")

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin agents migration: %w", err)
	}
	defer tx.Rollback()

	stmts := []string{
		`DROP TABLE IF EXISTS agents_new`,
		`CREATE TABLE agents_new (id TEXT PRIMARY KEY, workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE, daemon_id TEXT NOT NULL REFERENCES daemons(id), name TEXT NOT NULL, harness TEXT NOT NULL, model TEXT DEFAULT '', extra_args TEXT DEFAULT '', enabled INTEGER DEFAULT 1, personality TEXT DEFAULT '', telegram_token_hash TEXT DEFAULT '', telegram_token_raw TEXT DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`INSERT INTO agents_new (id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, personality, telegram_token_hash, telegram_token_raw, created_at, updated_at)
		 SELECT id, workspace_id, daemon_id, name, harness, model, extra_args, enabled, personality, telegram_token_hash, telegram_token_raw, created_at, updated_at FROM agents`,
		`DROP TABLE agents`,
		`ALTER TABLE agents_new RENAME TO agents`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("migrate agents telegram token constraint: %w\nSQL: %s", err, stmt)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit agents migration: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}
