package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			last_opened_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS recent_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			last_opened_at TEXT NOT NULL
		);`,
	}

	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("erro migrando banco: %w", err)
		}
	}
	return nil
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings(key, value, updated_at)
		VALUES(?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at;
	`, key, value, time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (s *Store) TouchRecentFile(ctx context.Context, path string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO recent_files(path, last_opened_at)
		VALUES(?, ?)
		ON CONFLICT(path) DO UPDATE SET last_opened_at=excluded.last_opened_at;
	`, path, time.Now().Format(time.RFC3339))
	return err
}

