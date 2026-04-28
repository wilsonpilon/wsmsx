package ui

import (
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/syntax"
)

func makeSplitViewTestTab(name string) *editorTab {
	syntaxEntry := newSyntaxHighlightEntry(syntax.DialectMSXBasicOfficial)
	tab := &editorTab{
		entry:         syntaxEntry.entry,
		syntaxEntry:   syntaxEntry,
		ruler:         newRulerWidget(),
		lineNums:      newLineNumbersWidget(),
		status:        widget.NewLabel(""),
		blockTag:      widget.NewLabel(""),
		clipTag:       widget.NewLabel(""),
		syntaxTag:     widget.NewLabel(""),
		syntaxDialect: syntax.DialectMSXBasicOfficial,
	}
	tab.item = container.NewTabItem(name, widget.NewLabel("placeholder"))
	return tab
}

func TestSetSyntaxSplitViewRebuildsTabContent(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })

	tabA := makeSplitViewTestTab("A")
	tabB := makeSplitViewTestTab("B")

	ui := &editorUI{
		fyneApp:         a,
		tabState:        map[*container.TabItem]*editorTab{tabA.item: tabA, tabB.item: tabB},
		syntaxSplitView: false,
		status:          widget.NewLabel(""),
	}

	tabA.item.Content = ui.tabEditorContent(tabA)
	tabB.item.Content = ui.tabEditorContent(tabB)
	beforeA := tabA.item.Content
	beforeB := tabB.item.Content

	ui.setSyntaxSplitView(true)

	if !ui.syntaxSplitView {
		t.Fatalf("expected split view to be enabled")
	}
	if beforeA == tabA.item.Content {
		t.Fatalf("expected tab A content to be rebuilt when enabling split view")
	}
	if beforeB == tabB.item.Content {
		t.Fatalf("expected tab B content to be rebuilt when enabling split view")
	}
	if ui.status.Text != "View: Split Syntax Preview" {
		t.Fatalf("unexpected status text when enabling split view: %q", ui.status.Text)
	}

	splitA := tabA.item.Content
	splitB := tabB.item.Content
	ui.setSyntaxSplitView(false)

	if ui.syntaxSplitView {
		t.Fatalf("expected split view to be disabled")
	}
	if splitA == tabA.item.Content {
		t.Fatalf("expected tab A content to be rebuilt when disabling split view")
	}
	if splitB == tabB.item.Content {
		t.Fatalf("expected tab B content to be rebuilt when disabling split view")
	}
	if ui.status.Text != "View: Inline Syntax Highlight" {
		t.Fatalf("unexpected status text when disabling split view: %q", ui.status.Text)
	}
}

func TestSetSyntaxSplitViewNoOpWhenStateUnchanged(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		tabState:        map[*container.TabItem]*editorTab{tab.item: tab},
		syntaxSplitView: false,
	}

	tab.item.Content = ui.tabEditorContent(tab)
	before := tab.item.Content

	ui.setSyntaxSplitView(false)

	if before != tab.item.Content {
		t.Fatalf("expected content to remain unchanged when split view state does not change")
	}
}

