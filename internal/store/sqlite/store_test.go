package sqlite

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSetAndGetSetting(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()
	if err := store.SetSetting(ctx, "font_size", "16"); err != nil {
		t.Fatalf("failed to write setting: %v", err)
	}
	got, err := store.GetSetting(ctx, "font_size")
	if err != nil {
		t.Fatalf("failed to read setting: %v", err)
	}
	if got != "16" {
		t.Fatalf("expected 16, got %s", got)
	}
}
