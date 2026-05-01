package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/input"
)

func newConvertCaseTestUI(tab *editorTab) *editorUI {
	ui := &editorUI{
		inEditor:  true,
		status:    widget.NewLabel(""),
		activeTab: tab,
		entry:     tab.entry,
		ruler:     tab.ruler,
		lineNums:  tab.lineNums,
		blockTag:  tab.blockTag,
		clipTag:   tab.clipTag,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		resolver:  input.NewResolver(),
	}
	ui.bindTabEntry(tab)
	return ui
}

func TestCapitalizeText(t *testing.T) {
	got := capitalizeText("hELLo, wORLD 123abc")
	want := "Hello, World 123abc"
	if got != want {
		t.Fatalf("unexpected capitalize result: got=%q want=%q", got, want)
	}
}

func TestFindSelectionRangePrefersCursorEdge(t *testing.T) {
	text := "AA test BB test CC"
	selected := "test"
	cursor := strings.Index(text, selected) + len(selected)

	start, end, ok := findSelectionRange(text, selected, cursor)
	if !ok {
		t.Fatal("expected selection range")
	}
	if got := text[start:end]; got != "test" {
		t.Fatalf("unexpected selected text: %q", got)
	}
	if start != 3 {
		t.Fatalf("expected first 'test' start=3, got=%d", start)
	}
}

func TestCurrentLineRange(t *testing.T) {
	text := "one\nTwo TEST\nthree"
	cursor := strings.Index(text, "TEST")
	start, end, ok := currentLineRange(text, cursor)
	if !ok {
		t.Fatal("expected current line range")
	}
	if got := text[start:end]; got != "Two TEST" {
		t.Fatalf("unexpected line slice: %q", got)
	}
}

func TestConvertCaseUsesMarkedBlockFirst(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("aa bb cc")
	ui.applyCursorPosition(0, 0)

	tab.hasBlockBegin = true
	tab.blockBegin = 3 // b from "bb"
	tab.hasBlockEnd = true
	tab.blockEnd = 5

	ui.execute(input.CmdConvertUppercase)

	if got := ui.entry.Text; got != "aa BB cc" {
		t.Fatalf("unexpected text after uppercase block conversion: %q", got)
	}
	if !strings.Contains(ui.status.Text, "applied to block") {
		t.Fatalf("unexpected status: %q", ui.status.Text)
	}
}

func TestConvertCaseFallsBackToCurrentLine(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("UPPER\nTwo TEST\nTHREE")
	ui.applyCursorPosition(1, 2)

	ui.execute(input.CmdConvertLowercase)

	if got := ui.entry.Text; got != "UPPER\ntwo test\nTHREE" {
		t.Fatalf("unexpected text after lowercase line conversion: %q", got)
	}
	if !strings.Contains(ui.status.Text, "applied to current line") {
		t.Fatalf("unexpected status: %q", ui.status.Text)
	}
}

func TestConvertCaseEmptyBlockMessage(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("abc")
	ui.applyCursorPosition(0, 0)

	tab.hasBlockBegin = true
	tab.blockBegin = 1
	tab.hasBlockEnd = true
	tab.blockEnd = 1

	ui.execute(input.CmdConvertUppercase)

	if got := ui.entry.Text; got != "abc" {
		t.Fatalf("text should remain unchanged, got %q", got)
	}
	if got := ui.status.Text; got != "Ctrl+K,\": empty block (B and K at same position)" {
		t.Fatalf("unexpected status: %q", got)
	}
}

func TestConvertCaseEmptyLineMessage(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("\nabc")
	ui.applyCursorPosition(0, 0)

	ui.execute(input.CmdConvertUppercase)

	if got := ui.entry.Text; got != "\nabc" {
		t.Fatalf("text should remain unchanged, got %q", got)
	}
	if got := ui.status.Text; got != "Ctrl+K,\": current line is empty" {
		t.Fatalf("unexpected status: %q", got)
	}
}

func TestConvertCaseEmptyDocumentMessage(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("")
	ui.applyCursorPosition(0, 0)

	ui.execute(input.CmdConvertUppercase)

	if got := ui.status.Text; got != "Ctrl+K,\": document is empty" {
		t.Fatalf("unexpected status: %q", got)
	}
}

func TestConvertCaseLowercaseChordInEmptyDocumentMessage(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("")
	ui.applyCursorPosition(0, 0)

	ui.execute(input.CmdConvertLowercase)

	if got := ui.status.Text; got != "Ctrl+K,': document is empty" {
		t.Fatalf("unexpected status: %q", got)
	}
}

func TestConvertCaseCapitalizeChordInEmptyLineMessage(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newConvertCaseTestUI(tab)
	tab.entry.SetText("\nabc")
	ui.applyCursorPosition(0, 0)

	ui.execute(input.CmdConvertCapitalize)

	if got := ui.entry.Text; got != "\nabc" {
		t.Fatalf("text should remain unchanged, got %q", got)
	}
	if got := ui.status.Text; got != "Ctrl+K,.: current line is empty" {
		t.Fatalf("unexpected status: %q", got)
	}
}
