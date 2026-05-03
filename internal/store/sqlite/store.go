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

type ProgramSnapshot struct {
	FileName       string
	FilePath       string
	ContentSHA1    string
	ContentBytes   int
	RenumStart     int
	RenumIncrement int
	RenumFromLine  int
	UpdatedAt      string
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
		`CREATE TABLE IF NOT EXISTS program_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			content_sha1 TEXT NOT NULL,
			content_bytes INTEGER NOT NULL DEFAULT 0,
			renum_start INTEGER NOT NULL DEFAULT 0,
			renum_increment INTEGER NOT NULL DEFAULT 0,
			renum_from_line INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			UNIQUE(file_name, file_path, content_sha1)
		);`,
		`CREATE TABLE IF NOT EXISTS keybinds (
			command_id TEXT PRIMARY KEY,
			label TEXT NOT NULL,
			shortcut TEXT NOT NULL,
			context TEXT NOT NULL,
			implemented INTEGER NOT NULL DEFAULT 1,
			configurable INTEGER NOT NULL DEFAULT 1,
			updated_at TEXT NOT NULL
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

func (s *Store) UpsertProgramSnapshot(ctx context.Context, snapshot ProgramSnapshot) error {
	if snapshot.FileName == "" || snapshot.ContentSHA1 == "" {
		return fmt.Errorf("invalid program snapshot: file_name and content_sha1 are required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO program_snapshots(
			file_name, file_path, content_sha1, content_bytes,
			renum_start, renum_increment, renum_from_line, updated_at
		)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_name, file_path, content_sha1) DO UPDATE SET
			content_bytes=excluded.content_bytes,
			renum_start=excluded.renum_start,
			renum_increment=excluded.renum_increment,
			renum_from_line=excluded.renum_from_line,
			updated_at=excluded.updated_at;
	`, snapshot.FileName, snapshot.FilePath, snapshot.ContentSHA1, snapshot.ContentBytes,
		snapshot.RenumStart, snapshot.RenumIncrement, snapshot.RenumFromLine,
		time.Now().Format(time.RFC3339))
	return err
}

func (s *Store) GetLatestProgramSnapshot(ctx context.Context, fileName, filePath string) (ProgramSnapshot, error) {
	var snap ProgramSnapshot
	err := s.db.QueryRowContext(ctx, `
		SELECT file_name, file_path, content_sha1, content_bytes,
		       renum_start, renum_increment, renum_from_line, updated_at
		FROM program_snapshots
		WHERE file_name = ? AND file_path = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`, fileName, filePath).Scan(
		&snap.FileName,
		&snap.FilePath,
		&snap.ContentSHA1,
		&snap.ContentBytes,
		&snap.RenumStart,
		&snap.RenumIncrement,
		&snap.RenumFromLine,
		&snap.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return ProgramSnapshot{}, nil
	}
	return snap, err
}
