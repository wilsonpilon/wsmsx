package ui

import (
	"testing"

	"ws7/internal/store/sqlite"
)

func TestApplyKeybindFilters(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "a", Context: "editor", Implemented: true, Configurable: true},
		{CommandID: "b", Context: "global", Implemented: false, Configurable: true},
		{CommandID: "c", Context: "editor", Implemented: false, Configurable: false},
	}
	got := applyKeybindFilters(records, "editor", "Not Implemented", "Locked")
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("filtered indexes = %#v, want [2]", got)
	}
}

func TestFindShortcutConflict(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "save", Label: "Save", Shortcut: "Ctrl+K,S"},
		{CommandID: "open", Label: "Open", Shortcut: "Ctrl+O,K"},
	}
	conflict, ok := findShortcutConflict(records, "Ctrl+K,S", "open")
	if !ok {
		t.Fatal("expected conflict")
	}
	if conflict.CommandID != "save" {
		t.Fatalf("conflict command = %q, want save", conflict.CommandID)
	}
}

func TestBuildKeybindImportPreviewSuccess(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "save", Label: "Save", Shortcut: "Ctrl+K,S", Configurable: true},
		{CommandID: "open_switch", Label: "Open", Shortcut: "Ctrl+O,K", Configurable: true},
	}
	jsonData := []byte(`[
		{"command_id":"save","shortcut":"Ctrl+K,Z"}
	]`)
	preview, err := buildKeybindImportPreview(records, jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preview.Errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", preview.Errors)
	}
	if len(preview.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(preview.Changes))
	}
	if preview.Result[0].Shortcut != "Ctrl+K,Z" {
		t.Fatalf("updated shortcut = %q, want Ctrl+K,Z", preview.Result[0].Shortcut)
	}
}

func TestBuildKeybindImportPreviewDetectsConflict(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "save", Label: "Save", Shortcut: "Ctrl+K,S", Configurable: true},
		{CommandID: "open_switch", Label: "Open", Shortcut: "Ctrl+O,K", Configurable: true},
	}
	jsonData := []byte(`[
		{"command_id":"save","shortcut":"Ctrl+K,S"},
		{"command_id":"open_switch","shortcut":"Ctrl+K,S"}
	]`)
	preview, err := buildKeybindImportPreview(records, jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preview.Errors) == 0 {
		t.Fatal("expected conflict validation error")
	}
}

func TestBuildKeybindImportPreviewSkipsLockedCommands(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "word_count", Label: "Word Count", Shortcut: "", Configurable: false},
	}
	jsonData := []byte(`[
		{"command_id":"word_count","shortcut":"Ctrl+W"}
	]`)
	preview, err := buildKeybindImportPreview(records, jsonData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(preview.Changes) != 0 {
		t.Fatalf("expected no changes for locked command, got %d", len(preview.Changes))
	}
	if len(preview.Skipped) != 1 {
		t.Fatalf("expected 1 skipped record, got %d", len(preview.Skipped))
	}
}

func TestBuildKeybindImportPreviewSupportsWrappedObjectFormat(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "save", Label: "Save", Shortcut: "Ctrl+K,S", Configurable: true},
		{CommandID: "open_switch", Label: "Open", Shortcut: "Ctrl+O,K", Configurable: true},
	}
	t.Run("keybinds", func(t *testing.T) {
		jsonData := []byte(`{
		"keybinds": [
			{"command_id":"save","shortcut":"Ctrl+K,Z"}
		]
		}`)
		preview, err := buildKeybindImportPreview(records, jsonData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(preview.Errors) != 0 {
			t.Fatalf("expected no validation errors, got %#v", preview.Errors)
		}
		if preview.Result[0].Shortcut != "Ctrl+K,Z" {
			t.Fatalf("updated shortcut = %q, want Ctrl+K,Z", preview.Result[0].Shortcut)
		}
	})

	t.Run("bindings", func(t *testing.T) {
		jsonData := []byte(`{
		"bindings": [
			{"command_id":"save","shortcut":"Ctrl+K,Z"}
		]
		}`)
		preview, err := buildKeybindImportPreview(records, jsonData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(preview.Errors) != 0 {
			t.Fatalf("expected no validation errors, got %#v", preview.Errors)
		}
		if preview.Result[0].Shortcut != "Ctrl+K,Z" {
			t.Fatalf("updated shortcut = %q, want Ctrl+K,Z", preview.Result[0].Shortcut)
		}
	})

	t.Run("items", func(t *testing.T) {
		jsonData := []byte(`{
		"items": [
			{"command_id":"save","shortcut":"Ctrl+K,Z"}
		]
		}`)
		preview, err := buildKeybindImportPreview(records, jsonData)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(preview.Errors) != 0 {
			t.Fatalf("expected no validation errors, got %#v", preview.Errors)
		}
		if preview.Result[0].Shortcut != "Ctrl+K,Z" {
			t.Fatalf("updated shortcut = %q, want Ctrl+K,Z", preview.Result[0].Shortcut)
		}
	})
}

func TestBuildKeybindImportPreviewRejectsWrappedObjectWithoutSupportedWrapper(t *testing.T) {
	records := []sqlite.KeybindRecord{
		{CommandID: "save", Label: "Save", Shortcut: "Ctrl+K,S", Configurable: true},
	}
	jsonData := []byte(`{"payload":[{"command_id":"save","shortcut":"Ctrl+K,Z"}]}`)
	if _, err := buildKeybindImportPreview(records, jsonData); err == nil {
		t.Fatal("expected error for object payload without keybinds/bindings/items field")
	}
}
