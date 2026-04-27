package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/config"
	"ws7/internal/input"
	"ws7/internal/store/sqlite"
)

var errSaveCanceled = errors.New("save canceled")

const ctrlKTimeout = 2 * time.Second

type editorTab struct {
	item     *container.TabItem
	entry    *cursorEntry
	ruler    *rulerWidget
	lineNums *lineNumbersWidget
	status   *widget.Label
	blockTag *widget.Label
	clipTag  *widget.Label

	name      string
	filePath  string
	dirty     bool
	cursorRow int
	cursorCol int
	topLine   int

	blockBegin    int
	blockEnd      int
	hasBlockBegin bool
	hasBlockEnd   bool
}

type editorUI struct {
	fyneApp  fyne.App
	window   fyne.Window
	entry    *cursorEntry
	ruler    *rulerWidget
	lineNums *lineNumbersWidget
	status   *widget.Label
	blockTag *widget.Label
	clipTag  *widget.Label
	resolver *input.Resolver
	store    *sqlite.Store
	browser  *fileBrowser

	filePath        string
	dirty           bool
	inEditor        bool
	cursorRow       int
	cursorCol       int
	topLine         int
	prefixTimeoutID uint64
	prefixExpired   uint32

	tabs         *container.DocTabs
	tabState     map[*container.TabItem]*editorTab
	activeTab    *editorTab
	untitledSeed int

	internalBlockClipboard string
}

func Run() error {
	a := app.NewWithID("ws7.editor")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	fontPath := filepath.Join(cwd, "res", "SourceCodePro-Bold.ttf")
	if th, thErr := newSourceCodeProTheme(fontPath); thErr == nil {
		a.Settings().SetTheme(th)
	}

	dbPath, err := config.DBPath()
	if err != nil {
		return err
	}
	store, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	ui := &editorUI{
		fyneApp:  a,
		window:   a.NewWindow("WS7 Editor"),
		resolver: input.NewResolver(),
		store:    store,
		tabState: map[*container.TabItem]*editorTab{},
	}

	// Resolve start directory: last used or cwd
	startDir, _ := store.GetSetting(context.Background(), "last_dir")
	if startDir == "" || !dirExists(startDir) {
		startDir = cwd
	}

	ui.browser = newFileBrowser(startDir, func(path string) {
		ui.openInEditor(path)
	})
	ui.browser.onDirChange = func(dir string) {
		_ = store.SetSetting(context.Background(), "last_dir", dir)
	}

	ui.window.Resize(fyne.NewSize(980, 680))
	ui.showBrowser()
	ui.window.ShowAndRun()
	return nil
}

func (e *editorUI) ensureTabs() {
	if e.tabs != nil {
		return
	}
	e.tabs = container.NewDocTabs()
	e.tabs.SetTabLocation(container.TabLocationTop)
	e.tabs.CloseIntercept = func(item *container.TabItem) {
		tab := e.tabState[item]
		if tab == nil {
			return
		}
		e.requestCloseTab(tab)
	}
	e.tabs.OnSelected = func(item *container.TabItem) {
		tab := e.tabState[item]
		if tab == nil {
			return
		}
		e.activeTab = tab
		e.entry = tab.entry
		e.ruler = tab.ruler
		e.lineNums = tab.lineNums
		e.status = tab.status
		e.blockTag = tab.blockTag
		e.clipTag = tab.clipTag
		e.filePath = tab.filePath
		e.dirty = tab.dirty
		e.cursorRow = tab.cursorRow
		e.cursorCol = tab.cursorCol
		e.topLine = tab.topLine
		e.updateBlockIndicator()
		e.updateInternalClipboardIndicator()
		e.updateTitle()
		e.syncLineNumbers()
	}
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	clean := filepath.Clean(path)
	abs, err := filepath.Abs(clean)
	if err == nil {
		clean = abs
	}
	return strings.ToLower(clean)
}

func (e *editorUI) findTabByPath(path string) *editorTab {
	want := normalizePath(path)
	if want == "" {
		return nil
	}
	for _, tab := range e.tabState {
		if normalizePath(tab.filePath) == want {
			return tab
		}
	}
	return nil
}

func (e *editorUI) bindTabEntry(tab *editorTab) {
	tab.entry.Wrapping = fyne.TextWrapOff
	tab.entry.SetMinRowsVisible(30)
	tab.entry.OnChanged = func(_ string) {
		tab.dirty = true
		if e.activeTab == tab {
			e.dirty = true
			e.updateTitle()
			if e.inEditor {
				e.syncLineNumbers()
			}
		} else {
			e.refreshTabTitle(tab)
		}
	}
	tab.entry.onCursorMoved = func(row, col int) {
		tab.cursorRow = row
		tab.cursorCol = col
		tab.ruler.UpdateCursor(row, col)
		if e.activeTab == tab {
			e.cursorRow = row
			e.cursorCol = col
			e.ruler.UpdateCursor(row, col)
			e.syncLineNumbers()
			e.updateCursorStatus()
		}
	}
	tab.entry.onViewportOffset = func(offsetY float32) {
		if e.activeTab == tab && e.inEditor {
			e.applyViewportOffset(offsetY)
		}
	}
	tab.entry.onRuneBeforeInput = func(r rune) bool {
		if !e.inEditor || e.activeTab != tab {
			return false
		}
		e.consumePrefixTimeoutIfNeeded()
		if !e.resolver.HasPrefix() {
			return false
		}
		e.handleCtrl(strings.ToUpper(string(r)))
		return true
	}
	tab.entry.onShortcut = func(shortcut fyne.Shortcut) bool {
		if e.activeTab != tab {
			return false
		}
		return e.handleEditorShortcut(shortcut)
	}
}

func (e *editorUI) nextUntitledName() string {
	e.untitledSeed++
	if e.untitledSeed == 1 {
		return "untitled"
	}
	return fmt.Sprintf("untitled-%d", e.untitledSeed)
}

func (e *editorUI) newEditorTab() *editorTab {
	name := e.nextUntitledName()
	tab := &editorTab{
		entry:    newCursorEntry(),
		ruler:    newRulerWidget(),
		lineNums: newLineNumbersWidget(),
		status:   widget.NewLabel(""),
		blockTag: widget.NewLabel(""),
		clipTag:  widget.NewLabel(""),
		name:     name,
	}
	tab.blockTag.TextStyle = fyne.TextStyle{Bold: true}
	tab.clipTag.TextStyle = fyne.TextStyle{Bold: true}
	rightIndicators := container.NewHBox(tab.blockTag, tab.clipTag)
	statusBar := container.NewBorder(nil, nil, nil, rightIndicators, tab.status)
	e.bindTabEntry(tab)
	tab.item = container.NewTabItem(name, container.NewBorder(tab.ruler, statusBar, tab.lineNums, nil, tab.entry))
	e.tabState[tab.item] = tab
	e.tabs.Append(tab.item)
	e.tabs.Select(tab.item)
	e.refreshTabTitle(tab)
	return tab
}

func (e *editorUI) refreshTabTitle(tab *editorTab) {
	if tab == nil || tab.item == nil {
		return
	}
	base := tab.name
	if tab.filePath != "" {
		base = filepath.Base(tab.filePath)
	}
	if tab.dirty {
		base += "*"
		tab.item.Icon = theme.WarningIcon()
	} else {
		tab.item.Icon = theme.DocumentIcon()
	}
	tab.item.Text = base
	if e.tabs != nil {
		e.tabs.Refresh()
	}
}

func (e *editorUI) closeActiveTab() {
	if e.tabs == nil || e.activeTab == nil {
		e.showBrowser()
		return
	}
	e.requestCloseTab(e.activeTab)
}

func (e *editorUI) requestCloseTab(tab *editorTab) {
	if tab == nil {
		return
	}
	if tab.dirty {
		dialog.ShowConfirm(
			"Close Tab",
			"This tab has unsaved changes. Close anyway?",
			func(ok bool) {
				if ok {
					e.closeTabImmediately(tab)
				}
			},
			e.window,
		)
		return
	}
	e.closeTabImmediately(tab)
}

func (e *editorUI) closeTabImmediately(tab *editorTab) {
	if e.tabs == nil || tab == nil || tab.item == nil {
		return
	}

	selectedBefore := e.tabs.Selected()
	removeIdx := -1
	for i, item := range e.tabs.Items {
		if item == tab.item {
			removeIdx = i
			break
		}
	}
	if removeIdx < 0 {
		return
	}

	delete(e.tabState, tab.item)
	e.tabs.Remove(tab.item)
	if len(e.tabs.Items) == 0 {
		e.activeTab = nil
		e.showBrowser()
		return
	}

	if selectedBefore != tab.item {
		if selectedBefore != nil {
			e.tabs.Select(selectedBefore)
		} else {
			e.tabs.SelectIndex(0)
		}
		e.window.Canvas().Focus(e.entry)
		return
	}

	if removeIdx < len(e.tabs.Items) {
		e.tabs.SelectIndex(removeIdx) // select right neighbor
	} else {
		e.tabs.SelectIndex(len(e.tabs.Items) - 1) // fallback left neighbor
	}
	e.window.Canvas().Focus(e.entry)
}

// ── View switching ────────────────────────────────────────────────────────────

func (e *editorUI) showBrowser() {
	e.inEditor = false
	e.resetPrefixState()
	e.window.SetMainMenu(e.makeOpeningMenu())
	e.window.SetTitle("WS7 Editor - Opening Menu")
	e.window.SetContent(e.browser.Content)
	e.window.Canvas().Focus(e.browser.list)
}

func (e *editorUI) showEditor(path string) {
	e.inEditor = true
	e.resetPrefixState()
	e.window.SetMainMenu(e.makeEditorMenu())
	e.ensureTabs()
	if len(e.tabs.Items) == 0 {
		e.newEditorTab()
	}
	e.window.SetContent(e.tabs)
	if path == "" {
		e.updateTitle()
		e.window.Canvas().Focus(e.entry)
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}
	e.entry.SetText(string(data))
	e.filePath = path
	e.dirty = false
	e.cursorRow = 0
	e.cursorCol = 0
	e.topLine = 0
	e.entry.CursorRow = 0
	e.entry.CursorColumn = 0
	e.ruler.UpdateCursor(0, 0)
	e.syncLineNumbers()
	if e.activeTab != nil {
		e.activeTab.filePath = path
		e.activeTab.dirty = false
		e.activeTab.cursorRow = 0
		e.activeTab.cursorCol = 0
		e.activeTab.topLine = 0
		e.refreshTabTitle(e.activeTab)
	}
	_ = e.store.TouchRecentFile(context.Background(), path)
	_ = e.store.SetSetting(context.Background(), "last_file", path)
	_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(path))
	e.browser.loadDir(filepath.Dir(path)) // keep browser in sync
	e.updateTitle()
	e.window.Canvas().Focus(e.entry)
}

func (e *editorUI) openInEditor(path string) {
	if existing := e.findTabByPath(path); existing != nil {
		e.showEditor("")
		e.tabs.Select(existing.item)
		e.window.Canvas().Focus(e.entry)
		if e.status != nil {
			e.status.SetText("File already open: switched to existing tab")
		}
		return
	}
	if e.inEditor {
		e.showEditor(path)
		return
	}
	e.showEditor(path)
}

// ── New file ─────────────────────────────────────────────────────────────────

func (e *editorUI) newFile() {
	if !e.inEditor {
		e.showEditor("")
		return
	}
	e.newEditorTab()
	e.showEditor("")
}

// ── Ctrl key bindings ─────────────────────────────────────────────────────────

func (e *editorUI) bindCtrlKeys() {
	canvas := e.window.Canvas()

	// Letters
	for _, key := range []fyne.KeyName{
		fyne.KeyA, fyne.KeyB, fyne.KeyC, fyne.KeyD, fyne.KeyE,
		fyne.KeyF, fyne.KeyG, fyne.KeyI, fyne.KeyJ, fyne.KeyK,
		fyne.KeyL, fyne.KeyN, fyne.KeyO, fyne.KeyP, fyne.KeyQ,
		fyne.KeyR, fyne.KeyS, fyne.KeyT, fyne.KeyU, fyne.KeyV,
		fyne.KeyW, fyne.KeyX, fyne.KeyY, fyne.KeyZ,
	} {
		k := key
		canvas.AddShortcut(
			&desktop.CustomShortcut{KeyName: k, Modifier: fyne.KeyModifierControl},
			func(_ fyne.Shortcut) {
				if e.inEditor {
					e.handleCtrl(strings.ToUpper(string(k)))
				}
			},
		)
	}

	// Digit keys 0–9 (used for markers after Ctrl+K or Ctrl+Q prefix)
	for _, key := range []fyne.KeyName{
		fyne.Key0, fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4,
		fyne.Key5, fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9,
	} {
		k := key
		canvas.AddShortcut(
			&desktop.CustomShortcut{KeyName: k, Modifier: fyne.KeyModifierControl},
			func(_ fyne.Shortcut) {
				if e.inEditor && e.resolver.HasPrefix() {
					e.handleCtrl(string(k))
				}
			},
		)
	}

	// Bracket keys [ ] (used for clipboard after Ctrl+K prefix)
	for _, pair := range []struct {
		key fyne.KeyName
		ch  string
	}{
		{fyne.KeyLeftBracket, "["},
		{fyne.KeyRightBracket, "]"},
	} {
		p := pair
		canvas.AddShortcut(
			&desktop.CustomShortcut{KeyName: p.key, Modifier: fyne.KeyModifierControl},
			func(_ fyne.Shortcut) {
				if e.inEditor && e.resolver.HasPrefix() {
					e.handleCtrl(p.ch)
				}
			},
		)
	}

	// ? key (used for status / change printer after prefix)
	for _, mod := range []fyne.KeyModifier{
		fyne.KeyModifierControl,
		fyne.KeyModifierControl | fyne.KeyModifierShift,
	} {
		m := mod
		canvas.AddShortcut(
			&desktop.CustomShortcut{KeyName: fyne.KeySlash, Modifier: m},
			func(_ fyne.Shortcut) {
				if e.inEditor {
					e.handleCtrl("?")
				}
			},
		)
	}
}

func (e *editorUI) handleEditorShortcut(shortcut fyne.Shortcut) bool {
	if !e.inEditor {
		return false
	}

	custom, ok := shortcut.(*desktop.CustomShortcut)
	if !ok {
		return false
	}

	mods := custom.Modifier
	ctrlOnly := mods == fyne.KeyModifierControl
	ctrlShift := mods == (fyne.KeyModifierControl | fyne.KeyModifierShift)
	if !ctrlOnly && !ctrlShift {
		return false
	}

	// Custom direct shortcut: close current tab.
	// If a resolver prefix is active (e.g. Ctrl+K then W), keep chord behavior.
	if custom.KeyName == fyne.KeyW && ctrlOnly && !e.resolver.HasPrefix() {
		e.requestCloseTab(e.activeTab)
		return true
	}

	switch custom.KeyName {
	case fyne.KeyA, fyne.KeyB, fyne.KeyC, fyne.KeyD, fyne.KeyE,
		fyne.KeyF, fyne.KeyG, fyne.KeyI, fyne.KeyJ, fyne.KeyK,
		fyne.KeyL, fyne.KeyN, fyne.KeyO, fyne.KeyP, fyne.KeyQ,
		fyne.KeyR, fyne.KeyS, fyne.KeyT, fyne.KeyU, fyne.KeyV,
		fyne.KeyW, fyne.KeyX, fyne.KeyY, fyne.KeyZ:
		e.handleCtrl(strings.ToUpper(string(custom.KeyName)))
		return true
	case fyne.Key0, fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4,
		fyne.Key5, fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9:
		if e.resolver.HasPrefix() {
			e.handleCtrl(string(custom.KeyName))
			return true
		}
	case fyne.KeyLeftBracket:
		if e.resolver.HasPrefix() {
			e.handleCtrl("[")
			return true
		}
	case fyne.KeyRightBracket:
		if e.resolver.HasPrefix() {
			e.handleCtrl("]")
			return true
		}
	case fyne.KeySlash:
		e.handleCtrl("?")
		return true
	}

	return false
}

func (e *editorUI) handleCtrl(letter string) {
	e.consumePrefixTimeoutIfNeeded()

	prevPrefix := e.resolver.CurrentPrefix()
	cmd, pending, err := e.resolver.Resolve(letter)
	if err != nil {
		e.status.SetText("⚠ " + err.Error())
		return
	}
	if pending {
		e.startPrefixTimeout()
		switch cmd {
		case input.CmdPrefixK:
			e.status.SetText("Ctrl+K ✓  waiting for next key...")
		case input.CmdPrefixKQ:
			e.status.SetText("Ctrl+K,Q ✓  waiting: X=Exit...")
		case input.CmdPrefixO:
			e.status.SetText("Ctrl+O ✓  waiting for next key...")
		case input.CmdPrefixON:
			e.status.SetText("Ctrl+O,N ✓  waiting: D=Note  V=Convert...")
		case input.CmdPrefixP:
			e.status.SetText("Ctrl+P ✓  waiting for next key...")
		case input.CmdPrefixQ:
			e.status.SetText("Ctrl+Q ✓  waiting for next key...")
		case input.CmdPrefixQN:
			e.status.SetText("Ctrl+Q,N ✓  waiting: G=GoToNote...")
		default:
			e.status.SetText("Ctrl+? ✓  waiting for next key...")
		}
		return
	}
	// Chord complete — show the completed chord in status before executing
	e.showCompletedChordStatus(cmd, prevPrefix, letter)
	e.resetPrefixState()
	e.execute(cmd)
}

// showCompletedChordStatus sets a brief status message for a completed chord.
func (e *editorUI) showCompletedChordStatus(cmd input.Command, prevPrefix, letter string) {
	if label, ok := cmdChordLabel[cmd]; ok {
		e.status.SetText(label + " ✓")
		return
	}
	// Fallback: build from prefix + letter
	if prevPrefix != "" {
		e.status.SetText("Ctrl+" + prevPrefix + "," + strings.ToUpper(letter) + " ✓")
	} else {
		e.status.SetText("Ctrl+" + strings.ToUpper(letter) + " ✓")
	}
}

// cmdChordLabel maps completed commands to a human-readable chord string.
var cmdChordLabel = map[input.Command]string{
	input.CmdSave:              "Ctrl+K,S",
	input.CmdSaveAs:            "Ctrl+K,T",
	input.CmdSaveAndClose:      "Ctrl+K,D",
	input.CmdPrint:             "Ctrl+K,P",
	input.CmdFileCopy:          "Ctrl+K,O",
	input.CmdFileDelete:        "Ctrl+K,J",
	input.CmdFileRename:        "Ctrl+K,E",
	input.CmdChangeDirectory:   "Ctrl+K,L",
	input.CmdRunPSCommand:      "Ctrl+K,F",
	input.CmdMarkBlockBegin:    "Ctrl+K,B",
	input.CmdMarkBlockEnd:      "Ctrl+K,K",
	input.CmdMoveBlock:         "Ctrl+K,V",
	input.CmdCopyBlock:         "Ctrl+K,C",
	input.CmdDeleteBlock:       "Ctrl+K,Y",
	input.CmdExit:              "Ctrl+K,Q,X",
	input.CmdOpenSwitch:        "Ctrl+O,K",
	input.CmdStatus:            "Ctrl+O,?",
	input.CmdAutoAlign:         "Ctrl+O,A",
	input.CmdChangePrinter:     "Ctrl+P,?",
	input.CmdFind:              "Ctrl+Q,F",
	input.CmdFindReplace:       "Ctrl+Q,A",
	input.CmdGoToChar:          "Ctrl+Q,G",
	input.CmdGoToPage:          "Ctrl+Q,I",
	input.CmdGoDocBegin:        "Ctrl+Q,R",
	input.CmdGoDocEnd:          "Ctrl+Q,C",
	input.CmdDeleteLineRight:   "Ctrl+Q,Y",
	input.CmdGoPrevPosition:    "Ctrl+Q,P",
	input.CmdGoLastFindReplace: "Ctrl+Q,V",
	input.CmdGoBlockBegin:      "Ctrl+Q,B",
	input.CmdGoBlockEnd:        "Ctrl+Q,K",
}

func (e *editorUI) startPrefixTimeout() {
	id := atomic.AddUint64(&e.prefixTimeoutID, 1)
	atomic.StoreUint32(&e.prefixExpired, 0)
	time.AfterFunc(ctrlKTimeout, func() {
		if atomic.LoadUint64(&e.prefixTimeoutID) != id {
			return
		}
		atomic.StoreUint32(&e.prefixExpired, 1)
	})
}

func (e *editorUI) resetPrefixState() {
	atomic.AddUint64(&e.prefixTimeoutID, 1)
	atomic.StoreUint32(&e.prefixExpired, 0)
	e.resolver.ClearPrefix()
}

func (e *editorUI) consumePrefixTimeoutIfNeeded() {
	if atomic.SwapUint32(&e.prefixExpired, 0) == 1 {
		e.resolver.ClearPrefix()
		e.status.SetText("⚠ Ctrl prefix timed out")
	}
}

func (e *editorUI) execute(cmd input.Command) {
	// ── Marker commands (dynamic) ─────────────────────────────────────────────
	if digit, ok := input.IsSetMarker(cmd); ok {
		e.cmdSetMarker(digit)
		return
	}
	if digit, ok := input.IsGoToMarker(cmd); ok {
		e.cmdGoToMarker(digit)
		return
	}

	switch cmd {
	// ── Cursor movement ───────────────────────────────────────────────────────
	case input.CmdCursorLeft:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
		e.status.SetText("Ctrl+S: left")
	case input.CmdCursorRight:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
		e.status.SetText("Ctrl+D: right")
	case input.CmdCursorUp:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyUp})
		e.status.SetText("Ctrl+E: up")
	case input.CmdCursorDown:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
		e.status.SetText("Ctrl+X: down")
	case input.CmdPageUp:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyPageUp})
		e.status.SetText("Ctrl+R: page up")
	case input.CmdPageDown:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyPageDown})
		e.status.SetText("Ctrl+C: page down")

	// ── Edit / delete ─────────────────────────────────��───────────────────────
	case input.CmdDeleteLine:
		e.deleteCurrentLine()
		e.status.SetText("Ctrl+Y: line deleted")
	case input.CmdDeleteWord:
		e.cmdDeleteWordRight()
		e.status.SetText("Ctrl+T: word deleted")
	case input.CmdDeleteLineRight:
		e.cmdDeleteLineRight()
		e.status.SetText("Ctrl+Q,Y: text right deleted")
	case input.CmdDeleteLineLeft:
		e.cmdDeleteLineLeft()
		e.status.SetText("Ctrl+Q,DEL: text left deleted")
	case input.CmdDeleteBlock:
		e.cmdDeleteBlockMarked()
	case input.CmdUndo:
		e.entry.TypedShortcut(&fyne.ShortcutUndo{})
		e.status.SetText("Ctrl+U: undo")
	case input.CmdNewTab:
		e.newFile()
		e.status.SetText("Ctrl+N: new untitled tab")
	case input.CmdScrollUp:
		e.status.SetText("Ctrl+W: scroll up (next block)")
	case input.CmdInsertLine:
		e.cmdInsertLineAtCursor()
		e.status.SetText("Insert Line (legacy)")
	case input.CmdInsertMode:
		e.status.SetText("Ctrl+V: insert/overtype (next block)")

	// ── Block marking ─────────────────────────────────────────────────────────
	case input.CmdMarkBlockBegin:
		e.cmdMarkBlockBegin()
	case input.CmdMarkBlockEnd:
		e.cmdMarkBlockEnd()
	case input.CmdMarkPreviousBlock:
		e.cmdNotImplemented("Mark Previous Block (Ctrl+K,U)")

	// ── Block move / copy ─────────────────────────────────────────────────────
	case input.CmdMoveBlock:
		e.cmdMoveBlockMarked()
	case input.CmdMoveBlockOtherWin:
		e.cmdNotImplemented("Move Block from Other Window (Ctrl+K,G)")
	case input.CmdCopyBlock:
		e.cmdCopyBlockMarked()
	case input.CmdCopyBlockOtherWin:
		e.cmdNotImplemented("Copy Block from Other Window (Ctrl+K,A)")
	case input.CmdCopyFromClipboard:
		e.entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: e.window.Clipboard()})
		e.status.SetText("Ctrl+K,[: pasted from clipboard")
	case input.CmdCopyToClipboard:
		e.entry.TypedShortcut(&fyne.ShortcutCopy{})
		e.status.SetText("Ctrl+K,]: copied to clipboard")
	case input.CmdCopyToFile:
		e.cmdFileCopy()

	// ── Block settings ────────────────────────────────────────────────────────
	case input.CmdColumnBlockMode:
		e.cmdNotImplemented("Column Block Mode (Ctrl+K,N)")
	case input.CmdColumnReplaceMode:
		e.cmdNotImplemented("Column Replace Mode (Ctrl+K,I)")

	// ── Find / navigate ───────────────────────────────────────────────────────
	case input.CmdFind:
		e.cmdFind()
	case input.CmdFindReplace:
		e.cmdFindReplace()
	case input.CmdRepeatFind:
		e.cmdRepeatFind()
	case input.CmdGoToChar:
		e.cmdGoToChar()
	case input.CmdGoToPage:
		e.cmdGoToPage()
	case input.CmdGoToFontTag:
		e.cmdNotImplemented("Go to Font Tag (Ctrl+Q,=)")
	case input.CmdGoToStyleTag:
		e.cmdNotImplemented("Go to Style Tag (Ctrl+Q,<)")
	case input.CmdGoToNote:
		e.cmdNotImplemented("Go to Note (Ctrl+Q,N,G)")
	case input.CmdGoPrevPosition:
		e.cmdNotImplemented("Go to Previous Position (Ctrl+Q,P)")
	case input.CmdGoLastFindReplace:
		e.cmdNotImplemented("Go to Last Find/Replace (Ctrl+Q,V)")
	case input.CmdGoBlockBegin:
		e.cmdNotImplemented("Go to Beginning of Block (Ctrl+Q,B)")
	case input.CmdGoBlockEnd:
		e.cmdNotImplemented("Go to End of Block (Ctrl+Q,K)")
	case input.CmdGoDocBegin:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyHome})
		e.status.SetText("Ctrl+Q,R: document start")
	case input.CmdGoDocEnd:
		e.entry.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEnd})
		e.status.SetText("Ctrl+Q,C: document end")
	case input.CmdScrollContUp:
		e.cmdNotImplemented("Scroll Continuously Up (Ctrl+Q,W)")
	case input.CmdScrollContDown:
		e.cmdNotImplemented("Scroll Continuously Down (Ctrl+Q,Z)")

	// ── Note ─────────────────────────────────────────────────────────────────
	case input.CmdEditNote:
		e.cmdNotImplemented("Edit Note (Ctrl+O,N,D)")
	case input.CmdConvertNote:
		e.cmdNotImplemented("Convert Note (Ctrl+O,N,V)")

	// ── Settings ──────────────────────────────────────────────────────────────
	case input.CmdAutoAlign:
		e.cmdNotImplemented("Auto Align (Ctrl+O,A)")
	case input.CmdCloseDialog:
		e.status.SetText("Ctrl+O,Enter: close dialog")

	// ── File commands ─────────────────────────────────────────────────────────
	case input.CmdSave:
		e.saveWithPrompt(func(err error) {
			if err != nil {
				if errors.Is(err, errSaveCanceled) {
					return
				}
				e.status.SetText("Error saving: " + err.Error())
				return
			}
			e.status.SetText("File saved")
		})
	case input.CmdSaveAs:
		e.saveAsDialog(func(err error) {
			if err != nil {
				if errors.Is(err, errSaveCanceled) {
					return
				}
				e.status.SetText("Error saving as: " + err.Error())
				return
			}
			e.status.SetText("File saved with new name")
		})
	case input.CmdSaveAndClose:
		e.saveWithPrompt(func(err error) {
			if err != nil {
				if errors.Is(err, errSaveCanceled) {
					return
				}
				e.status.SetText("Error saving: " + err.Error())
				return
			}
			e.closeToBrowser()
		})
	case input.CmdOpenSwitch:
		e.cmdOpenSwitch()
	case input.CmdClose:
		e.cmdClose()
	case input.CmdPrint:
		e.cmdPrint()
	case input.CmdChangePrinter:
		e.cmdChangePrinter()
	case input.CmdFileCopy:
		e.cmdFileCopy()
	case input.CmdFileDelete:
		e.cmdFileDelete()
	case input.CmdFileRename:
		e.cmdFileRename()
	case input.CmdChangeDirectory:
		e.cmdChangeDirectory()
	case input.CmdRunPSCommand:
		e.cmdRunPSCommand()
	case input.CmdStatus:
		e.cmdStatus()
	case input.CmdExit:
		e.cmdExitMSXide()
	}
}

// ── Menus ─────────────────────────────────────────────────────────────────────

func (e *editorUI) makeOpeningFileMenu() *fyne.Menu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New                       S", func() { e.newFile() }),
		fyne.NewMenuItem("Open Document...          D", func() { e.cmdOpenDocument() }),
		fyne.NewMenuItem("Open Nondocument...       N", func() { e.cmdOpenNondocument() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Print...                  P", func() { e.cmdNotImplemented("Print") }),
		fyne.NewMenuItem("Print from keyboard...    K", func() { e.cmdNotImplemented("Print from keyboard") }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Copy...                   O", func() { e.cmdFileCopy() }),
		fyne.NewMenuItem("Delete...                 Y", func() { e.cmdFileDelete() }),
		fyne.NewMenuItem("Rename...                 E", func() { e.cmdFileRename() }),
		fyne.NewMenuItem("Protect/Unprotect...      C", func() { e.cmdProtectUnprotect() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Change Drive...           L", func() { e.cmdChangeDrive() }),
		fyne.NewMenuItem("Change Filename Display", func() { e.cmdChangeFilenameDisplay() }),
		fyne.NewMenuItem("Run CMD Command...        R", func() { e.cmdRunCMDCommand() }),
		fyne.NewMenuItem("Status...                 ?", func() { e.cmdStatus() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit MSXStar              X", func() { e.cmdExitMSXStar() }),
	)
	return fileMenu
}

func (e *editorUI) makeEditorFileMenu() *fyne.Menu {
	return fyne.NewMenu("File",
		fyne.NewMenuItem("New                      Ctrl+N", func() { e.newFile() }),
		fyne.NewMenuItem("Open/Switch              Ctrl+O,K", func() { e.cmdOpenSwitch() }),
		fyne.NewMenuItem("Close                    Ctrl+W", func() { e.cmdClose() }),
		fyne.NewMenuItem("Save                     Ctrl+K,S", func() {
			e.saveWithPrompt(func(err error) {
				if err != nil {
					if errors.Is(err, errSaveCanceled) {
						return
					}
					e.status.SetText("Error saving: " + err.Error())
					return
				}
				e.status.SetText("File saved")
			})
		}),
		fyne.NewMenuItem("Save As...               Ctrl+K,T", func() {
			e.saveAsDialog(func(err error) {
				if err != nil {
					if errors.Is(err, errSaveCanceled) {
						return
					}
					e.status.SetText("Error saving as: " + err.Error())
					return
				}
				e.status.SetText("File saved with new name")
			})
		}),
		fyne.NewMenuItem("Save and Close           Ctrl+K,D", func() {
			e.saveWithPrompt(func(err error) {
				if err != nil {
					if errors.Is(err, errSaveCanceled) {
						return
					}
					e.status.SetText("Error saving: " + err.Error())
					return
				}
				e.closeToBrowser()
			})
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Print...                 Ctrl+K,S Ctrl+K,P", func() { e.cmdPrint() }),
		fyne.NewMenuItem("Change Printer...        Ctrl+P,?", func() { e.cmdChangePrinter() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Copy...                  Ctrl+K,O", func() { e.cmdFileCopy() }),
		fyne.NewMenuItem("Delete...                Ctrl+K,J", func() { e.cmdFileDelete() }),
		fyne.NewMenuItem("Rename...                Ctrl+K,E", func() { e.cmdFileRename() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Change Drive/Directory... Ctrl+K,L", func() { e.cmdChangeDirectory() }),
		fyne.NewMenuItem("Run PS Command...        Ctrl+K,F", func() { e.cmdRunPSCommand() }),
		fyne.NewMenuItem("Status                   Ctrl+O,?", func() { e.cmdStatus() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit MSXide              Ctrl+K,Q,X", func() { e.cmdExitMSXide() }),
	)
}

func (e *editorUI) makeOpeningMenu() *fyne.MainMenu {
	fileMenu := e.makeOpeningFileMenu()
	utilitiesMenu := fyne.NewMenu("Utilities",
		fyne.NewMenuItem("(none)", nil),
	)

	additionalMenu := fyne.NewMenu("Additional",
		fyne.NewMenuItem("(none)", nil),
	)

	return fyne.NewMainMenu(fileMenu, utilitiesMenu, additionalMenu)
}

func (e *editorUI) makeEditorMenu() *fyne.MainMenu {
	fileMenu := e.makeEditorFileMenu()
	editMenu := e.makeEditorEditMenu()
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("(none)", nil),
	)
	insertMenu := fyne.NewMenu("Insert",
		fyne.NewMenuItem("(none)", nil),
	)
	utilitiesMenu := fyne.NewMenu("Utilities",
		fyne.NewMenuItem("(none)", nil),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, viewMenu, insertMenu, utilitiesMenu)
}

func (e *editorUI) makeEditorEditMenu() *fyne.Menu {
	// ── Move submenu ──────────────────────────────────────────────────────────
	moveItem := fyne.NewMenuItem("Move", nil)
	moveItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Block                         Ctrl+K,V", func() { e.execute(input.CmdMoveBlock) }),
		fyne.NewMenuItem("Block from Other Window [NI]  Ctrl+K,G", func() { e.execute(input.CmdMoveBlockOtherWin) }),
	)

	// ── Copy submenu ──────────────────────────────────────────────────────────
	copyItem := fyne.NewMenuItem("Copy", nil)
	copyItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Block                         Ctrl+K,C", func() { e.execute(input.CmdCopyBlock) }),
		fyne.NewMenuItem("Block from Other Window [NI]  Ctrl+K,A", func() { e.execute(input.CmdCopyBlockOtherWin) }),
		fyne.NewMenuItem("From Windows Clipboard        Ctrl+K,[", func() { e.execute(input.CmdCopyFromClipboard) }),
		fyne.NewMenuItem("To Windows Clipboard          Ctrl+K,]", func() { e.execute(input.CmdCopyToClipboard) }),
		fyne.NewMenuItem("To Another File               Ctrl+K,W", func() { e.execute(input.CmdCopyToFile) }),
	)

	// ── Delete submenu ────────────────────────────────────────────────────────
	deleteItem := fyne.NewMenuItem("Delete", nil)
	deleteItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Block                         Ctrl+K,Y", func() { e.execute(input.CmdDeleteBlock) }),
		fyne.NewMenuItem("Word                          Ctrl+T", func() { e.execute(input.CmdDeleteWord) }),
		fyne.NewMenuItem("Line                          Ctrl+Y", func() { e.execute(input.CmdDeleteLine) }),
		fyne.NewMenuItem("Line Left of Cursor           Ctrl+Q,[DEL]", func() { e.execute(input.CmdDeleteLineLeft) }),
		fyne.NewMenuItem("Line Right of Cursor          Ctrl+Q,Y", func() { e.execute(input.CmdDeleteLineRight) }),
		fyne.NewMenuItem("To Character                  Ctrl+T", func() { e.execute(input.CmdDeleteWord) }),
	)

	// ── Go to Marker submenu ──────────────────────────────────────────────────
	goToMarkerItem := fyne.NewMenuItem("Go to Marker", nil)
	var goMarkerItems []*fyne.MenuItem
	for _, d := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		digit := d
		goMarkerItems = append(goMarkerItems, fyne.NewMenuItem(digit, func() {
			e.execute(input.MarkerGoCmd(digit))
		}))
	}
	goToMarkerItem.ChildMenu = fyne.NewMenu("", goMarkerItems...)

	// ── Go to Other submenu ───────────────────────────────────────────────────
	goToOtherItem := fyne.NewMenuItem("Go to Other", nil)
	goToOtherItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Font Tag                      Ctrl+Q,=", func() { e.execute(input.CmdGoToFontTag) }),
		fyne.NewMenuItem("Style Tag                     Ctrl+Q,<", func() { e.execute(input.CmdGoToStyleTag) }),
		fyne.NewMenuItem("Note                          Ctrl+Q,N,G", func() { e.execute(input.CmdGoToNote) }),
		fyne.NewMenuItem("Previous Position             Ctrl+Q,P", func() { e.execute(input.CmdGoPrevPosition) }),
		fyne.NewMenuItem("Last Find/Replace             Ctrl+Q,V", func() { e.execute(input.CmdGoLastFindReplace) }),
		fyne.NewMenuItem("Beginning of Block            Ctrl+Q,B", func() { e.execute(input.CmdGoBlockBegin) }),
		fyne.NewMenuItem("End of Block                  Ctrl+Q,K", func() { e.execute(input.CmdGoBlockEnd) }),
		fyne.NewMenuItem("Document Beginning            Ctrl+Q,R", func() { e.execute(input.CmdGoDocBegin) }),
		fyne.NewMenuItem("Document End                  Ctrl+Q,C", func() { e.execute(input.CmdGoDocEnd) }),
		fyne.NewMenuItem("Scroll Continuously Up        Ctrl+Q,W", func() { e.execute(input.CmdScrollContUp) }),
		fyne.NewMenuItem("Scroll Continuously Down      Ctrl+Q,Z", func() { e.execute(input.CmdScrollContDown) }),
	)

	// ── Set Marker submenu ────────────────────────────────────────────────────
	setMarkerItem := fyne.NewMenuItem("Set Marker", nil)
	var setMarkerItems []*fyne.MenuItem
	for _, d := range []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"} {
		digit := d
		setMarkerItems = append(setMarkerItems, fyne.NewMenuItem(digit, func() {
			e.execute(input.MarkerSetCmd(digit))
		}))
	}
	setMarkerItem.ChildMenu = fyne.NewMenu("", setMarkerItems...)

	// ── Note Options submenu ──────────────────────────────────────────────────
	noteOptionsItem := fyne.NewMenuItem("Note Options", nil)
	noteOptionsItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Starting Number for Note...", func() { e.cmdNotImplemented("Starting Number for Note") }),
		fyne.NewMenuItem("Convert Note...               Ctrl+O,N,V", func() { e.execute(input.CmdConvertNote) }),
		fyne.NewMenuItem("Convert at Print...           .cv", func() { e.cmdNotImplemented("Convert at Print (.cv)") }),
		fyne.NewMenuItem("Endnote Location              .pe", func() { e.cmdNotImplemented("Endnote Location (.pe)") }),
	)

	// ── Edit Settings submenu ─────────────────────────────────────────────────
	editSettingsItem := fyne.NewMenuItem("Edit Settings", nil)
	editSettingsItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Column Block Mode             Ctrl+K,N", func() { e.execute(input.CmdColumnBlockMode) }),
		fyne.NewMenuItem("Column Replace Mode           Ctrl+K,I", func() { e.execute(input.CmdColumnReplaceMode) }),
		fyne.NewMenuItem("Auto Align                    Ctrl+O,A", func() { e.execute(input.CmdAutoAlign) }),
		fyne.NewMenuItem("Closes Dialog                 Ctrl+O,[ENTER]", func() { e.execute(input.CmdCloseDialog) }),
	)

	return fyne.NewMenu("Edit",
		fyne.NewMenuItem("Undo                          Ctrl+U", func() { e.execute(input.CmdUndo) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Mark Block Beginning          Ctrl+K,B", func() { e.execute(input.CmdMarkBlockBegin) }),
		fyne.NewMenuItem("Mark Block End                Ctrl+K,K", func() { e.execute(input.CmdMarkBlockEnd) }),
		moveItem,
		copyItem,
		deleteItem,
		fyne.NewMenuItem("Mark Previous Block           Ctrl+K,U", func() { e.execute(input.CmdMarkPreviousBlock) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Find...                       Ctrl+Q,F", func() { e.cmdFind() }),
		fyne.NewMenuItem("Find and Replace...           Ctrl+Q,A", func() { e.cmdFindReplace() }),
		fyne.NewMenuItem("Next Find                     Ctrl+L", func() { e.cmdRepeatFind() }),
		fyne.NewMenuItem("Go to Character...            Ctrl+Q,G", func() { e.cmdGoToChar() }),
		fyne.NewMenuItem("Go to Page...                 Ctrl+Q,I", func() { e.cmdGoToPage() }),
		goToMarkerItem,
		goToOtherItem,
		setMarkerItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Edit Note                     Ctrl+O,N,D", func() { e.execute(input.CmdEditNote) }),
		noteOptionsItem,
		fyne.NewMenuItemSeparator(),
		editSettingsItem,
	)
}

// ── Menu command handlers ─────────────────────────────────────────────────────

func (e *editorUI) cmdOpenDocument() {
	if e.inEditor {
		e.openFileDialogStd()
		return
	}
	// Browser is already visible — just ensure the list has focus
	e.window.Canvas().Focus(e.browser.list)
}

func (e *editorUI) cmdOpenSwitch() {
	e.withDiscardConfirmation("Open another file", "The current file has unsaved changes. Discard and open another file?", func() {
		e.openFileDialogStd()
	})
}

func (e *editorUI) cmdClose() {
	e.closeToBrowser()
}

func (e *editorUI) cmdOpenNondocument() {
	// Treated identically to Open Document for MSX-BASIC (no rich formatting)
	e.cmdOpenDocument()
}

func (e *editorUI) cmdNotImplemented(name string) {
	dialog.ShowInformation(name, name+" will be implemented in a future update.", e.window)
}

func (e *editorUI) cmdFileCopy() {
	var sourcePath string
	var content string

	if e.inEditor {
		if e.filePath == "" && strings.TrimSpace(e.entry.Text) == "" {
			dialog.ShowInformation("Copy", "No content or current file to copy.", e.window)
			return
		}
		sourcePath = e.filePath
		content = e.entry.Text
	} else {
		entry, ok := e.selectedBrowserFile()
		if !ok {
			dialog.ShowInformation("Copy", "Select a file first.", e.window)
			return
		}
		data, err := os.ReadFile(entry.fullPath)
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		sourcePath = entry.fullPath
		content = string(data)
	}

	e.copyAsDialog(sourcePath, content)
}

func (e *editorUI) cmdFileRename() {
	if e.inEditor {
		if e.filePath == "" {
			dialog.ShowInformation("Rename", "Save the file before renaming.", e.window)
			return
		}
		e.promptRename(e.filePath, func(newPath string) {
			if err := os.Rename(e.filePath, newPath); err != nil {
				dialog.ShowError(err, e.window)
				return
			}
			e.filePath = newPath
			e.updateTitle()
			e.browser.loadDir(filepath.Dir(newPath))
			_ = e.store.SetSetting(context.Background(), "last_file", newPath)
			_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(newPath))
			e.status.SetText("File renamed")
		})
		return
	}

	entry, ok := e.selectedBrowserFile()
	if !ok {
		dialog.ShowInformation("Rename", "Select a file first.", e.window)
		return
	}
	e.promptRename(entry.fullPath, func(newPath string) {
		if err := os.Rename(entry.fullPath, newPath); err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		e.browser.Refresh()
		e.status.SetText("File renamed")
	})
}

func (e *editorUI) cmdProtectUnprotect() {
	dialog.ShowInformation("Protect/Unprotect", "Protect/Unprotect will be implemented in the next update.", e.window)
}

func (e *editorUI) cmdChangeDrive() {
	e.cmdChangeDirectory()
}

func (e *editorUI) cmdChangeFilenameDisplay() {
	dialog.ShowInformation("Change Filename Display", "Change Filename Display will be implemented in the next update.", e.window)
}

func (e *editorUI) cmdRunCMDCommand() {
	e.cmdRunPSCommand()
}

func (e *editorUI) cmdStatus() {
	mode := "Opening Menu"
	if e.inEditor {
		mode = "Editor"
	}
	name := "[New]"
	if e.filePath != "" {
		name = e.filePath
	}
	msg := fmt.Sprintf("Mode: %s\nFile: %s\nModified: %t", mode, name, e.dirty)
	dialog.ShowInformation("Status", msg, e.window)
}

func (e *editorUI) cmdExitMSXStar() {
	e.cmdExitMSXide()
}

func (e *editorUI) cmdFileDelete() {
	if e.inEditor {
		if e.filePath == "" {
			dialog.ShowInformation("Delete File", "Save the file before deleting it from disk.", e.window)
			return
		}
		path := e.filePath
		dialog.ShowConfirm(
			"Delete file",
			fmt.Sprintf("Do you really want to permanently delete:\n%s", path),
			func(ok bool) {
				if !ok {
					return
				}
				if err := os.Remove(path); err != nil {
					dialog.ShowError(err, e.window)
					return
				}
				e.closeToBrowser()
				e.browser.Refresh()
			},
			e.window,
		)
		return
	}
	// Browser is active: offer to delete the selected file
	entry, ok := e.selectedBrowserFile()
	if !ok {
		dialog.ShowInformation("Delete File", "No file selected.", e.window)
		return
	}
	dialog.ShowConfirm(
		"Delete file",
		fmt.Sprintf("Do you really want to permanently delete:\n%s", entry.fullPath),
		func(ok bool) {
			if !ok {
				return
			}
			if err := os.Remove(entry.fullPath); err != nil {
				dialog.ShowError(err, e.window)
				return
			}
			e.browser.Refresh()
		},
		e.window,
	)
}

// openFileDialogStd shows the OS file-open dialog (used from editor mode).
func (e *editorUI) openFileDialogStd() {
	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			e.status.SetText("Error opening: " + err.Error())
			return
		}
		if reader == nil {
			return
		}
		defer func() { _ = reader.Close() }()
		e.openInEditor(reader.URI().Path())
	}, e.window)

	lastDir, _ := e.store.GetSetting(context.Background(), "last_dir")
	if lastDir != "" {
		u, err := storage.ParseURI("file://" + filepath.ToSlash(lastDir))
		if err == nil {
			if lister, lErr := storage.ListerForURI(u); lErr == nil {
				d.SetLocation(lister)
			}
		}
	}
	d.Show()
}

// ── File I/O ──────────────────────────────────────────────────────────────────

func (e *editorUI) saveWithPrompt(onDone func(error)) {
	if e.filePath == "" {
		e.saveAsDialog(onDone)
		return
	}
	if err := os.WriteFile(e.filePath, []byte(e.entry.Text), 0o644); err != nil {
		onDone(err)
		return
	}
	e.dirty = false
	e.updateTitle()
	_ = e.store.TouchRecentFile(context.Background(), e.filePath)
	_ = e.store.SetSetting(context.Background(), "last_file", e.filePath)
	_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(e.filePath))
	onDone(nil)
}

func (e *editorUI) saveAsDialog(onDone func(error)) {
	d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			onDone(err)
			return
		}
		if writer == nil {
			onDone(errSaveCanceled)
			return
		}
		e.filePath = writer.URI().Path()
		_, wErr := writer.Write([]byte(e.entry.Text))
		cErr := writer.Close()
		if wErr != nil {
			onDone(wErr)
			return
		}
		if cErr != nil {
			onDone(cErr)
			return
		}
		e.dirty = false
		e.updateTitle()
		_ = e.store.TouchRecentFile(context.Background(), e.filePath)
		_ = e.store.SetSetting(context.Background(), "last_file", e.filePath)
		_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(e.filePath))
		e.browser.loadDir(filepath.Dir(e.filePath))
		onDone(nil)
	}, e.window)

	lastDir, _ := e.store.GetSetting(context.Background(), "last_dir")
	if lastDir != "" {
		u, err := storage.ParseURI("file://" + filepath.ToSlash(lastDir))
		if err == nil {
			if lister, lErr := storage.ListerForURI(u); lErr == nil {
				d.SetLocation(lister)
			}
		}
	}
	d.Show()
}

func (e *editorUI) copyAsDialog(sourcePath, content string) {
	d := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if writer == nil {
			return
		}
		_, wErr := writer.Write([]byte(content))
		cErr := writer.Close()
		if wErr != nil {
			dialog.ShowError(wErr, e.window)
			return
		}
		if cErr != nil {
			dialog.ShowError(cErr, e.window)
			return
		}
		e.status.SetText("Copy created: " + writer.URI().Path())
	}, e.window)

	startDir := filepath.Dir(sourcePath)
	if startDir == "." || startDir == "" {
		if lastDir, _ := e.store.GetSetting(context.Background(), "last_dir"); lastDir != "" {
			startDir = lastDir
		}
	}
	if startDir != "" {
		u, err := storage.ParseURI("file://" + filepath.ToSlash(startDir))
		if err == nil {
			if lister, lErr := storage.ListerForURI(u); lErr == nil {
				d.SetLocation(lister)
			}
		}
	}
	if sourcePath != "" {
		d.SetFileName(filepath.Base(sourcePath))
	}
	d.Show()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// ── Line number sync ──────────────────────────────────────────────────────────

// syncLineNumbers keeps gutter rows aligned with the Entry viewport.
func (e *editorUI) syncLineNumbers() {
	lineCount := strings.Count(e.entry.Text, "\n") + 1
	if lineCount < 1 {
		lineCount = 1
	}
	visLines := e.visibleLineCount()
	if visLines < 1 {
		visLines = 1
	}

	// Keep cursor visible inside viewport bounds, but preserve real topLine
	// that came from the Entry internal scroll offset.
	topLine := e.topLine
	if e.cursorRow < topLine {
		topLine = e.cursorRow
	}
	if e.cursorRow >= topLine+visLines {
		topLine = e.cursorRow - visLines + 1
	}

	maxTop := lineCount - visLines
	if maxTop < 0 {
		maxTop = 0
	}
	if topLine < 0 {
		topLine = 0
	}
	if topLine > maxTop {
		topLine = maxTop
	}

	e.topLine = topLine
	e.lineNums.Update(lineCount, topLine, e.cursorRow)
}

func (e *editorUI) lineHeightPx() float32 {
	charH := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true}).Height
	if charH < 1 {
		charH = 16
	}
	return charH + 2
}

func (e *editorUI) visibleLineCount() int {
	entryH := e.entry.Size().Height
	if entryH < 1 {
		entryH = 600 // fallback before first layout
	}
	lineH := e.lineHeightPx()
	if lineH <= 0 {
		lineH = 18
	}
	vis := int(entryH/lineH) + 1
	if vis < 1 {
		vis = 1
	}
	return vis
}

func (e *editorUI) applyViewportOffset(offsetY float32) {
	if offsetY < 0 {
		offsetY = 0
	}
	lineH := e.lineHeightPx()
	if lineH <= 0 {
		return
	}
	e.topLine = int(offsetY / lineH)
	if e.activeTab != nil {
		e.activeTab.topLine = e.topLine
	}
	e.syncLineNumbers()
}

func (e *editorUI) updateTitle() {
	name := "[New]"
	if e.filePath != "" {
		name = filepath.Base(e.filePath)
	}
	dirty := ""
	if e.dirty {
		dirty = "*"
	}
	if e.activeTab != nil {
		e.activeTab.filePath = e.filePath
		e.activeTab.dirty = e.dirty
		e.refreshTabTitle(e.activeTab)
	}
	e.window.SetTitle(fmt.Sprintf("WS7 Editor - %s%s", name, dirty))
}

func (e *editorUI) updateCursorStatus() {
	if e.inEditor {
		if e.activeTab != nil {
			e.activeTab.cursorRow = e.cursorRow
			e.activeTab.cursorCol = e.cursorCol
		}
		e.status.SetText(fmt.Sprintf("Ln: %-4d  Col: %-4d", e.cursorRow+1, e.cursorCol+1))
	}
}

func (e *editorUI) applyCursorPosition(row, col int) {
	e.entry.CursorRow = row
	e.entry.CursorColumn = col
	e.entry.Refresh()
	e.cursorRow = row
	e.cursorCol = col
	e.ruler.UpdateCursor(row, col)
	e.syncLineNumbers()
}

func (e *editorUI) closeToBrowser() {
	if e.inEditor {
		e.closeActiveTab()
		return
	}
	e.showBrowser()
}

func (e *editorUI) withDiscardConfirmation(title, message string, next func()) {
	if !e.inEditor || !e.dirty {
		next()
		return
	}
	dialog.ShowConfirm(title, message, func(ok bool) {
		if ok {
			next()
		}
	}, e.window)
}

func (e *editorUI) selectedBrowserFile() (fileEntry, bool) {
	idx := e.browser.selectedIdx
	if idx < 0 || idx >= len(e.browser.entries) {
		return fileEntry{}, false
	}
	entry := e.browser.entries[idx]
	if entry.isDir {
		return fileEntry{}, false
	}
	return entry, true
}

func (e *editorUI) promptRename(oldPath string, onRename func(string)) {
	entry := widget.NewEntry()
	entry.SetText(filepath.Base(oldPath))
	dialog.ShowForm("Rename", "Rename", "Cancel", []*widget.FormItem{
		widget.NewFormItem("New name", entry),
	}, func(ok bool) {
		if !ok {
			return
		}
		newName := strings.TrimSpace(entry.Text)
		if newName == "" {
			dialog.ShowInformation("Rename", "Enter a valid name.", e.window)
			return
		}
		newPath := newName
		if !filepath.IsAbs(newPath) {
			newPath = filepath.Join(filepath.Dir(oldPath), newName)
		}
		if filepath.Clean(newPath) == filepath.Clean(oldPath) {
			return
		}
		onRename(newPath)
	}, e.window)
}

func (e *editorUI) cmdPrint() {
	dialog.ShowInformation("Print", "Printing will be implemented in a future update.", e.window)

}

func (e *editorUI) cmdChangePrinter() {
	dialog.ShowInformation("Change Printer", "Printer selection will be implemented in a future update.", e.window)
}

func (e *editorUI) cmdChangeDirectory() {
	entry := widget.NewEntry()
	startDir := e.browser.dir
	if startDir == "" {
		if e.filePath != "" {
			startDir = filepath.Dir(e.filePath)
		} else {
			cwd, _ := os.Getwd()
			startDir = cwd
		}
	}
	entry.SetText(startDir)
	dialog.ShowForm("Change Drive/Directory", "Open", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Directory", entry),
	}, func(ok bool) {
		if !ok {
			return
		}
		target := strings.TrimSpace(entry.Text)
		if target == "" {
			dialog.ShowInformation("Change Drive/Directory", "Enter a valid directory.", e.window)
			return
		}
		abs, err := filepath.Abs(target)
		if err != nil || !dirExists(abs) {
			dialog.ShowInformation("Change Drive/Directory", "Invalid directory.", e.window)
			return
		}
		e.browser.loadDir(abs)
		e.showBrowser()
	}, e.window)
}

func (e *editorUI) cmdRunPSCommand() {
	entry := widget.NewEntry()
	dialog.ShowForm("Run PS Command", "Run", "Cancel", []*widget.FormItem{
		widget.NewFormItem("PowerShell", entry),
	}, func(ok bool) {
		if !ok {
			return
		}
		command := strings.TrimSpace(entry.Text)
		if command == "" {
			return
		}
		cmd := exec.Command("powershell", "-NoProfile", "-Command", command)
		if e.browser.dir != "" {
			cmd.Dir = e.browser.dir
		}
		output, err := cmd.CombinedOutput()
		result := widget.NewMultiLineEntry()
		result.SetMinRowsVisible(20)
		text := strings.TrimSpace(string(output))
		if text == "" {
			text = "(no output)"
		}
		if err != nil {
			text = text + "\n\nError: " + err.Error()
		}
		result.SetText(text)
		result.Disable()
		dialog.ShowCustom("PS Output", "Close", result, e.window)
	}, e.window)
}

func (e *editorUI) cmdExitMSXide() {
	e.withDiscardConfirmation("Exit MSXide", "The current file has unsaved changes. Do you really want to exit?", func() {
		e.window.Close()
	})
}

func (e *editorUI) cmdMarkBlockBegin() {
	if e.activeTab == nil {
		return
	}
	off := e.cursorByteOffset()
	e.activeTab.blockBegin = off
	e.activeTab.hasBlockBegin = true
	e.updateBlockIndicator()
	e.status.SetText(fmt.Sprintf("Ctrl+K,B: block begin marked at %d", off))
}

func (e *editorUI) cmdMarkBlockEnd() {
	if e.activeTab == nil {
		return
	}
	off := e.cursorByteOffset()
	e.activeTab.blockEnd = off
	e.activeTab.hasBlockEnd = true
	e.updateBlockIndicator()
	start, end, ok := e.activeBlockRange()
	if !ok {
		e.status.SetText(fmt.Sprintf("Ctrl+K,K: block end marked at %d", off))
		return
	}
	e.status.SetText(fmt.Sprintf("Ctrl+K,K: block marked (%d chars)", end-start))
}

func (e *editorUI) cmdCopyBlockMarked() {
	start, end, ok := e.activeBlockRange()
	if !ok {
		if e.activeTab != nil && e.activeTab.hasBlockBegin && e.activeTab.hasBlockEnd {
			e.status.SetText("Ctrl+K,C: empty block (B and K at same position)")
			return
		}
		e.status.SetText("Ctrl+K,C: block is not fully marked")
		return
	}
	e.internalBlockClipboard = e.entry.Text[start:end]
	e.updateInternalClipboardIndicator()
	e.status.SetText(fmt.Sprintf("Ctrl+K,C: copied %d chars to internal clipboard", len(e.internalBlockClipboard)))
}

func (e *editorUI) cmdDeleteBlockMarked() {
	start, end, ok := e.activeBlockRange()
	if !ok {
		if e.activeTab != nil && e.activeTab.hasBlockBegin && e.activeTab.hasBlockEnd {
			e.status.SetText("Ctrl+K,Y: empty block (B and K at same position)")
			return
		}
		e.status.SetText("Ctrl+K,Y: block is not fully marked")
		return
	}
	newText := deleteTextRange(e.entry.Text, start, end)
	e.entry.SetText(newText)
	row, col := offsetToRowCol(newText, start)
	e.applyCursorPosition(row, col)
	e.clearActiveBlockMarks()
	e.status.SetText(fmt.Sprintf("Ctrl+K,Y: deleted block (%d chars)", end-start))
}

func (e *editorUI) cmdMoveBlockMarked() {
	start, end, ok := e.activeBlockRange()
	if !ok {
		if e.activeTab != nil && e.activeTab.hasBlockBegin && e.activeTab.hasBlockEnd {
			e.status.SetText("Ctrl+K,V: empty block (B and K at same position)")
			return
		}
		if e.internalBlockClipboard == "" {
			e.status.SetText("Ctrl+K,V: block is not fully marked")
			return
		}
		insertAt := e.cursorByteOffset()
		text := e.entry.Text
		insertAt = clampOffset(insertAt, len(text))
		newText := text[:insertAt] + e.internalBlockClipboard + text[insertAt:]
		e.entry.SetText(newText)
		row, col := offsetToRowCol(newText, insertAt+len(e.internalBlockClipboard))
		e.applyCursorPosition(row, col)
		e.status.SetText(fmt.Sprintf("Ctrl+K,V: pasted %d chars from internal clipboard", len(e.internalBlockClipboard)))
		return
	}
	dest := e.cursorByteOffset()
	newText, newCursor := moveTextRange(e.entry.Text, start, end, dest)
	e.internalBlockClipboard = e.entry.Text[start:end]
	e.updateInternalClipboardIndicator()
	e.entry.SetText(newText)
	row, col := offsetToRowCol(newText, newCursor)
	e.applyCursorPosition(row, col)
	e.clearActiveBlockMarks()
	e.status.SetText(fmt.Sprintf("Ctrl+K,V: moved %d chars", end-start))
}

func (e *editorUI) clearActiveBlockMarks() {
	if e.activeTab == nil {
		return
	}
	e.activeTab.hasBlockBegin = false
	e.activeTab.hasBlockEnd = false
	e.activeTab.blockBegin = 0
	e.activeTab.blockEnd = 0
	e.updateBlockIndicator()
}

func blockIndicatorForMarks(hasBegin, hasEnd bool) string {
	switch {
	case hasBegin && hasEnd:
		return "[WS7-BLOCK:B,K] "
	case hasBegin:
		return "[WS7-BLOCK:B] "
	case hasEnd:
		return "[WS7-BLOCK:K] "
	default:
		return ""
	}
}

func internalClipboardIndicator(text string) string {
	if text == "" {
		return ""
	}
	return fmt.Sprintf("[WS7-CLIP:%d]", len(text))
}

func (e *editorUI) updateBlockIndicator() {
	if e.activeTab == nil || e.activeTab.blockTag == nil {
		return
	}
	e.activeTab.blockTag.SetText(blockIndicatorForMarks(e.activeTab.hasBlockBegin, e.activeTab.hasBlockEnd))
}

func (e *editorUI) updateInternalClipboardIndicator() {
	if e.activeTab == nil || e.activeTab.clipTag == nil {
		return
	}
	e.activeTab.clipTag.SetText(internalClipboardIndicator(e.internalBlockClipboard))
}

func (e *editorUI) activeBlockRange() (start, end int, ok bool) {
	if e.activeTab == nil || !e.activeTab.hasBlockBegin || !e.activeTab.hasBlockEnd {
		return 0, 0, false
	}
	return normalizeBlockRange(e.activeTab.blockBegin, e.activeTab.blockEnd, len(e.entry.Text))
}

func normalizeBlockRange(a, b, textLen int) (start, end int, ok bool) {
	a = clampOffset(a, textLen)
	b = clampOffset(b, textLen)
	if a <= b {
		start, end = a, b
	} else {
		start, end = b, a
	}
	if start == end {
		return 0, 0, false
	}
	return start, end, true
}

func clampOffset(pos, textLen int) int {
	if pos < 0 {
		return 0
	}
	if pos > textLen {
		return textLen
	}
	return pos
}

func deleteTextRange(text string, start, end int) string {
	start = clampOffset(start, len(text))
	end = clampOffset(end, len(text))
	if start >= end {
		return text
	}
	return text[:start] + text[end:]
}

func moveTextRange(text string, start, end, dest int) (string, int) {
	start, end, ok := normalizeBlockRange(start, end, len(text))
	if !ok {
		return text, clampOffset(dest, len(text))
	}

	block := text[start:end]
	if dest >= start && dest <= end {
		dest = start
	}

	without := deleteTextRange(text, start, end)
	if dest > end {
		dest -= (end - start)
	}
	dest = clampOffset(dest, len(without))
	return without[:dest] + block + without[dest:], dest + len(block)
}

func offsetToRowCol(text string, pos int) (row, col int) {
	pos = clampOffset(pos, len(text))
	before := text[:pos]
	row = strings.Count(before, "\n")
	lastNL := strings.LastIndex(before, "\n")
	col = pos - lastNL - 1
	return row, col
}

func (e *editorUI) deleteCurrentLine() {
	all := e.entry.Text
	if all == "" {
		return
	}
	pos := cursorOffset(all, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 {
		return
	}
	start := strings.LastIndex(all[:pos], "\n") + 1
	end := strings.Index(all[pos:], "\n")
	if end == -1 {
		all = all[:start]
	} else {
		all = all[:start] + all[pos+end+1:]
	}
	e.entry.SetText(all)
}

func cursorOffset(text string, row, col int) int {
	if row < 0 || col < 0 {
		return -1
	}
	lines := strings.Split(text, "\n")
	if row >= len(lines) {
		return len(text)
	}
	offset := 0
	for i := 0; i < row; i++ {
		offset += len(lines[i]) + 1
	}
	if col > len(lines[row]) {
		col = len(lines[row])
	}
	return offset + col
}

// ── Find / Replace / Navigate ─────────────────────────────────────────────────

// findState holds persistent dialog fields between searches.
var findState struct {
	term      string
	replace   string
	backward  bool
	wholeWord bool
	matchCase bool
	useRegex  bool
	// searchFrom is the byte offset to start next search from (used for forward repeat).
	searchFrom int
}

// buildFindDialog builds and shows the custom Find dialog with all search options.
// onSearch is called with the option snapshot when the user clicks Search/Find.
func (e *editorUI) buildFindDialog(withReplace bool) {
	termEntry := widget.NewEntry()
	termEntry.SetMinRowsVisible(1)
	termEntry.SetText(findState.term)
	termEntry.Wrapping = fyne.TextWrapOff
	termEntry.MultiLine = false

	replEntry := widget.NewEntry()
	replEntry.SetMinRowsVisible(1)
	replEntry.SetText(findState.replace)
	replEntry.Wrapping = fyne.TextWrapOff
	replEntry.MultiLine = false

	chkBackward := widget.NewCheck("← Backward", func(_ bool) {})
	chkBackward.Checked = findState.backward

	chkWord := widget.NewCheck("Whole Word", func(_ bool) {})
	chkWord.Checked = findState.wholeWord

	chkCase := widget.NewCheck("Match Case", func(_ bool) {})
	chkCase.Checked = findState.matchCase

	chkRegex := widget.NewCheck("Regular Expression", func(_ bool) {})
	chkRegex.Checked = findState.useRegex

	termLabel := widget.NewLabel("Find:")
	termLabel.TextStyle = fyne.TextStyle{Bold: true}
	termRow := container.New(&minWidthLayout{minW: 560}, termEntry)

	checks := container.NewGridWithColumns(2,
		chkBackward, chkWord,
		chkCase, chkRegex,
	)

	var content fyne.CanvasObject
	if withReplace {
		replLabel := widget.NewLabel("Replace with:")
		replLabel.TextStyle = fyne.TextStyle{Bold: true}
		replRow := container.New(&minWidthLayout{minW: 560}, replEntry)
		content = container.NewVBox(
			termLabel, termRow,
			replLabel, replRow,
			widget.NewSeparator(),
			checks,
		)
	} else {
		content = container.NewVBox(
			termLabel, termRow,
			widget.NewSeparator(),
			checks,
		)
	}

	confirmLabel := "Find"
	title := "Find"
	if withReplace {
		confirmLabel = "Replace"
		title = "Find and Replace"
	}

	dialog.ShowCustomConfirm(title, confirmLabel, "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		term := strings.TrimSpace(termEntry.Text)
		if term == "" {
			return
		}
		findState.term = term
		findState.replace = replEntry.Text
		findState.backward = chkBackward.Checked
		findState.wholeWord = chkWord.Checked
		findState.matchCase = chkCase.Checked
		findState.useRegex = chkRegex.Checked

		if withReplace {
			e.doReplace(term, findState.replace)
		} else {
			findState.searchFrom = e.cursorByteOffset()
			e.doFindFrom(term, findState.searchFrom)
		}
	}, e.window)
}

// minWidthLayout forces its single child to fill at least minW pixels wide.
type minWidthLayout struct{ minW float32 }

func (l *minWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(l.minW, 0)
	}
	ms := objects[0].MinSize()
	if ms.Width < l.minW {
		ms.Width = l.minW
	}
	return ms
}

func (l *minWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Move(fyne.NewPos(0, 0))
		o.Resize(size)
	}
}

func (e *editorUI) cmdFind() {
	findState.searchFrom = e.cursorByteOffset()
	e.buildFindDialog(false)
}

func (e *editorUI) cmdFindReplace() {
	e.buildFindDialog(true)
}

func (e *editorUI) cmdRepeatFind() {
	if findState.term == "" {
		e.cmdFind()
		return
	}
	e.doFindFrom(findState.term, findState.searchFrom)
}

// cursorByteOffset returns the byte offset of the current cursor in the text.
func (e *editorUI) cursorByteOffset() int {
	return cursorOffset(e.entry.Text, e.entry.CursorRow, e.entry.CursorColumn)
}

// doFind always starts from cursor (convenience entry-point).
func (e *editorUI) doFind(term string) {
	findState.searchFrom = e.cursorByteOffset()
	e.doFindFrom(term, findState.searchFrom)
}

// doFindFrom searches for term using current findState options, starting at fromOffset.
func (e *editorUI) doFindFrom(term string, _ int) {
	// overload: fromOffset stored in findState.searchFrom by callers
	e.doFindAt(term, findState.searchFrom)
}

// doFindAt performs the actual search with all option flags.
func (e *editorUI) doFindAt(term string, fromOffset int) {
	text := e.entry.Text

	// Build search function
	type match struct{ start, end int }

	findAll := func(src, pat string) []match {
		if len(pat) == 0 {
			return nil
		}
		var results []match
		haystack := src
		needle := pat
		if !findState.matchCase {
			haystack = strings.ToLower(src)
			needle = strings.ToLower(pat)
		}
		if findState.useRegex {
			// Escape if not valid regex — just use as literal on error
			re, err := regexp.Compile(needle)
			if err != nil {
				e.status.SetText("⚠ Invalid regex: " + err.Error())
				return nil
			}
			for _, loc := range re.FindAllStringIndex(haystack, -1) {
				results = append(results, match{loc[0], loc[1]})
			}
		} else {
			offset := 0
			for {
				i := strings.Index(haystack[offset:], needle)
				if i < 0 {
					break
				}
				start := offset + i
				end := start + len(needle)
				// Whole-word check
				if findState.wholeWord {
					before := start > 0 && isWordChar(src[start-1])
					after := end < len(src) && isWordChar(src[end])
					if before || after {
						offset = start + 1
						continue
					}
				}
				results = append(results, match{start, end})
				offset = end
			}
		}
		return results
	}

	all := findAll(text, term)
	if len(all) == 0 {
		e.status.SetText("Not found: " + term)
		return
	}

	// Pick the right match based on direction
	var chosen match
	found := false

	if findState.backward {
		// Last match whose START is strictly before fromOffset
		for i := len(all) - 1; i >= 0; i-- {
			if all[i].start < fromOffset {
				chosen = all[i]
				found = true
				break
			}
		}
		if !found {
			// wrap around: last match
			chosen = all[len(all)-1]
		}
	} else {
		// First match whose START is >= fromOffset (or == fromOffset+1 on repeat)
		// Use > fromOffset-1 so a match AT cursor is found on first search,
		// but advance past it on repeat by storing end after a hit.
		for _, m := range all {
			if m.start >= fromOffset {
				chosen = m
				found = true
				break
			}
		}
		if !found {
			// wrap around: first match
			chosen = all[0]
		}
	}

	// Place cursor at the match start
	before := text[:chosen.start]
	row := strings.Count(before, "\n")
	lastNL := strings.LastIndex(before, "\n")
	col := chosen.start - lastNL - 1
	e.applyCursorPosition(row, col)

	// Advance searchFrom past this hit for next Ctrl+L
	if findState.backward {
		findState.searchFrom = chosen.start
	} else {
		findState.searchFrom = chosen.end
	}

	matchedText := text[chosen.start:chosen.end]
	e.status.SetText(fmt.Sprintf("Found: '%s'  Ln %d  Col %d", matchedText, row+1, col+1))
}

func isWordChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func (e *editorUI) doReplace(term, repl string) {
	text := e.entry.Text

	// Build the correct comparison text
	compare := text
	pattern := term
	if !findState.matchCase {
		compare = strings.ToLower(text)
		pattern = strings.ToLower(term)
	}

	if findState.useRegex {
		re, err := regexp.Compile(pattern)
		if err != nil {
			e.status.SetText("⚠ Invalid regex: " + err.Error())
			return
		}
		newText := re.ReplaceAllString(text, repl)
		count := len(re.FindAllString(text, -1))
		if count == 0 {
			e.status.SetText("Not found: " + term)
			return
		}
		e.entry.SetText(newText)
		e.status.SetText(fmt.Sprintf("Replaced %d occurrence(s)", count))
		return
	}

	count := strings.Count(compare, pattern)
	if count == 0 {
		e.status.SetText("Not found: " + term)
		return
	}

	var newText string
	if findState.matchCase {
		newText = strings.ReplaceAll(text, term, repl)
	} else {
		// Case-insensitive replace preserving original structure
		var sb strings.Builder
		remaining := text
		cmpRemain := compare
		for {
			idx := strings.Index(cmpRemain, pattern)
			if idx < 0 {
				sb.WriteString(remaining)
				break
			}
			sb.WriteString(remaining[:idx])
			sb.WriteString(repl)
			remaining = remaining[idx+len(term):]
			cmpRemain = cmpRemain[idx+len(pattern):]
		}
		newText = sb.String()
	}
	e.entry.SetText(newText)
	e.status.SetText(fmt.Sprintf("Replaced %d occurrence(s)", count))
}

func (e *editorUI) cmdGoToChar() {
	charEntry := widget.NewEntry()
	dialog.ShowForm("Go to Character", "Go", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Character", charEntry),
	}, func(ok bool) {
		if !ok || strings.TrimSpace(charEntry.Text) == "" {
			return
		}
		e.doFind(charEntry.Text)
	}, e.window)
}

func (e *editorUI) cmdGoToPage() {
	pageEntry := widget.NewEntry()
	dialog.ShowForm("Go to Page", "Go", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Page number", pageEntry),
	}, func(ok bool) {
		if !ok {
			return
		}
		e.status.SetText("Go to Page: " + pageEntry.Text + " (next block)")
	}, e.window)
}

// ── Marker management ─────────────────────────────────────────────────────────

var markers [10]int // byte offsets (0 = not set)

func (e *editorUI) cmdSetMarker(digit string) {
	idx := markerIndex(digit)
	pos := cursorOffset(e.entry.Text, e.entry.CursorRow, e.entry.CursorColumn)
	markers[idx] = pos
	e.status.SetText(fmt.Sprintf("Marker %s set (offset %d)", digit, pos))
}

func (e *editorUI) cmdGoToMarker(digit string) {
	idx := markerIndex(digit)
	pos := markers[idx]
	text := e.entry.Text
	if pos < 0 || pos > len(text) {
		e.status.SetText(fmt.Sprintf("Marker %s not set", digit))
		return
	}
	before := text[:pos]
	row := strings.Count(before, "\n")
	lastNL := strings.LastIndex(before, "\n")
	col := pos - lastNL - 1
	e.applyCursorPosition(row, col)
	e.status.SetText(fmt.Sprintf("Marker %s: line %d col %d", digit, row+1, col+1))
}

func markerIndex(digit string) int {
	if digit == "0" {
		return 9
	}
	if len(digit) > 0 && digit[0] >= '1' && digit[0] <= '9' {
		return int(digit[0] - '1')
	}
	return 0
}

// ── Inline text edit helpers ──────────────────────────────────────────────────

func (e *editorUI) cmdDeleteWordRight() {
	text := e.entry.Text
	pos := cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 || pos >= len(text) {
		return
	}
	// Skip non-spaces then spaces
	end := pos
	for end < len(text) && text[end] != ' ' && text[end] != '\n' {
		end++
	}
	for end < len(text) && text[end] == ' ' {
		end++
	}
	e.entry.SetText(text[:pos] + text[end:])
}

func (e *editorUI) cmdDeleteLineRight() {
	text := e.entry.Text
	pos := cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 {
		return
	}
	nlIdx := strings.Index(text[pos:], "\n")
	if nlIdx < 0 {
		e.entry.SetText(text[:pos])
		return
	}
	e.entry.SetText(text[:pos] + text[pos+nlIdx:])
}

func (e *editorUI) cmdDeleteLineLeft() {
	text := e.entry.Text
	pos := cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 {
		return
	}
	lineStart := strings.LastIndex(text[:pos], "\n") + 1
	e.entry.SetText(text[:lineStart] + text[pos:])
}

func (e *editorUI) cmdInsertLineAtCursor() {
	text := e.entry.Text
	pos := cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 {
		pos = len(text)
	}
	e.entry.SetText(text[:pos] + "\n" + text[pos:])
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
