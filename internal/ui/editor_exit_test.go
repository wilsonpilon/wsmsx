package ui

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
)

func TestUnsavedTabsCount(t *testing.T) {
	tests := []struct {
		name     string
		tabState map[*container.TabItem]*editorTab
		inEditor bool
		dirty    bool
		want     int
	}{
		{
			name:     "no tabs",
			tabState: map[*container.TabItem]*editorTab{},
			want:     0,
		},
		{
			name: "one dirty tab",
			tabState: map[*container.TabItem]*editorTab{
				&container.TabItem{}: {dirty: true},
				&container.TabItem{}: {dirty: false},
			},
			want: 1,
		},
		{
			name: "two dirty tabs",
			tabState: map[*container.TabItem]*editorTab{
				&container.TabItem{}: {dirty: true},
				&container.TabItem{}: {dirty: true},
				&container.TabItem{}: {dirty: false},
			},
			want: 2,
		},
		{
			name:     "legacy fallback when editor dirty without tab state",
			tabState: map[*container.TabItem]*editorTab{},
			inEditor: true,
			dirty:    true,
			want:     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ui := &editorUI{
				tabState: tc.tabState,
				inEditor: tc.inEditor,
				dirty:    tc.dirty,
			}
			if got := ui.unsavedTabsCount(); got != tc.want {
				t.Fatalf("unsavedTabsCount() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRequestAppExitWithMultipleUnsavedTabs(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })

	w := a.NewWindow("exit-test")
	t.Cleanup(w.Close)

	closed := false

	ui := &editorUI{
		window: w,
		closeWindow: func() {
			closed = true
		},
		tabState: map[*container.TabItem]*editorTab{
			&container.TabItem{}: {dirty: true},
			&container.TabItem{}: {dirty: true},
			&container.TabItem{}: {dirty: false},
		},
	}

	var called int
	var gotTitle string
	var gotMessage string
	ui.confirmDialog = func(title, message string, onResult func(bool), parent fyne.Window) {
		called++
		gotTitle = title
		gotMessage = message
		if parent != w {
			t.Fatalf("unexpected confirm parent window")
		}
		onResult(false)
	}

	ui.requestAppExit()

	if called != 1 {
		t.Fatalf("expected confirmation dialog to be shown once, got %d", called)
	}
	if gotTitle != "Exit MSXide" {
		t.Fatalf("unexpected dialog title: %q", gotTitle)
	}
	if !strings.Contains(gotMessage, "2 tab(s)") {
		t.Fatalf("expected message to mention dirty tab count, got %q", gotMessage)
	}
	if closed {
		t.Fatalf("window must remain open when user cancels exit")
	}

	ui.confirmDialog = func(_ string, _ string, onResult func(bool), _ fyne.Window) {
		onResult(true)
	}
	ui.requestAppExit()

	if !closed {
		t.Fatalf("window should close after confirming exit")
	}
}

func TestRequestAppExitWithoutUnsavedTabsClosesDirectly(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })

	w := a.NewWindow("exit-direct")
	t.Cleanup(w.Close)

	closed := false

	ui := &editorUI{
		window: w,
		closeWindow: func() {
			closed = true
		},
		tabState: map[*container.TabItem]*editorTab{},
	}

	ui.requestAppExit()

	if !closed {
		t.Fatalf("window should close directly when there are no unsaved tabs")
	}
}

