package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Constants for database configuration
const (
	// DefaultLimit is the default number of results returned
	DefaultLimit = 50
	// MaxLimit is the maximum number of results that can be requested
	MaxLimit = 1000
	// MaxMessageLength is the maximum allowed message text length (Telegram limit)
	MaxMessageLength = 4096
	// DBFilePermissions are the permissions for the database file (owner read/write only)
	DBFilePermissions = 0600
	// StoreDirPermissions are the permissions for the store directory (owner only)
	StoreDirPermissions = 0700
)

// Store represents the local message/chat database.
type Store struct {
	db         *sql.DB
	ftsEnabled bool // whether FTS5 is available and enabled
}

// Open opens or creates the store database.
func Open(storeDir string) (*Store, error) {
	// Create store directory with restricted permissions (owner only)
	if err := os.MkdirAll(storeDir, StoreDirPermissions); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}

	// Ensure directory has correct permissions even if it existed
	if err := os.Chmod(storeDir, StoreDirPermissions); err != nil {
		return nil, fmt.Errorf("set store dir permissions: %w", err)
	}

	dbPath := filepath.Join(storeDir, "tgcli.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool for better performance
	db.SetMaxOpenConns(1) // SQLite is single-writer, so limit connections
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Connections never expire

	// Enable WAL mode and optimizations
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000", // 5 second busy timeout
		"PRAGMA cache_size=-64000", // 64MB cache
		"PRAGMA temp_store=MEMORY", // Store temp tables in memory
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("execute %s: %w", pragma, err)
		}
	}

	s := &Store{db: db, ftsEnabled: false}

	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}

	// FTS5 is not reliably available in pure Go SQLite (modernc.org/sqlite)
	// Always use LIKE search for compatibility
	s.ftsEnabled = false

	// Set database file permissions (owner read/write only)
	if err := os.Chmod(dbPath, DBFilePermissions); err != nil {
		// Non-fatal, but verify it anyway
		if info, statErr := os.Stat(dbPath); statErr == nil {
			if info.Mode().Perm() != DBFilePermissions {
				return nil, fmt.Errorf("database file has incorrect permissions: %v (expected %v)", info.Mode().Perm(), DBFilePermissions)
			}
		}
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

type migration struct {
	version int
	name    string
	up      func(*Store) error
}

var migrations = []migration{
	{version: 1, name: "core schema", up: migrateCoreSchema},
	{version: 2, name: "fts and indices", up: migrateFTSAndIndices},
}

// migrate creates/updates database schema.
func (s *Store) migrate() error {
	// Create migrations tracking table
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Check which migrations are applied
	applied := map[int]bool{}
	rows, err := s.db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate applied migrations: %w", err)
	}

	// Apply pending migrations
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := m.up(s); err != nil {
			return fmt.Errorf("apply migration %03d %s: %w", m.version, m.name, err)
		}
		if _, err := s.db.Exec(
			`INSERT INTO schema_migrations(version, name, applied_at) VALUES(?, ?, ?)`,
			m.version,
			m.name,
			time.Now().UTC().Unix(),
		); err != nil {
			return fmt.Errorf("record migration %03d: %w", m.version, err)
		}
	}

	return nil
}

func migrateCoreSchema(s *Store) error {
	// Create tables one at a time for better error handling
	tables := []string{
		`CREATE TABLE chats (
			id INTEGER PRIMARY KEY,
			type TEXT NOT NULL,
			title TEXT,
			username TEXT,
			last_message_id INTEGER,
			last_message_ts INTEGER,
			unread_count INTEGER DEFAULT 0,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			first_name TEXT,
			last_name TEXT,
			username TEXT,
			phone TEXT,
			is_bot INTEGER DEFAULT 0,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE messages (
			id INTEGER PRIMARY KEY,
			chat_id INTEGER NOT NULL,
			from_user_id INTEGER,
			date INTEGER NOT NULL,
			text TEXT,
			reply_to_message_id INTEGER,
			media_type TEXT,
			media_path TEXT,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE INDEX idx_messages_chat_date ON messages(chat_id, date DESC)`,
		`CREATE INDEX idx_messages_from_user ON messages(from_user_id, date DESC)`,
		`CREATE INDEX idx_messages_text ON messages(text)`,
		`CREATE INDEX idx_chats_last_message ON chats(last_message_ts DESC)`,
	}

	for _, ddl := range tables {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("execute DDL: %w", err)
		}
	}

	return nil
}

// migrateFTSAndIndices adds FTS5 full-text search and additional indices.
func migrateFTSAndIndices(s *Store) error {
	// Check if FTS5 is available
	var ftsAvailable int
	err := s.db.QueryRow("SELECT COUNT(*) FROM pragma_compile_options WHERE compile_options LIKE '%FTS5%'").Scan(&ftsAvailable)
	if err != nil || ftsAvailable == 0 {
		// FTS5 not available, skip FTS table creation
		// Just add additional indices
		additionalIndices := []string{
			`CREATE INDEX IF NOT EXISTS idx_messages_media_type ON messages(media_type) WHERE media_type IS NOT NULL AND media_type != ''`,
			`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL AND username != ''`,
		}

		for _, ddl := range additionalIndices {
			if _, err := s.db.Exec(ddl); err != nil {
				return fmt.Errorf("create index: %w", err)
			}
		}
		return nil
	}

	// Create FTS5 virtual table for full-text search
	ftsStatements := []string{
		`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
			text,
			content=messages,
			content_rowid=rowid,
			tokenize='porter unicode61'
		)`,
		// Triggers to keep FTS in sync
		`CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON messages BEGIN
			INSERT INTO messages_fts(rowid, text) VALUES (new.rowid, new.text);
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON messages BEGIN
			DELETE FROM messages_fts WHERE rowid = old.rowid;
		END`,
		`CREATE TRIGGER IF NOT EXISTS messages_fts_update AFTER UPDATE ON messages BEGIN
			DELETE FROM messages_fts WHERE rowid = old.rowid;
			INSERT INTO messages_fts(rowid, text) VALUES (new.rowid, new.text);
		END`,
		// Additional useful indices
		`CREATE INDEX IF NOT EXISTS idx_messages_media_type ON messages(media_type) WHERE media_type IS NOT NULL AND media_type != ''`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL AND username != ''`,
	}

	for _, ddl := range ftsStatements {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("FTS setup: %w", err)
		}
	}

	// Build initial FTS index from existing messages
	if _, err := s.db.Exec(`INSERT INTO messages_fts(rowid, text) SELECT rowid, text FROM messages`); err != nil {
		// May fail if already populated, that's ok
	}

	return nil
}
