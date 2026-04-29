package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/basic/renum"
	"ws7/internal/input"
)

func TestMakeEditorEditMenuStartsWithBasicSubmenu(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorEditMenu()
	if len(menu.Items) < 1 {
		t.Fatalf("expected Edit menu items")
	}
	if menu.Items[0].Label != "BASIC" {
		t.Fatalf("first Edit menu item = %q, want BASIC", menu.Items[0].Label)
	}
	if menu.Items[0].ChildMenu == nil {
		t.Fatalf("expected BASIC submenu")
	}
	if len(menu.Items[0].ChildMenu.Items) != 2 {
		t.Fatalf("BASIC submenu items = %d, want 2", len(menu.Items[0].ChildMenu.Items))
	}
	if menu.Items[0].ChildMenu.Items[0].Label != "DELETE...                     Ctrl+Q,D" {
		t.Fatalf("first BASIC submenu item = %q, want DELETE...                     Ctrl+Q,D", menu.Items[0].ChildMenu.Items[0].Label)
	}
	if menu.Items[0].ChildMenu.Items[1].Label != "RENUM...                      Ctrl+Q,E" {
		t.Fatalf("second BASIC submenu item = %q, want RENUM...                      Ctrl+Q,E", menu.Items[0].ChildMenu.Items[1].Label)
	}
}

func TestHandleEditorShortcutTriggersBasicDeleteChord(t *testing.T) {
	ui := &editorUI{
		inEditor:  true,
		resolver:  input.NewResolver(),
		status:    widget.NewLabel(""),
		activeTab: &editorTab{},
	}

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+Q prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyD, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q,D to be handled")
	}
	if ui.resolver.HasPrefix() {
		t.Fatal("expected prefix state to be cleared after Ctrl+Q,D")
	}
	if got := ui.status.Text; got != "Ctrl+Q,D ✓" {
		t.Fatalf("status = %q, want %q", got, "Ctrl+Q,D ✓")
	}
}

func TestHandleEditorShortcutTriggersBasicRenumChord(t *testing.T) {
	ui := &editorUI{
		inEditor:  true,
		resolver:  input.NewResolver(),
		status:    widget.NewLabel(""),
		activeTab: &editorTab{},
	}

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+Q prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q,E to be handled")
	}
	if ui.resolver.HasPrefix() {
		t.Fatal("expected prefix state to be cleared after Ctrl+Q,E")
	}
	if got := ui.status.Text; got != "Ctrl+Q,E ✓" {
		t.Fatalf("status = %q, want %q", got, "Ctrl+Q,E ✓")
	}
}

func TestFormatBasicDeleteWarnings(t *testing.T) {
	refs := []renum.Reference{
		{SourceLine: 10, Command: "GOTO", Target: 200, Category: renum.WarningCategoryFlow, Severity: renum.WarningSeverityWarning},
		{SourceLine: 20, Command: "LIST", Target: 200, Category: renum.WarningCategoryListing, Severity: renum.WarningSeverityInfo},
	}
	got := formatBasicDeleteWarnings(200, 220, refs)
	if !strings.Contains(got, "Deletion blocked: lines 200 to 220") {
		t.Fatalf("warning text missing delete range: %q", got)
	}
	if !strings.Contains(got, "Flow references (severity: warning)") {
		t.Fatalf("warning text missing flow section: %q", got)
	}
	if !strings.Contains(got, "Listing references (severity: info)") {
		t.Fatalf("warning text missing listing section: %q", got)
	}
	if !strings.Contains(got, "Source 10: GOTO 200") {
		t.Fatalf("warning text missing flow reference: %q", got)
	}
	if !strings.Contains(got, "Source 20: LIST 200") {
		t.Fatalf("warning text missing listing reference: %q", got)
	}
}

func TestFormatBasicDeleteWarningsTruncatesLongList(t *testing.T) {
	refs := make([]renum.Reference, 0, 13)
	for i := 0; i < 13; i++ {
		refs = append(refs, renum.Reference{SourceLine: i + 1, Command: "GOTO", Target: i + 100})
	}
	got := formatBasicDeleteWarnings(100, 200, refs)
	if !strings.Contains(got, "and 1 more item(s)") {
		t.Fatalf("expected truncation note, got %q", got)
	}
}

func TestResolveBasicDeleteScope(t *testing.T) {
	ui := &editorUI{activeTab: &editorTab{entry: newCursorEntry()}}
	ui.activeTab.entry.SetText("10 PRINT \"A\"\n\n20 PRINT \"B\"\n30 PRINT \"C\"\n")

	ui.activeTab.entry.CursorRow = 0
	from, to, err := ui.resolveBasicDeleteScope("Current line only", "", "", 10, 30)
	if err != nil || from != 10 || to != 10 {
		t.Fatalf("Current line only = (%d,%d,%v), want (10,10,nil)", from, to, err)
	}

	ui.activeTab.entry.CursorRow = 1
	from, to, err = ui.resolveBasicDeleteScope("Cursor to end", "", "", 10, 30)
	if err != nil || from != 20 || to != 30 {
		t.Fatalf("Cursor to end = (%d,%d,%v), want (20,30,nil)", from, to, err)
	}

	ui.activeTab.entry.CursorRow = 1
	from, to, err = ui.resolveBasicDeleteScope("Cursor to beginning", "", "", 10, 30)
	if err != nil || from != 10 || to != 10 {
		t.Fatalf("Cursor to beginning = (%d,%d,%v), want (10,10,nil)", from, to, err)
	}

	from, to, err = ui.resolveBasicDeleteScope("Line range", "20", "30", 10, 30)
	if err != nil || from != 20 || to != 30 {
		t.Fatalf("Line range = (%d,%d,%v), want (20,30,nil)", from, to, err)
	}
}

func TestResolveBasicDeleteScopeRejectsUnnumberedCurrentLine(t *testing.T) {
	ui := &editorUI{activeTab: &editorTab{entry: newCursorEntry()}, status: widget.NewLabel("")}
	ui.activeTab.entry.SetText("REM test\n20 PRINT \"A\"\n")
	ui.activeTab.entry.CursorRow = 0
	_, _, err := ui.resolveBasicDeleteScope("Current line only", "", "", 20, 20)
	if err == nil || !strings.Contains(err.Error(), "cursor is not on a numbered BASIC line") {
		t.Fatalf("unexpected error: %v", err)
	}
}
