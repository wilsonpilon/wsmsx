package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/input"
)

func newBoldTestUI(tab *editorTab) *editorUI {
	ui := &editorUI{
		inEditor:  true,
		status:    widget.NewLabel(""),
		activeTab: tab,
		entry:     tab.entry,
		ruler:     tab.ruler,
		lineNums:  tab.lineNums,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		resolver:  input.NewResolver(),
	}
	ui.bindTabEntry(tab)
	return ui
}

func TestStyleBoldTogglesOnAndOff(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newBoldTestUI(tab)

	if tab.isBold {
		t.Fatal("expected isBold=false initially")
	}

	// First toggle: bold ON
	ui.execute(input.CmdStyleBold)

	if !tab.isBold {
		t.Fatal("expected isBold=true after first toggle")
	}
	if !tab.entry.TextStyle.Bold {
		t.Fatal("expected entry TextStyle.Bold=true")
	}
	if !tab.ruler.bold {
		t.Fatal("expected ruler.bold=true")
	}
	if !tab.lineNums.bold {
		t.Fatal("expected lineNums.bold=true")
	}
	if !tab.floatingRuler.bold {
		t.Fatal("expected floatingRuler.bold=true")
	}
	if ui.status.Text == "" {
		t.Fatal("expected non-empty status after bold toggle")
	}

	// Second toggle: bold OFF
	ui.execute(input.CmdStyleBold)

	if tab.isBold {
		t.Fatal("expected isBold=false after second toggle")
	}
	if tab.entry.TextStyle != (fyne.TextStyle{}) {
		t.Fatalf("expected entry TextStyle reset, got %+v", tab.entry.TextStyle)
	}
	if tab.ruler.bold {
		t.Fatal("expected ruler.bold=false after second toggle")
	}
	if tab.lineNums.bold {
		t.Fatal("expected lineNums.bold=false after second toggle")
	}
	if tab.floatingRuler.bold {
		t.Fatal("expected floatingRuler.bold=false after second toggle")
	}
}

func TestStyleBoldChordIsCtrlPB(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := newBoldTestUI(tab)
	tab.entry.SetText("hello")

	// Trigger via chord Ctrl+P then Ctrl+B
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+P prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+P prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyB, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+P,B chord to be handled")
	}

	if !tab.isBold {
		t.Fatal("expected bold to be enabled after Ctrl+P,B chord")
	}
}
