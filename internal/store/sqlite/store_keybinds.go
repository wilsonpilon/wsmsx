package sqlite

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"
)

// KeybindRecord stores a command shortcut binding persisted in SQLite.
type KeybindRecord struct {
	CommandID    string
	Label        string
	Shortcut     string
	Context      string
	Implemented  bool
	Configurable bool
	UpdatedAt    string
}

func (s *Store) SeedKeybinds(ctx context.Context, keybinds []KeybindRecord) error {
	now := time.Now().Format(time.RFC3339)
	for _, kb := range keybinds {
		if strings.TrimSpace(kb.CommandID) == "" {
			continue
		}
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO keybinds(command_id, label, shortcut, context, implemented, configurable, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(command_id) DO NOTHING;
		`,
			strings.TrimSpace(kb.CommandID),
			strings.TrimSpace(kb.Label),
			strings.TrimSpace(kb.Shortcut),
			strings.TrimSpace(kb.Context),
			boolToInt(kb.Implemented),
			boolToInt(kb.Configurable),
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) UpsertKeybind(ctx context.Context, kb KeybindRecord) error {
	if strings.TrimSpace(kb.CommandID) == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO keybinds(command_id, label, shortcut, context, implemented, configurable, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(command_id) DO UPDATE SET
			label=excluded.label,
			shortcut=excluded.shortcut,
			context=excluded.context,
			implemented=excluded.implemented,
			configurable=excluded.configurable,
			updated_at=excluded.updated_at;
	`,
		strings.TrimSpace(kb.CommandID),
		strings.TrimSpace(kb.Label),
		strings.TrimSpace(kb.Shortcut),
		strings.TrimSpace(kb.Context),
		boolToInt(kb.Implemented),
		boolToInt(kb.Configurable),
		time.Now().Format(time.RFC3339),
	)
	return err
}

func (s *Store) ListKeybinds(ctx context.Context) ([]KeybindRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT command_id, label, shortcut, context, implemented, configurable, updated_at
		FROM keybinds;
	`)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []KeybindRecord
	for rows.Next() {
		var (
			rec          KeybindRecord
			implemented  int
			configurable int
		)
		if scanErr := rows.Scan(&rec.CommandID, &rec.Label, &rec.Shortcut, &rec.Context, &implemented, &configurable, &rec.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		rec.Implemented = implemented != 0
		rec.Configurable = configurable != 0
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Context != out[j].Context {
			return out[i].Context < out[j].Context
		}
		if out[i].Shortcut != out[j].Shortcut {
			return out[i].Shortcut < out[j].Shortcut
		}
		return out[i].CommandID < out[j].CommandID
	})

	return out, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
