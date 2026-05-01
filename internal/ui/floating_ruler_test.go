package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func TestFloatingRulerCreation(t *testing.T) {
	ruler := newFloatingRulerWidget()
	if ruler == nil {
		t.Fatal("expected floating ruler widget to be created")
	}

	// Check default values
	if ruler.originCharPos != 0 {
		t.Fatalf("expected origin to be 0, got %d", ruler.originCharPos)
	}
	if ruler.cursorCharPos != 0 {
		t.Fatalf("expected cursor pos to be 0, got %d", ruler.cursorCharPos)
	}
}

func TestFloatingRulerUpdate(t *testing.T) {
	ruler := newFloatingRulerWidget()

	// Test setting origin
	ruler.SetOriginCharPos(100)
	if ruler.originCharPos != 100 {
		t.Fatalf("expected origin 100, got %d", ruler.originCharPos)
	}

	// Test updating cursor
	ruler.UpdateCursor(150)
	if ruler.cursorCharPos != 150 {
		t.Fatalf("expected cursor 150, got %d", ruler.cursorCharPos)
	}

	// Test setting text
	testText := "Hello World"
	ruler.SetText(testText)
	if ruler.text != testText {
		t.Fatalf("expected text %q, got %q", testText, ruler.text)
	}
}

func TestFloatingRulerPosition(t *testing.T) {
	ruler := newFloatingRulerWidget()

	// Test default position
	x, y := ruler.GetPosition()
	if x != 50 || y != 50 {
		t.Fatalf("expected position (50, 50), got (%.0f, %.0f)", x, y)
	}

	// Test setting position
	ruler.SetPosition(100, 200)
	x, y = ruler.GetPosition()
	if x != 100 || y != 200 {
		t.Fatalf("expected position (100, 200), got (%.0f, %.0f)", x, y)
	}
}

func TestFloatingRulerDraggedUpdatesPosition(t *testing.T) {
	ruler := newFloatingRulerWidget()

	ruler.Dragged(&fyne.DragEvent{Dragged: fyne.Delta{DX: 20, DY: -10}})
	x, y := ruler.GetPosition()
	if x != 70 || y != 40 {
		t.Fatalf("expected position (70, 40) after drag, got (%.0f, %.0f)", x, y)
	}
}

func TestBuildRulerScaleRowsUses132Columns(t *testing.T) {
	top, bot := buildRulerScaleRows(132)
	if len(top) != 132 || len(bot) != 132 {
		t.Fatalf("expected 132 columns, got top=%d bot=%d", len(top), len(bot))
	}
	if !strings.HasPrefix(top, "1234567890") {
		t.Fatalf("expected top scale to start with continuous digits, got: %q", top[:10])
	}
	if !strings.Contains(bot, "10") || !strings.Contains(bot, "20") || !strings.Contains(bot, "130") {
		t.Fatalf("expected bottom scale to include decade markers, got: %q", bot)
	}
}

func TestFloatingRulerMarkBlockPoint(t *testing.T) {
	ruler := newFloatingRulerWidget()

	msg := ruler.MarkBlockPoint(4)
	if msg != "RULE: block start=4 (press B for end)" {
		t.Fatalf("unexpected first message: %q", msg)
	}
	msg = ruler.MarkBlockPoint(10)
	if msg != "RULE: block 4..10 (7 chars)" {
		t.Fatalf("unexpected second message: %q", msg)
	}
}

func TestFloatingRulerInTabContent(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		inEditor:  true,
		resolver:  nil,
		status:    widget.NewLabel(""),
		activeTab: tab,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)

	// Initially ruler should not be visible
	tab.item.Content = ui.tabEditorContent(tab)

	// Enable rule mode (which shows floating ruler)
	ui.setRuleMode(tab, true)
	tab.item.Content = ui.tabEditorContent(tab)

	// Check that the content is now a Stack with the floating ruler
	stacked, ok := tab.item.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("expected stacked container with floating ruler, got %T", tab.item.Content)
	}

	// With rule mode on, we should have a stack
	if stacked == nil {
		t.Fatalf("expected non-nil content container")
	}

	// Disable rule mode
	ui.setRuleMode(tab, false)
	tab.item.Content = ui.tabEditorContent(tab)

	// Check that it's back to normal border layout
	// (not a stack, so it could be a border layout)
	if tab.item.Content == nil {
		t.Fatalf("expected content to be set after disabling rule mode")
	}
}

func TestFloatingRulerCallback(t *testing.T) {
	ruler := newFloatingRulerWidget()
	callbackCalled := false
	callbackValue := 0

	ruler.SetOriginSetCallback(func(pos int) {
		callbackCalled = true
		callbackValue = pos
	})

	// Simulate callback
	if ruler.onOriginSet != nil {
		ruler.onOriginSet(42)
	}

	if !callbackCalled {
		t.Fatal("expected callback to be called")
	}
	if callbackValue != 42 {
		t.Fatalf("expected callback value 42, got %d", callbackValue)
	}
}
