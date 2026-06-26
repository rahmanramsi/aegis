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

func (s *Store) Close() error {
	return s.DB.Close()
}
