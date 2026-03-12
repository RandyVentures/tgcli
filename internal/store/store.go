package store

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store represents the local message/chat database.
type Store struct {
	db *sql.DB
}

// Open opens or creates the store database.
func Open(storeDir string) (*Store, error) {
	dbPath := filepath.Join(storeDir, "tgcli.db")
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable FTS5
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	s := &Store{db: db}

	// TODO: Run migrations, create tables
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	return s, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate creates/updates database schema.
func (s *Store) migrate() error {
	// TODO: Create tables (chats, users, messages, messages_fts)
	// For now, just a placeholder
	schema := `
		CREATE TABLE IF NOT EXISTS meta (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`
	
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}

	return nil
}
