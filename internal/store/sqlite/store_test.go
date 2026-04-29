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

func TestProgramSnapshotUpsertAndGetLatest(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()
	fileName := "demo.asc"
	filePath := "c:/projects/demo.asc"

	err = store.UpsertProgramSnapshot(ctx, ProgramSnapshot{
		FileName:       fileName,
		FilePath:       filePath,
		ContentSHA1:    "sha-a",
		ContentBytes:   42,
		RenumStart:     100,
		RenumIncrement: 10,
		RenumFromLine:  0,
	})
	if err != nil {
		t.Fatalf("failed to insert snapshot A: %v", err)
	}

	err = store.UpsertProgramSnapshot(ctx, ProgramSnapshot{
		FileName:       fileName,
		FilePath:       filePath,
		ContentSHA1:    "sha-b",
		ContentBytes:   55,
		RenumStart:     200,
		RenumIncrement: 5,
		RenumFromLine:  100,
	})
	if err != nil {
		t.Fatalf("failed to insert snapshot B: %v", err)
	}

	got, err := store.GetLatestProgramSnapshot(ctx, fileName, filePath)
	if err != nil {
		t.Fatalf("failed to read latest snapshot: %v", err)
	}
	if got.ContentSHA1 != "sha-b" {
		t.Fatalf("latest snapshot SHA1 = %q, want sha-b", got.ContentSHA1)
	}
	if got.RenumStart != 200 || got.RenumIncrement != 5 || got.RenumFromLine != 100 {
		t.Fatalf("latest renum defaults mismatch: %+v", got)
	}
}
