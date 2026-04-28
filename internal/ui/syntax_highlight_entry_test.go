package ui

import (
	"testing"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/test"

	"ws7/internal/syntax"
)

func hasTokenKind(lines [][]syntax.Token, kind syntax.TokenKind) bool {
	for _, line := range lines {
		for _, tok := range line {
			if tok.Kind == kind {
				return true
			}
		}
	}
	return false
}

func TestSyntaxHighlightEntryUpdatesOnTyping(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(theme.DefaultTheme())
	t.Cleanup(func() { a.Quit() })

	e := newSyntaxHighlightEntry(syntax.DialectMSXBasicOfficial)
	e.SetText("10 PRIN")
	e.entry.CursorRow = 0
	e.entry.CursorColumn = len(e.entry.Text)

	e.TypedRune('T')

	if e.Text() != "10 PRINT" {
		t.Fatalf("expected typed text to be updated, got %q", e.Text())
	}
	if !hasTokenKind(e.highlights, syntax.TokenKeyword) {
		t.Fatalf("expected keyword token after typing PRINT")
	}
}

func TestSyntaxHighlightEntryUpdatesOnLoadFlow(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(theme.DefaultTheme())
	t.Cleanup(func() { a.Quit() })

	e := newSyntaxHighlightEntry(syntax.DialectMSXBasicOfficial)
	tab := &editorTab{
		entry:         e.entry,
		syntaxEntry:   e,
		syntaxDialect: syntax.DialectMSXBasicOfficial,
	}
	ui := &editorUI{}

	tab.entry.SetText("10 print\n20 rem hello")
	ui.warmupSyntaxForTab(tab)

	if !hasTokenKind(e.highlights, syntax.TokenKeyword) {
		t.Fatalf("expected keyword token after load warmup")
	}
	if !hasTokenKind(e.highlights, syntax.TokenComment) {
		t.Fatalf("expected comment token after load warmup")
	}
}

