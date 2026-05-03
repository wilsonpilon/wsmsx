package ui

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/basic/calc"
	"ws7/internal/basic/msxtoken"
	"ws7/internal/basic/renum"
	"ws7/internal/config"
	"ws7/internal/input"
	"ws7/internal/store/sqlite"
	"ws7/internal/version"
)

var errSaveCanceled = errors.New("save canceled")

const ctrlKTimeout = 2 * time.Second

const defaultMSXBasicASCIIExt = ".asc"
const settingEditorThemeKey = "editor_theme"
const settingEditorFontFamilyKey = "editor_font_family"
const settingEditorFontWeightKey = "editor_font_weight"
const settingEditorFontSizeKey = "editor_font_size"
const settingEditorFontItalicKey = "editor_font_italic"
const settingEditorSaveTokenizedKey = "editor_save_tokenized"
const settingWS7BaseDirKey = "tool_ws7_base_dir"
const settingOpenMSXExeKey = "tool_openmsx_exe"
const settingOpenMSXMachineKey = "tool_openmsx_machine"
const settingOpenMSXExt1Key = "tool_openmsx_ext_1"
const settingOpenMSXExt2Key = "tool_openmsx_ext_2"
const settingOpenMSXExt3Key = "tool_openmsx_ext_3"
const settingOpenMSXExt4Key = "tool_openmsx_ext_4"
const settingMSXBas2RomExeKey = "tool_msxbas2rom_exe"
const settingBasicDignifiedExeKey = "tool_basic_dignified_exe"
const settingMSXEncodingExeKey = "tool_msx_encoding_exe"
const defaultRenumStartLine = 10
const defaultRenumIncrement = 10
const defaultRenumFromLine = 0

var msxSourceExtensions = []string{defaultMSXBasicASCIIExt, ".amx", ".bas", ".ldr", ".txt"}

var basicLineNumberRE = regexp.MustCompile(`^\s*(\d+)`)

type remoteHelpDoc struct {
	Title string
	URL   string
	Slug  string
}

var openMSXRemoteHelpDocs = []remoteHelpDoc{
	{Title: "Setup Guide", URL: "https://openmsx.org/manual/setup.html", Slug: "setup-guide"},
	{Title: "User's Manual", URL: "https://openmsx.org/manual/user.html", Slug: "user-manual"},
	{Title: "Console Command Reference", URL: "https://openmsx.org/manual/commands.html", Slug: "console-command-reference"},
	{Title: "Disk Manipulator", URL: "https://openmsx.org/manual/diskmanipulator.html", Slug: "disk-manipulator"},
	{Title: "Control openMSX", URL: "https://openmsx.org/manual/openmsx-control.html", Slug: "control-openmsx"},
}

const (
	openMSXHelpRefreshInterval = 24 * time.Hour
	openMSXHelpCacheMaxAge     = 30 * 24 * time.Hour
)

type newFileType struct {
	ID         string
	Label      string
	DefaultExt string
	Enabled    bool
}

var allNewFileTypes = []newFileType{
	{ID: "msx-basic-ascii", Label: "MSX BASIC ASCII (*.asc)", DefaultExt: ".asc", Enabled: true},
	{ID: "msx-basic-amx", Label: "MSX BASIC Tokenized/AMX (*.amx)", DefaultExt: ".amx", Enabled: true},
	{ID: "assembly", Label: "Assembly (*.asm)", DefaultExt: ".asm", Enabled: false},
	{ID: "c", Label: "C (*.c)", DefaultExt: ".c", Enabled: false},
}

// maxUndoLevels is the maximum number of undo states kept per editor tab.
const maxUndoLevels = 200

// undoState captures a snapshot of editor text and cursor position for undo.
type undoState struct {
	text      string
	cursorRow int
	cursorCol int
}

type editorTab struct {
	item          *container.TabItem
	entry         *cursorEntry
	ruler         *rulerWidget
	floatingRuler *floatingRulerWidget // floating measurement ruler
	lineNums      *lineNumbersWidget
	status        *widget.Label
	blockTag      *widget.Label
	clipTag       *widget.Label

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

	// undo history
	undoStack     []undoState
	lastKnownText string
	undoing       bool

	ruleMode      bool
	isBold        bool
	tokenizedSave bool
}

type editorUI struct {
	fyneApp          fyne.App
	window           fyne.Window
	allowWindowClose bool
	confirmDialog    func(title, message string, onResult func(bool), parent fyne.Window)
	closeWindow      func()
	entry            *cursorEntry
	ruler            *rulerWidget
	lineNums         *lineNumbersWidget
	status           *widget.Label
	blockTag         *widget.Label
	clipTag          *widget.Label
	resolver         *input.Resolver
	store            *sqlite.Store
	browser          *fileBrowser
	openMSXBridge    *openMSXBridgeSession

	filePath        string
	dirty           bool
	inEditor        bool
	cursorRow       int
	cursorCol       int
	topLine         int
	prefixTimeoutID uint64
	prefixExpired   uint32

	tabs             *container.DocTabs
	tabState         map[*container.TabItem]*editorTab
	activeTab        *editorTab
	untitledSeed     map[string]int
	editorThemeID    string
	editorFontFamily string
	editorFontWeight string
	editorFontSize   float32
	editorFontItalic bool
	saveTokenized    bool
	styleTokenItem   *fyne.MenuItem

	internalBlockClipboard string
	calculatorLastResult   string
	calculatorLastValue    float64
	calculatorHasLastValue bool
}

func Run() error {
	a := app.NewWithID("ws7.editor")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	resDir := filepath.Join(cwd, "res")
	if th, thErr := newConfiguredEditorTheme(resDir, defaultEditorThemeID, defaultEditorFontFamily, defaultEditorFontWeight, defaultEditorFontSize); thErr == nil {
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
		fyneApp:          a,
		window:           a.NewWindow(version.Full() + " - Editor"),
		resolver:         input.NewResolver(),
		store:            store,
		tabState:         map[*container.TabItem]*editorTab{},
		editorThemeID:    defaultEditorThemeID,
		editorFontFamily: defaultEditorFontFamily,
		editorFontWeight: defaultEditorFontWeight,
		editorFontSize:   defaultEditorFontSize,
	}
	ui.window.SetCloseIntercept(func() {
		if ui.allowWindowClose {
			ui.window.SetCloseIntercept(nil)
			ui.window.Close()
			return
		}
		ui.requestAppExit()
	})

	if savedEditorThemeID, _ := store.GetSetting(context.Background(), settingEditorThemeKey); savedEditorThemeID != "" {
		ui.editorThemeID = normalizeEditorThemeID(savedEditorThemeID)
	}
	if savedFamily, _ := store.GetSetting(context.Background(), settingEditorFontFamilyKey); strings.TrimSpace(savedFamily) != "" {
		ui.editorFontFamily = normalizeEditorFontFamily(savedFamily)
	}
	if savedWeight, _ := store.GetSetting(context.Background(), settingEditorFontWeightKey); strings.TrimSpace(savedWeight) != "" {
		ui.editorFontWeight = normalizeEditorFontWeight(ui.editorFontFamily, savedWeight)
	}
	if savedSize, _ := store.GetSetting(context.Background(), settingEditorFontSizeKey); strings.TrimSpace(savedSize) != "" {
		if parsed, parseErr := strconv.ParseFloat(strings.TrimSpace(savedSize), 32); parseErr == nil {
			ui.editorFontSize = normalizeEditorFontSize(float32(parsed))
		}
	}
	if savedItalic, _ := store.GetSetting(context.Background(), settingEditorFontItalicKey); strings.EqualFold(strings.TrimSpace(savedItalic), "true") {
		ui.editorFontItalic = true
	}
	if savedTokenized, _ := store.GetSetting(context.Background(), settingEditorSaveTokenizedKey); strings.EqualFold(strings.TrimSpace(savedTokenized), "true") {
		ui.saveTokenized = true
	}
	if !editorFontFamilySupportsItalic(ui.editorFontFamily) {
		ui.editorFontItalic = false
	}
	if th, thErr := newConfiguredEditorTheme(resDir, ui.editorThemeID, ui.editorFontFamily, ui.editorFontWeight, ui.editorFontSize); thErr == nil {
		a.Settings().SetTheme(th)
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
		e.bindActiveTab(tab)
	}
}

func (e *editorUI) bindActiveTab(tab *editorTab) {
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
	e.saveTokenized = tab.tokenizedSave
	e.syncTokenizedMenu()
	e.updateBlockIndicator()
	e.updateInternalClipboardIndicator()
	e.updateTitle()
	e.syncLineNumbers()
}

func (e *editorUI) syncTokenizedMenu() {
	if e.styleTokenItem != nil {
		e.styleTokenItem.Checked = e.saveTokenized
	}
	if e.window != nil && e.window.MainMenu() != nil {
		e.window.MainMenu().Refresh()
	}
}

func (e *editorUI) setTokenizedSaveState(enabled bool, persist bool, showStatus bool) {
	e.saveTokenized = enabled
	if e.activeTab != nil {
		e.activeTab.tokenizedSave = enabled
	}
	e.syncTokenizedMenu()
	if persist && e.store != nil {
		_ = e.store.SetSetting(context.Background(), settingEditorSaveTokenizedKey, strconv.FormatBool(enabled))
	}
	if showStatus && e.status != nil {
		if enabled {
			e.status.SetText("Tokenized save: on")
		} else {
			e.status.SetText("Tokenized save: off")
		}
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
	prevOnChanged := tab.entry.OnChanged
	tab.entry.OnChanged = func(text string) {
		if prevOnChanged != nil {
			prevOnChanged(text)
		}
		// Push undo state (old text) before recording the new text.
		if !tab.undoing {
			state := undoState{text: tab.lastKnownText, cursorRow: tab.cursorRow, cursorCol: tab.cursorCol}
			tab.undoStack = append(tab.undoStack, state)
			if len(tab.undoStack) > maxUndoLevels {
				tab.undoStack = tab.undoStack[len(tab.undoStack)-maxUndoLevels:]
			}
		}
		tab.lastKnownText = text
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

		// Keep floating ruler cursor position synchronized with the editor cursor.
		if tab.floatingRuler != nil {
			charPos := absoluteCharPos(tab.entry.Text, row, col)
			tab.floatingRuler.UpdateCursor(charPos)
			tab.floatingRuler.SetText(tab.entry.Text)
		}

		if e.activeTab == tab {
			e.cursorRow = row
			e.cursorCol = col
			e.ruler.UpdateCursor(row, col)
			e.syncLineNumbers()
			e.updateCursorStatus()
		}
	}
	tab.entry.onViewportOffset = func(x, offsetY float32) {
		if e.activeTab == tab && e.inEditor {
			e.applyViewportOffset(offsetY)
		}
	}
	tab.entry.onSecondaryTapped = func(row, col int) {}
	tab.entry.onKeyBeforeInput = func(key *fyne.KeyEvent) bool {
		if key == nil || key.Name != fyne.KeyEscape {
			return false
		}
		if !tab.ruleMode {
			return false
		}
		e.setRuleMode(tab, false)
		if e.activeTab == tab && e.status != nil {
			e.status.SetText("RULE: off")
		}
		return true
	}
	tab.entry.onRuneBeforeInput = func(r rune) bool {
		if !e.inEditor || e.activeTab != tab {
			return false
		}
		if tab.ruleMode && (r == 'b' || r == 'B') {
			if tab.floatingRuler != nil {
				charPos := absoluteCharPos(tab.entry.Text, tab.cursorRow, tab.cursorCol)
				msg := tab.floatingRuler.MarkBlockPoint(charPos)
				if e.status != nil {
					e.status.SetText(msg)
				}
			}
			return true
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
	return e.nextUntitledNameForExt(defaultMSXBasicASCIIExt)
}

func enabledNewFileTypes() []newFileType {
	types := make([]newFileType, 0, len(allNewFileTypes))
	for _, fileType := range allNewFileTypes {
		if fileType.Enabled {
			types = append(types, fileType)
		}
	}
	return types
}

func defaultNewFileType() newFileType {
	if enabled := enabledNewFileTypes(); len(enabled) > 0 {
		return enabled[0]
	}
	return newFileType{ID: "msx-basic-ascii", Label: "MSX BASIC ASCII (*.asc)", DefaultExt: defaultMSXBasicASCIIExt, Enabled: true}
}

func normalizeFileExt(ext string) string {
	ext = strings.TrimSpace(strings.ToLower(ext))
	if ext == "" {
		return defaultMSXBasicASCIIExt
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func (e *editorUI) nextUntitledNameForExt(ext string) string {
	ext = normalizeFileExt(ext)
	if e.untitledSeed == nil {
		e.untitledSeed = map[string]int{}
	}
	e.untitledSeed[ext]++
	if e.untitledSeed[ext] == 1 {
		return "untitled" + ext
	}
	return fmt.Sprintf("untitled-%d%s", e.untitledSeed[ext], ext)
}

func displayDocumentName(filePath, fallback string) string {
	if filePath != "" {
		return filepath.Base(filePath)
	}
	if fallback != "" {
		return fallback
	}
	return "[New]"
}

func msxSourceFileFilter() storage.FileFilter {
	return storage.NewExtensionFileFilter(msxSourceExtensions)
}

func normalizeMSXSourceFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "[New]" {
		return "untitled" + defaultMSXBasicASCIIExt
	}
	if ext := strings.ToLower(filepath.Ext(name)); ext != "" {
		for _, allowed := range msxSourceExtensions {
			if ext == allowed {
				return name
			}
		}
		return name
	}
	return name + defaultMSXBasicASCIIExt
}

func suggestMSXSourceFileName(filePath, fallback string) string {
	return normalizeMSXSourceFileName(displayDocumentName(filePath, fallback))
}

func suggestMSXSaveFileName(filePath, fallback string, tokenized bool) string {
	name := strings.TrimSpace(displayDocumentName(filePath, fallback))
	if name == "" || name == "[New]" {
		if tokenized {
			return "untitled.bas"
		}
		return "untitled" + defaultMSXBasicASCIIExt
	}
	if tokenized {
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".bas" {
			return name
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if base == "" {
			base = "untitled"
		}
		return base + ".bas"
	}
	if ext := strings.ToLower(filepath.Ext(name)); ext != "" {
		return name
	}
	return name + defaultMSXBasicASCIIExt
}

func (e *editorUI) tabEditorContent(tab *editorTab) fyne.CanvasObject {
	if tab == nil {
		return widget.NewLabel("")
	}
	statusBar := container.NewBorder(nil, nil, nil, container.NewHBox(tab.blockTag, tab.clipTag), tab.status)

	top := container.New(&rulerStartAtTextLayout{gutter: tab.lineNums}, tab.ruler)

	mainContent := container.NewBorder(top, statusBar, tab.lineNums, nil, tab.entry)

	// If ruleMode is active, stack the floating ruler over the main content
	if tab.ruleMode && tab.floatingRuler != nil {
		return container.NewStack(mainContent, tab.floatingRuler)
	}
	return mainContent
}

// rulerStartAtTextLayout keeps the ruler aligned with the editable text area
// by reserving the same leading width used by the line-number gutter.
type rulerStartAtTextLayout struct{ gutter fyne.CanvasObject }

func (l *rulerStartAtTextLayout) gutterWidth() float32 {
	if l == nil || l.gutter == nil {
		return 0
	}
	return l.gutter.MinSize().Width
}

func (l *rulerStartAtTextLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var child fyne.Size
	for _, obj := range objects {
		if obj == nil || !obj.Visible() {
			continue
		}
		ms := obj.MinSize()
		if ms.Width > child.Width {
			child.Width = ms.Width
		}
		if ms.Height > child.Height {
			child.Height = ms.Height
		}
	}
	child.Width += l.gutterWidth()
	return child
}

func (l *rulerStartAtTextLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := l.gutterWidth()
	width := size.Width - x
	if width < 0 {
		width = 0
	}
	for _, obj := range objects {
		if obj == nil {
			continue
		}
		obj.Move(fyne.NewPos(x, 0))
		obj.Resize(fyne.NewSize(width, size.Height))
	}
}

func absoluteCharPos(text string, row, col int) int {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return col
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	if col > len(lines[row]) {
		col = len(lines[row])
	}
	pos := col
	for i := 0; i < row; i++ {
		pos += len(lines[i]) + 1
	}
	return pos
}

func (e *editorUI) newEditorTab(fileType newFileType) *editorTab {
	name := e.nextUntitledNameForExt(fileType.DefaultExt)

	tab := &editorTab{
		entry:         newCursorEntry(),
		ruler:         newRulerWidget(),
		floatingRuler: newFloatingRulerWidget(),
		lineNums:      newLineNumbersWidget(),
		status:        widget.NewLabel(""),
		blockTag:      widget.NewLabel(""),
		clipTag:       widget.NewLabel(""),
		name:          name,
		tokenizedSave: e.saveTokenized,
	}
	tab.blockTag.TextStyle = fyne.TextStyle{Bold: true}
	tab.clipTag.TextStyle = fyne.TextStyle{Bold: true}
	e.bindTabEntry(tab)
	e.applyEditorStyleToTab(tab)
	tab.item = container.NewTabItem(name, e.tabEditorContent(tab))
	e.tabState[tab.item] = tab
	e.tabs.Append(tab.item)
	e.tabs.Select(tab.item)
	e.bindActiveTab(tab)
	e.refreshTabTitle(tab)
	e.recordProgramSnapshot(tab, nil)
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
	e.window.SetTitle(version.Full() + " - Opening Menu")
	e.window.SetContent(e.browser.Content)
	e.window.Canvas().Focus(e.browser.list)
}

func (e *editorUI) showEditor(path string) {
	e.showEditorForType(path, nil)
}

func (e *editorUI) showEditorForType(path string, initialType *newFileType) {
	e.inEditor = true
	e.resetPrefixState()
	e.window.SetMainMenu(e.makeEditorMenu())
	e.ensureTabs()
	if len(e.tabs.Items) == 0 {
		fileType := defaultNewFileType()
		if initialType != nil {
			fileType = *initialType
		}
		e.newEditorTab(fileType)
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
	decodedText, isTokenized, decodeErr := msxtoken.DecodeProgramText(data)
	if decodeErr != nil {
		dialog.ShowError(fmt.Errorf("failed to decode MSX BASIC file: %w", decodeErr), e.window)
		return
	}
	if e.entry == nil && e.tabs != nil {
		if selected := e.tabs.Selected(); selected != nil {
			e.bindActiveTab(e.tabState[selected])
		}
	}
	if e.entry == nil {
		dialog.ShowError(fmt.Errorf("editor is not ready: no active tab selected"), e.window)
		return
	}
	// Suppress undo tracking while the initial file text is loaded.
	if e.activeTab != nil {
		e.activeTab.undoing = true
	}
	e.entry.SetText(decodedText)
	if e.activeTab != nil {
		e.activeTab.lastKnownText = decodedText
		e.activeTab.undoStack = nil
		e.activeTab.undoing = false
	}
	e.filePath = path
	e.dirty = false
	e.cursorRow = 0
	e.cursorCol = 0
	e.topLine = 0
	e.entry.CursorRow = 0
	e.entry.CursorColumn = 0
	if e.ruler != nil {
		e.ruler.UpdateCursor(0, 0)
	}
	e.syncLineNumbers()
	if e.activeTab != nil {
		e.activeTab.filePath = path
		e.activeTab.tokenizedSave = isTokenized
		e.activeTab.dirty = false
		e.activeTab.cursorRow = 0
		e.activeTab.cursorCol = 0
		e.activeTab.topLine = 0
		e.refreshTabTitle(e.activeTab)
		e.recordProgramSnapshot(e.activeTab, nil)
	}
	e.setTokenizedSaveState(isTokenized, false, false)
	if e.store != nil {
		_ = e.store.TouchRecentFile(context.Background(), path)
		_ = e.store.SetSetting(context.Background(), "last_file", path)
		_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(path))
	}
	if e.browser != nil {
		e.browser.loadDir(filepath.Dir(path)) // keep browser in sync
	}
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
	e.promptNewFileType(func(fileType newFileType) {
		e.newFileWithType(fileType)
	})
}

func (e *editorUI) promptNewFileType(onCreate func(newFileType)) {
	if onCreate == nil {
		return
	}
	options := enabledNewFileTypes()
	if len(options) == 0 {
		onCreate(defaultNewFileType())
		return
	}
	if e.window == nil {
		onCreate(defaultNewFileType())
		return
	}
	labels := make([]string, 0, len(options))
	for _, option := range options {
		labels = append(labels, option.Label)
	}
	selectType := widget.NewSelect(labels, nil)
	selectType.SetSelected(labels[0])
	dialog.ShowForm("New File", "Create", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Type", selectType),
	}, func(ok bool) {
		if !ok {
			return
		}
		idx := selectType.SelectedIndex()
		if idx < 0 || idx >= len(options) {
			onCreate(options[0])
			return
		}
		onCreate(options[idx])
	}, e.window)
}

func (e *editorUI) newFileWithType(fileType newFileType) {
	if !e.inEditor {
		e.showEditorForType("", &fileType)
		return
	}
	e.newEditorTab(fileType)
	e.showEditor("")
}

// ── Ctrl key bindings ─────────────────────────────────────────────────────────

func (e *editorUI) bindCtrlKeys() {
	canvas := e.window.Canvas()

	// Letters
	for _, key := range []fyne.KeyName{
		fyne.KeyA, fyne.KeyB, fyne.KeyC, fyne.KeyD, fyne.KeyE,
		fyne.KeyF, fyne.KeyG, fyne.KeyI, fyne.KeyJ, fyne.KeyK,
		fyne.KeyL, fyne.KeyM, fyne.KeyN, fyne.KeyO, fyne.KeyP, fyne.KeyQ,
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
		fyne.KeyL, fyne.KeyM, fyne.KeyN, fyne.KeyO, fyne.KeyP, fyne.KeyQ,
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
	input.CmdIncludeFile:       "Ctrl+K,R",
	input.CmdConvertUppercase:  "Ctrl+K,\"",
	input.CmdConvertLowercase:  "Ctrl+K,'",
	input.CmdConvertCapitalize: "Ctrl+K,.",
	input.CmdDeleteBlock:       "Ctrl+K,Y",
	input.CmdExit:              "Ctrl+K,Q,X",
	input.CmdOpenSwitch:        "Ctrl+O,K",
	input.CmdRule:              "Ctrl+Q,R",
	input.CmdCalculator:        "Ctrl+Q,M",
	input.CmdStatus:            "Ctrl+O,?",
	input.CmdAutoAlign:         "Ctrl+O,A",
	input.CmdStyleBold:         "Ctrl+P,B",
	input.CmdStyleFont:         "Ctrl+P,=",
	input.CmdChangePrinter:     "Ctrl+P,?",
	input.CmdFind:              "Ctrl+Q,F",
	input.CmdFindReplace:       "Ctrl+Q,A",
	input.CmdBasicRenum:        "Ctrl+Q,E",
	input.CmdGoToChar:          "Ctrl+Q,G",
	input.CmdGoToPage:          "Ctrl+Q,I",
	input.CmdGoDocBegin:        "Ctrl+O,L",
	input.CmdGoDocEnd:          "Ctrl+Q,C",
	input.CmdBasicDelete:       "Ctrl+Q,D",
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

	// ── Edit / delete ────────────────────────────────────────────────────────
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
	case input.CmdBasicDelete:
		e.cmdBasicDelete()
	case input.CmdBasicRenum:
		e.cmdBasicRenum()
	case input.CmdDeleteBlock:
		e.cmdDeleteBlockMarked()
	case input.CmdUndo:
		e.cmdUndo()
	case input.CmdNewTab:
		e.newFile()
		e.status.SetText("Ctrl+N: new file")
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
	case input.CmdIncludeFile:
		e.cmdIncludeFile()
	case input.CmdConvertUppercase:
		e.cmdConvertUppercase()
	case input.CmdConvertLowercase:
		e.cmdConvertLowercase()
	case input.CmdConvertCapitalize:
		e.cmdConvertCapitalize()

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
		e.status.SetText("Ctrl+O,L: document start")
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
	case input.CmdStyleBold:
		e.cmdStyleBold()
	case input.CmdStyleFont:
		e.cmdStyleFont()
	case input.CmdRule:
		e.cmdRule()
	case input.CmdCalculator:
		e.cmdCalculator()
	case input.CmdWordCount:
		e.cmdWordCount()
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
		fyne.NewMenuItem("New...                    S", func() { e.newFile() }),
		fyne.NewMenuItem("Open Document...          D", func() { e.cmdOpenDocument() }),
		fyne.NewMenuItem("Open Nondocument...       N", func() { e.cmdOpenNondocument() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Print... [NI]             P", func() { e.cmdNotImplemented("Print") }),
		fyne.NewMenuItem("Print from keyboard... [NI] K", func() { e.cmdNotImplemented("Print from keyboard") }),
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
		fyne.NewMenuItem("Copy Version+Build", func() { e.cmdCopyVersionBuild() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit MSXStar              X", func() { e.cmdExitMSXStar() }),
	)
	return fileMenu
}

func (e *editorUI) makeEditorFileMenu() *fyne.Menu {
	return fyne.NewMenu("File",
		fyne.NewMenuItem("New...                   Ctrl+N", func() { e.newFile() }),
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
		fyne.NewMenuItem("Copy Version+Build", func() { e.cmdCopyVersionBuild() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit MSXide              Ctrl+K,Q,X", func() { e.cmdExitMSXide() }),
	)
}

func (e *editorUI) makeOpeningMenu() *fyne.MainMenu {
	fileMenu := e.makeOpeningFileMenu()
	macrosItem := fyne.NewMenuItem("Macros", nil)
	macrosItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Play... [NI]               MP", func() { e.cmdNotImplemented("Macro Play") }),
		fyne.NewMenuItem("Record... [NI]             MR", func() { e.cmdNotImplemented("Macro Record") }),
		fyne.NewMenuItem("Edit/Create... [NI]        MD", func() { e.cmdNotImplemented("Macro Edit/Create") }),
		fyne.NewMenuItem("Single Step... [NI]        MS", func() { e.cmdNotImplemented("Macro Single Step") }),
		fyne.NewMenuItem("Copy... [NI]               MO", func() { e.cmdNotImplemented("Macro Copy") }),
		fyne.NewMenuItem("Delete... [NI]             MY", func() { e.cmdNotImplemented("Macro Delete") }),
		fyne.NewMenuItem("Rename... [NI]             ME", func() { e.cmdNotImplemented("Macro Rename") }),
	)

	utilitiesMenu := fyne.NewMenu("Utilities",
		macrosItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Configure...", func() { e.cmdConfigure() }),
	)

	additionalMenu := fyne.NewMenu("Additional",
		fyne.NewMenuItem("Character Editor... [NI]   AC", func() { e.cmdNotImplemented("Character Editor") }),
		fyne.NewMenuItem("Hexa Editor... [NI]        AH", func() { e.cmdNotImplemented("Hexa Editor") }),
		fyne.NewMenuItem("Sprite Editor... [NI]      AS", func() { e.cmdNotImplemented("Sprite Editor") }),
		fyne.NewMenuItem("Graphos... [NI]            AG", func() { e.cmdNotImplemented("Graphos") }),
		fyne.NewMenuItem("Noise Editor... [NI]       AN", func() { e.cmdNotImplemented("Noise Editor") }),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("README                    HR", func() { e.cmdOpenHelpReadme() }),
		fyne.NewMenuItem("MANUAL                    HM", func() { e.cmdOpenHelpManual() }),
		fyne.NewMenuItem("OUTLINE                   HO", func() { e.cmdOpenHelpOutline() }),
	)

	// Keep Help as the last top-level menu so it stays rightmost.
	return fyne.NewMainMenu(fileMenu, utilitiesMenu, additionalMenu, helpMenu)
}

func (e *editorUI) makeEditorMenu() *fyne.MainMenu {
	fileMenu := e.makeEditorFileMenu()
	editMenu := e.makeEditorEditMenu()
	insertMenu := fyne.NewMenu("Insert",
		fyne.NewMenuItem("Include File            Ctrl+K,R", func() { e.execute(input.CmdIncludeFile) }),
		fyne.NewMenuItem("Extended Character      Ctrl+M,G", func() { e.execute(input.CmdInsertExtendedChar) }),
	)
	convertCaseItem := fyne.NewMenuItem("Convert Case", nil)
	convertCaseItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Uppercase                Ctrl+K,\"", func() { e.execute(input.CmdConvertUppercase) }),
		fyne.NewMenuItem("Lowercase                Ctrl+K,'", func() { e.execute(input.CmdConvertLowercase) }),
		fyne.NewMenuItem("Capitalize               Ctrl+K,.", func() { e.execute(input.CmdConvertCapitalize) }),
	)
	tokenizedItem := fyne.NewMenuItem("Tokenized", func() { e.cmdToggleTokenizedSave() })
	tokenizedItem.Checked = e.saveTokenized
	e.styleTokenItem = tokenizedItem
	styleMenu := fyne.NewMenu("Style",
		fyne.NewMenuItem("Bold                     Ctrl+P,B", func() { e.execute(input.CmdStyleBold) }),
		fyne.NewMenuItem("Font...                  Ctrl+P,=", func() { e.execute(input.CmdStyleFont) }),
		tokenizedItem,
		convertCaseItem,
	)
	runOpenMSXItem := fyne.NewMenuItem("Execute on openMSX [NI]", func() { e.cmdNotImplemented("Execute on openMSX") })
	runMakeDiskItem := fyne.NewMenuItem("Make a Disk [NI]", func() { e.cmdNotImplemented("Make a Disk") })
	runBadigItem := fyne.NewMenuItem("Transpile on BADIG [NI]", func() { e.cmdNotImplemented("Transpile on BADIG") })
	runMSXBas2RomItem := fyne.NewMenuItem("Compile on msxbas2rom [NI]", func() { e.cmdNotImplemented("Compile on msxbas2rom") })
	runMenu := fyne.NewMenu("Run",
		runOpenMSXItem,
		runMakeDiskItem,
		runBadigItem,
		runMSXBas2RomItem,
	)
	utilitiesMenu := fyne.NewMenu("Utilities",
		fyne.NewMenuItem("RULE                       Ctrl+Q,R  ESC to exit", func() { e.cmdRule() }),
		fyne.NewMenuItem("Calculator                 Ctrl+Q,M", func() { e.execute(input.CmdCalculator) }),
		fyne.NewMenuItem("Word Count", func() { e.execute(input.CmdWordCount) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open openMSX", func() { e.cmdLaunchOpenMSX() }),
		fyne.NewMenuItem("Run msxbas2rom", func() { e.cmdLaunchMSXBas2Rom() }),
		fyne.NewMenuItem("Run BASIC Dignified", func() { e.cmdLaunchBasicDignified() }),
		fyne.NewMenuItem("Run MSX Encoding", func() { e.cmdLaunchMSXEncoding() }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Configure...", func() { e.cmdConfigure() }),
	)

	openMSXHelpItem := fyne.NewMenuItem("openMSX", nil)
	lastUpdateItem := fyne.NewMenuItem(openMSXHelpLastUpdatedLabel(), func() {})
	openMSXHelpItem.ChildMenu = fyne.NewMenu("",
		lastUpdateItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Setup Guide", func() { e.cmdOpenMSXHelpDoc(openMSXRemoteHelpDocs[0]) }),
		fyne.NewMenuItem("User's Manual", func() { e.cmdOpenMSXHelpDoc(openMSXRemoteHelpDocs[1]) }),
		fyne.NewMenuItem("Console Command Reference", func() { e.cmdOpenMSXHelpDoc(openMSXRemoteHelpDocs[2]) }),
		fyne.NewMenuItem("Disk Manipulator", func() { e.cmdOpenMSXHelpDoc(openMSXRemoteHelpDocs[3]) }),
		fyne.NewMenuItem("Control openMSX", func() { e.cmdOpenMSXHelpDoc(openMSXRemoteHelpDocs[4]) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Update Help", func() { e.cmdOpenMSXHelpUpdate() }),
	)

	helpMenu := fyne.NewMenu("HELP",
		fyne.NewMenuItem("ABOUT", func() { e.cmdAboutWS7() }),
		openMSXHelpItem,
	)

	return fyne.NewMainMenu(fileMenu, editMenu, insertMenu, styleMenu, runMenu, utilitiesMenu, helpMenu)
}

func (e *editorUI) setRuleMode(tab *editorTab, enabled bool) {
	if tab == nil || tab.ruleMode == enabled {
		return
	}
	tab.ruleMode = enabled
	if enabled && tab.floatingRuler != nil {
		origin := absoluteCharPos(tab.entry.Text, tab.cursorRow, tab.cursorCol)
		tab.floatingRuler.SetText(tab.entry.Text)
		tab.floatingRuler.SetOriginCharPos(origin)
		tab.floatingRuler.UpdateCursor(origin)
		tab.floatingRuler.ResetBlockSelection()
	}
	if tab.item != nil {
		tab.item.Content = e.tabEditorContent(tab)
	}
	if e.tabs != nil {
		e.tabs.Refresh()
	}
	if e.activeTab == tab && e.window != nil && e.entry != nil {
		e.window.Canvas().Focus(e.entry)
	}
}

func (e *editorUI) cmdRule() {
	if e.activeTab == nil {
		return
	}
	next := !e.activeTab.ruleMode
	e.setRuleMode(e.activeTab, next)
	if e.status != nil {
		if next {
			e.status.SetText("RULE: on (ESC to exit)")
		} else {
			e.status.SetText("RULE: off")
		}
	}
}

func (e *editorUI) cmdToggleTokenizedSave() {
	e.setTokenizedSaveState(!e.saveTokenized, true, true)
}

func (e *editorUI) cmdCalculator() {
	if e.window == nil {
		return
	}

	exprEntry := widget.NewEntry()
	exprEntry.SetPlaceHolder("Example: sqr(81) + (&H10 XOR 3) + (1 << 4)")
	exprEntry.SetMinRowsVisible(2)

	resultEntry := widget.NewMultiLineEntry()
	resultEntry.Wrapping = fyne.TextWrapWord
	resultEntry.Disable()
	resultEntry.SetMinRowsVisible(3)
	if strings.TrimSpace(e.calculatorLastResult) != "" {
		resultEntry.SetText(e.calculatorLastResult)
	} else {
		resultEntry.SetText("(no calculation yet)")
	}

	help := widget.NewLabel(
		"Supported: +  -  *  /  ^  sqr  int  hex()  bin()  dec()\n" +
			"Bitwise: XOR  AND  OR  NOT  <<  >>  rol(a,n)  ror(a,n)  shl(a,n)  shr(a,n)\n" +
			"Number input: decimal by default, &Hxxxx for hex, &Bxxxx for binary. '.' reuses last result.",
	)
	help.Wrapping = fyne.TextWrapWord

	var dlg *dialog.CustomDialog
	calculate := func() {
		res, err := calc.EvaluateWithLast(exprEntry.Text, e.calculatorLastValue, e.calculatorHasLastValue)
		if err != nil {
			msg := "Error: " + err.Error()
			resultEntry.SetText(msg)
			if e.status != nil {
				e.status.SetText("Calculator: " + err.Error())
			}
			return
		}
		formatted := fmt.Sprintf("Decimal: %s\nHex: %s\nBinary: %s", res.Decimal, res.Hex, res.Binary)
		e.calculatorLastResult = formatted
		e.calculatorLastValue = res.Value
		e.calculatorHasLastValue = true
		resultEntry.SetText(formatted)
		if e.status != nil {
			e.status.SetText("Calculator: calculated")
		}
	}
	exprEntry.OnSubmitted = func(string) {
		calculate()
	}
	okBtn := widget.NewButton("=", func() {
		calculate()
	})
	cancelBtn := widget.NewButton("Cancel", func() {
		if dlg != nil {
			dlg.Hide()
		}
		if e.status != nil {
			e.status.SetText("Calculator: canceled")
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Enter Mathematical Expression to be Calculated:"),
		exprEntry,
		widget.NewSeparator(),
		widget.NewLabel("Result of Last Calculation:"),
		resultEntry,
		widget.NewSeparator(),
		help,
		container.NewHBox(layout.NewSpacer(), okBtn, cancelBtn),
	)

	dlg = dialog.NewCustomWithoutButtons("Calculator", content, e.window)
	dlg.Resize(fyne.NewSize(760, 360))
	dlg.Show()
	if e.status != nil {
		e.status.SetText("Calculator: ready")
	}
}

func (e *editorUI) cmdWordCount() {
	if e.window == nil || e.entry == nil {
		return
	}

	text := e.entry.Text
	wordCount, charCount := countWordsAndChars(text)

	stats := fmt.Sprintf("Words:      %d\nCharacters: %d bytes", wordCount, charCount)
	result := widget.NewMultiLineEntry()
	result.SetText(stats)
	result.Disable()
	result.SetMinRowsVisible(8)

	dialog.ShowCustom("Word Count", "Close", result, e.window)
	if e.status != nil {
		e.status.SetText(fmt.Sprintf("Word Count: %d words, %d bytes", wordCount, charCount))
	}
}

func countWordsAndChars(text string) (int, int) {
	words := 0
	chars := len(text)
	inWord := false

	for _, r := range text {
		isSpace := r == ' ' || r == '\t' || r == '\n' || r == '\r'
		if !isSpace && !inWord {
			words++
			inWord = true
		} else if isSpace && inWord {
			inWord = false
		}
	}

	return words, chars
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
		fyne.NewMenuItem("Font Tag [NI]                 Ctrl+Q,=", func() { e.execute(input.CmdGoToFontTag) }),
		fyne.NewMenuItem("Style Tag [NI]                Ctrl+Q,<", func() { e.execute(input.CmdGoToStyleTag) }),
		fyne.NewMenuItem("Note [NI]                     Ctrl+Q,N,G", func() { e.execute(input.CmdGoToNote) }),
		fyne.NewMenuItem("Previous Position [NI]        Ctrl+Q,P", func() { e.execute(input.CmdGoPrevPosition) }),
		fyne.NewMenuItem("Last Find/Replace [NI]        Ctrl+Q,V", func() { e.execute(input.CmdGoLastFindReplace) }),
		fyne.NewMenuItem("Beginning of Block [NI]       Ctrl+Q,B", func() { e.execute(input.CmdGoBlockBegin) }),
		fyne.NewMenuItem("End of Block [NI]             Ctrl+Q,K", func() { e.execute(input.CmdGoBlockEnd) }),
		fyne.NewMenuItem("Document Beginning            Ctrl+O,L", func() { e.execute(input.CmdGoDocBegin) }),
		fyne.NewMenuItem("Document End                  Ctrl+Q,C", func() { e.execute(input.CmdGoDocEnd) }),
		fyne.NewMenuItem("Scroll Continuously Up [NI]   Ctrl+Q,W", func() { e.execute(input.CmdScrollContUp) }),
		fyne.NewMenuItem("Scroll Continuously Down [NI] Ctrl+Q,Z", func() { e.execute(input.CmdScrollContDown) }),
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
		fyne.NewMenuItem("Starting Number for Note... [NI]", func() { e.cmdNotImplemented("Starting Number for Note") }),
		fyne.NewMenuItem("Convert Note... [NI]          Ctrl+O,N,V", func() { e.execute(input.CmdConvertNote) }),
		fyne.NewMenuItem("Convert at Print... [NI]      .cv", func() { e.cmdNotImplemented("Convert at Print (.cv)") }),
		fyne.NewMenuItem("Endnote Location [NI]         .pe", func() { e.cmdNotImplemented("Endnote Location (.pe)") }),
	)

	// ── Edit Settings submenu ─────────────────────────────────────────────────
	editSettingsItem := fyne.NewMenuItem("Edit Settings", nil)
	editSettingsItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Column Block Mode [NI]        Ctrl+K,N", func() { e.execute(input.CmdColumnBlockMode) }),
		fyne.NewMenuItem("Column Replace Mode [NI]      Ctrl+K,I", func() { e.execute(input.CmdColumnReplaceMode) }),
		fyne.NewMenuItem("Auto Align [NI]               Ctrl+O,A", func() { e.execute(input.CmdAutoAlign) }),
		fyne.NewMenuItem("Closes Dialog                 Ctrl+O,[ENTER]", func() { e.execute(input.CmdCloseDialog) }),
	)

	basicItem := fyne.NewMenuItem("BASIC", nil)
	basicItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("DELETE...                     Ctrl+Q,D", func() { e.execute(input.CmdBasicDelete) }),
		fyne.NewMenuItem("RENUM...                      Ctrl+Q,E", func() { e.execute(input.CmdBasicRenum) }),
	)

	return fyne.NewMenu("Edit",
		basicItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Undo                          Ctrl+U", func() { e.execute(input.CmdUndo) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Mark Block Beginning          Ctrl+K,B", func() { e.execute(input.CmdMarkBlockBegin) }),
		fyne.NewMenuItem("Mark Block End                Ctrl+K,K", func() { e.execute(input.CmdMarkBlockEnd) }),
		moveItem,
		copyItem,
		deleteItem,
		fyne.NewMenuItem("Mark Previous Block [NI]      Ctrl+K,U", func() { e.execute(input.CmdMarkPreviousBlock) }),
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
		fyne.NewMenuItem("Edit Note [NI]                Ctrl+O,N,D", func() { e.execute(input.CmdEditNote) }),
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

func (e *editorUI) cmdIncludeFile() {
	if e.window == nil {
		if e.status != nil {
			e.status.SetText("Include File: unavailable")
		}
		return
	}
	if e.entry == nil {
		if e.status != nil {
			e.status.SetText("Include File: no active editor")
		}
		return
	}

	d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			if e.status != nil {
				e.status.SetText("Include File error: " + err.Error())
			}
			return
		}
		if reader == nil {
			return
		}
		defer func() { _ = reader.Close() }()

		path := reader.URI().Path()
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			if e.status != nil {
				e.status.SetText("Include File error: " + readErr.Error())
			}
			return
		}

		e.insertTextAtCursor(string(data), filepath.Base(path))
	}, e.window)

	lastDir := ""
	if e.store != nil {
		lastDir, _ = e.store.GetSetting(context.Background(), "last_dir")
	}
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

func (e *editorUI) cmdClose() {
	e.closeToBrowser()
}

func (e *editorUI) cmdOpenNondocument() {
	// Treated identically to Open Document for MSX-BASIC (no rich formatting)
	e.cmdOpenDocument()
}

func (e *editorUI) cmdOpenHelpReadme() {
	e.openMarkdownHelpDoc("README.md", "README")
}

func (e *editorUI) cmdOpenHelpManual() {
	e.openMarkdownHelpDoc("MANUAL.md", "MANUAL")
}

func (e *editorUI) cmdOpenHelpOutline() {
	e.openMarkdownHelpDoc("OUTLINE.md", "OUTLINE")
}

func (e *editorUI) applyCurrentEditorTheme() {
	cwd, err := os.Getwd()
	if err == nil {
		resDir := filepath.Join(cwd, "res")
		if th, thErr := newConfiguredEditorTheme(resDir, e.editorThemeID, e.editorFontFamily, e.editorFontWeight, e.editorFontSize); thErr == nil {
			e.fyneApp.Settings().SetTheme(th)
		}
	}

	for _, tab := range e.tabState {
		e.applyEditorStyleToTab(tab)
	}

	if e.inEditor {
		e.window.SetMainMenu(e.makeEditorMenu())
	}
	if e.window.Content() != nil {
		e.window.Content().Refresh()
	}
}

func (e *editorUI) showHTMLHelp(label, sourceURL string, rawHTML []byte) {
	segments := htmlToRichTextSegments(rawHTML, sourceURL)
	rt := widget.NewRichText(segments...)
	rt.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(rt)

	viewer := e.fyneApp.NewWindow(fmt.Sprintf("%s - %s", version.Full(), label))
	viewer.Resize(fyne.NewSize(980, 700))
	viewer.SetContent(container.NewBorder(
		widget.NewLabel(sourceURL),
		nil,
		nil,
		nil,
		scroll,
	))
	viewer.Show()
}

func (e *editorUI) showMarkdownHelp(label, source, markdown string) {
	md := widget.NewRichTextFromMarkdown(markdown)
	md.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(md)

	viewer := e.fyneApp.NewWindow(fmt.Sprintf("%s - %s", version.Full(), label))
	viewer.Resize(fyne.NewSize(920, 680))
	viewer.SetContent(container.NewBorder(
		widget.NewLabel(source),
		nil,
		nil,
		nil,
		scroll,
	))
	viewer.Show()
}

func (e *editorUI) openMarkdownHelpDoc(fileName, label string) {
	path, err := findProjectDocPath(fileName)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		dialog.ShowError(err, e.window)
		return
	}

	e.showMarkdownHelp(label, filepath.Base(path), string(data))
}

func buildInfoSummary() (compiledAt, goVersion, target string) {
	compiledAt = "n/a"
	goVersion = runtime.Version()
	target = runtime.GOOS + "/" + runtime.GOARCH

	if info, ok := debug.ReadBuildInfo(); ok {
		if strings.TrimSpace(info.GoVersion) != "" {
			goVersion = strings.TrimSpace(info.GoVersion)
		}
		goos := ""
		goarch := ""
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.time":
				if ts, err := time.Parse(time.RFC3339, s.Value); err == nil {
					compiledAt = ts.Local().Format("2006-01-02 15:04:05 MST")
				}
			case "GOOS":
				goos = strings.TrimSpace(s.Value)
			case "GOARCH":
				goarch = strings.TrimSpace(s.Value)
			}
		}
		if goos != "" && goarch != "" {
			target = goos + "/" + goarch
		}
	}

	return compiledAt, goVersion, target
}

func (e *editorUI) cmdAboutWS7() {
	compiledAt, goVersion, target := buildInfoSummary()
	about := strings.Join([]string{
		"# WS7",
		"",
		"- Version: `" + version.Version + "`",
		"- Build: `" + version.Build() + "`",
		"- Compiled at: `" + compiledAt + "`",
		"- Go compiler: `" + goVersion + "`",
		"- Target OS/Arch: `" + target + "`",
	}, "\n")
	e.showMarkdownHelp("ABOUT", "WS7 build information", about)
}

func openMSXHelpCacheDir() string {
	if base, err := os.UserCacheDir(); err == nil && strings.TrimSpace(base) != "" {
		return filepath.Join(base, "ws7", "help", "openmsx")
	}
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Join(cwd, "res", "help", "openmsx")
	}
	return filepath.Join(os.TempDir(), "ws7", "help", "openmsx")
}

func openMSXHelpFilePath(doc remoteHelpDoc) string {
	name := strings.TrimSpace(doc.Slug)
	if name == "" {
		name = "doc"
	}
	return filepath.Join(openMSXHelpCacheDir(), name+".html")
}

func openMSXHelpLastUpdated() (time.Time, bool) {
	var latest time.Time
	found := false
	for _, doc := range openMSXRemoteHelpDocs {
		path := openMSXHelpFilePath(doc)
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			continue
		}
		if !found || st.ModTime().After(latest) {
			latest = st.ModTime()
			found = true
		}
	}
	return latest, found
}

func openMSXHelpLastUpdatedLabel() string {
	if ts, ok := openMSXHelpLastUpdated(); ok {
		return "Ultima atualizacao: " + ts.Local().Format("2006-01-02 15:04")
	}
	return "Ultima atualizacao: n/a"
}

func pruneOpenMSXHelpCacheByAge(maxAge time.Duration, now time.Time) (int, error) {
	if maxAge <= 0 {
		return 0, nil
	}
	cacheDir := openMSXHelpCacheDir()
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	removed := 0
	deadline := now.Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.EqualFold(filepath.Ext(entry.Name()), ".html") {
			continue
		}
		full := filepath.Join(cacheDir, entry.Name())
		st, statErr := os.Stat(full)
		if statErr != nil {
			continue
		}
		if st.ModTime().Before(deadline) {
			if removeErr := os.Remove(full); removeErr == nil {
				removed++
			}
		}
	}
	return removed, nil
}

func downloadOpenMSXHelpHTML(doc remoteHelpDoc) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, doc.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WS7-HelpFetcher/1.0")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func ensureOpenMSXHelpFile(doc remoteHelpDoc, forceUpdate bool) (htmlPath string, source string, err error) {
	path := openMSXHelpFilePath(doc)

	cacheIsUsable := false
	cacheIsStale := true
	if st, statErr := os.Stat(path); statErr == nil && !st.IsDir() {
		cacheIsUsable = true
		cacheIsStale = time.Since(st.ModTime()) > openMSXHelpRefreshInterval
	}

	if !forceUpdate && cacheIsUsable && !cacheIsStale {
		return path, "cache", nil
	}

	raw, fetchErr := downloadOpenMSXHelpHTML(doc)
	if fetchErr != nil {
		if cacheIsUsable {
			return path, "cache-offline", nil
		}
		return "", "", fetchErr
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", "", err
	}
	return path, "online", nil
}

func (e *editorUI) cmdOpenMSXHelpDoc(doc remoteHelpDoc) {
	_, _ = pruneOpenMSXHelpCacheByAge(openMSXHelpCacheMaxAge, time.Now())
	htmlPath, source, err := ensureOpenMSXHelpFile(doc, false)
	if err != nil {
		// No cache and no internet — try the live URL in the system browser.
		if u, pErr := url.Parse(doc.URL); pErr == nil {
			_ = e.fyneApp.OpenURL(u)
		}
		return
	}
	if source == "cache-offline" {
		dialog.ShowInformation("openMSX Help", "Sem internet — exibindo cache local.", e.window)
	}
	if source == "online" && e.window != nil && e.inEditor {
		e.window.SetMainMenu(e.makeEditorMenu())
	}
	rawHTML, readErr := os.ReadFile(htmlPath)
	if readErr != nil {
		dialog.ShowError(readErr, e.window)
		return
	}
	e.showHTMLHelp("openMSX - "+doc.Title, doc.URL, rawHTML)
}

func (e *editorUI) cmdOpenMSXHelpUpdate() {
	_, _ = pruneOpenMSXHelpCacheByAge(openMSXHelpCacheMaxAge, time.Now())
	success := 0
	offlineFallback := 0
	fails := make([]string, 0)
	for _, doc := range openMSXRemoteHelpDocs {
		_, source, dlErr := ensureOpenMSXHelpFile(doc, true)
		if dlErr != nil {
			fails = append(fails, doc.Title+": "+dlErr.Error())
			continue
		}
		if source == "cache-offline" {
			offlineFallback++
		}
		success++
	}

	msg := fmt.Sprintf("Updated %d/%d openMSX help document(s).", success, len(openMSXRemoteHelpDocs))
	if offlineFallback > 0 {
		msg += fmt.Sprintf("\n\nSem internet, mantendo cache para %d documento(s).", offlineFallback)
	}
	if len(fails) > 0 {
		msg += "\n\nFailed:\n- " + strings.Join(fails, "\n- ")
	}
	dialog.ShowInformation("openMSX Help Update", msg, e.window)
	if e.window != nil && e.inEditor {
		e.window.SetMainMenu(e.makeEditorMenu())
	}
}

func findProjectDocPath(fileName string) (string, error) {
	if fileName == "" {
		return "", fmt.Errorf("empty document file name")
	}

	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, fileName)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), fileName)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("help document not found: %s", fileName)
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

func versionBuildTraceText() string {
	return fmt.Sprintf("Version: %s | Build: %s", version.Full(), version.Build())
}

func (e *editorUI) cmdCopyVersionBuild() {
	trace := versionBuildTraceText()
	e.window.Clipboard().SetContent(trace)
	if e.status != nil {
		e.status.SetText("Copied: " + trace)
		return
	}
	dialog.ShowInformation("Copy Version+Build", trace, e.window)
}

func (e *editorUI) cmdStatus() {
	mode := "Opening Menu"
	if e.inEditor {
		mode = "Editor"
	}
	name := displayDocumentName(e.filePath, "")
	if e.filePath != "" {
		name = e.filePath
	} else if e.activeTab != nil {
		name = displayDocumentName("", e.activeTab.name)
	}
	msg := fmt.Sprintf("Version: %s\nBuild: %s\nMode: %s\nFile: %s\nModified: %t", version.Full(), version.Build(), mode, name, e.dirty)
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
	d.SetFilter(msxSourceFileFilter())

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
	payload, err := e.savePayload()
	if err != nil {
		onDone(err)
		return
	}
	if err := os.WriteFile(e.filePath, payload, 0o644); err != nil {
		onDone(err)
		return
	}
	e.dirty = false
	e.updateTitle()
	_ = e.store.TouchRecentFile(context.Background(), e.filePath)
	_ = e.store.SetSetting(context.Background(), "last_file", e.filePath)
	_ = e.store.SetSetting(context.Background(), "last_dir", filepath.Dir(e.filePath))
	if e.activeTab != nil {
		e.recordProgramSnapshot(e.activeTab, nil)
	}
	onDone(nil)
}

func (e *editorUI) savePayload() ([]byte, error) {
	if !e.saveTokenized {
		return []byte(normalizeDOSLineEndings(e.entry.Text)), nil
	}
	return msxtoken.TokenizeProgram(e.entry.Text)
}

func normalizeDOSLineEndings(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.ReplaceAll(normalized, "\n", "\r\n")
}

func writeNormalizedASCII(writer io.Writer, content string) (int, error) {
	return writer.Write([]byte(normalizeDOSLineEndings(content)))
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
		payload, pErr := e.savePayload()
		if pErr != nil {
			_ = writer.Close()
			onDone(pErr)
			return
		}
		_, wErr := writer.Write(payload)
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
		if e.activeTab != nil {
			e.recordProgramSnapshot(e.activeTab, nil)
		}
		onDone(nil)
	}, e.window)
	d.SetFilter(msxSourceFileFilter())

	fallbackName := ""
	if e.activeTab != nil {
		fallbackName = e.activeTab.name
	}
	d.SetFileName(suggestMSXSaveFileName(e.filePath, fallbackName, e.saveTokenized))

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
		_, wErr := writeNormalizedASCII(writer, content)
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
	d.SetFilter(msxSourceFileFilter())

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
	d.SetFileName(suggestMSXSourceFileName(sourcePath, "untitled.asc"))
	d.Show()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// ── Line number sync ──────────────────────────────────────────────────────────

// syncLineNumbers keeps gutter rows aligned with the Entry viewport.
func (e *editorUI) syncLineNumbers() {
	if e.entry == nil || e.lineNums == nil {
		return
	}

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
	if e.activeTab != nil {
		e.activeTab.topLine = topLine
	}
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
	fallbackName := ""
	if e.activeTab != nil {
		fallbackName = e.activeTab.name
	}
	name := displayDocumentName(e.filePath, fallbackName)
	dirty := ""
	if e.dirty {
		dirty = "*"
	}
	if e.activeTab != nil {
		e.activeTab.filePath = e.filePath
		e.activeTab.dirty = e.dirty
		e.refreshTabTitle(e.activeTab)
	}
	if e.window == nil {
		return
	}
	e.window.SetTitle(fmt.Sprintf("%s - %s%s", version.Full(), name, dirty))
}

func (e *editorUI) updateCursorStatus() {
	if e.inEditor {
		if e.activeTab != nil {
			e.activeTab.cursorRow = e.cursorRow
			e.activeTab.cursorCol = e.cursorCol
		}
		if e.status != nil {
			e.status.SetText(fmt.Sprintf("Ln: %-4d  Col: %-4d", e.cursorRow+1, e.cursorCol+1))
		}
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
	e.showConfirm(title, message, func(ok bool) {
		if ok {
			next()
		}
	})
}

func (e *editorUI) showConfirm(title, message string, onResult func(bool)) {
	if e.confirmDialog != nil {
		e.confirmDialog(title, message, onResult, e.window)
		return
	}
	dialog.ShowConfirm(title, message, onResult, e.window)
}

func (e *editorUI) unsavedTabsCount() int {
	count := 0
	for _, tab := range e.tabState {
		if tab != nil && tab.dirty {
			count++
		}
	}
	if count == 0 && e.inEditor && e.dirty {
		return 1
	}
	return count
}

func (e *editorUI) closeWindowNow() {
	if e.window == nil {
		return
	}
	e.closeOpenMSXBridge()
	e.allowWindowClose = true
	if e.closeWindow != nil {
		e.closeWindow()
		return
	}
	e.window.Close()
}

func (e *editorUI) requestAppExit() {
	unsaved := e.unsavedTabsCount()
	if unsaved == 0 {
		e.closeWindowNow()
		return
	}
	message := fmt.Sprintf("There are unsaved changes in %d tab(s). Do you really want to exit?", unsaved)
	e.showConfirm("Exit MSXide", message, func(ok bool) {
		if ok {
			e.closeWindowNow()
		}
	})
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

func configureInitialDirectory(rawPath, fallbackDir string) string {
	rawPath = strings.TrimSpace(rawPath)
	fallbackDir = strings.TrimSpace(fallbackDir)

	if rawPath != "" {
		if dirExists(rawPath) {
			return rawPath
		}
		if dir := filepath.Dir(rawPath); dir != "" && dir != "." && dirExists(dir) {
			return dir
		}
	}

	if fallbackDir != "" && dirExists(fallbackDir) {
		return fallbackDir
	}

	if cwd, err := os.Getwd(); err == nil && dirExists(cwd) {
		return cwd
	}

	return ""
}

var ws7SubdirectoryNames = []string{"TEMP", "DSKA", "DSKB", "RES", "UTIL"}

func normalizeWS7BaseDirectory(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	clean := filepath.Clean(base)
	name := filepath.Base(clean)
	if strings.EqualFold(name, "ws7.exe") {
		return filepath.Dir(clean)
	}
	return clean
}

func validateWS7BaseDirectory(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("directory is empty")
	}

	clean := filepath.Clean(raw)
	if strings.EqualFold(filepath.Base(clean), "ws7.exe") {
		st, err := os.Stat(clean)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("ws7.exe was not found")
			}
			return err
		}
		if st.IsDir() {
			return fmt.Errorf("ws7.exe path points to a directory")
		}
		return nil
	}

	st, err := os.Stat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist")
		}
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}

func buildWS7SubdirectoryPaths(baseDir string) map[string]string {
	base := normalizeWS7BaseDirectory(baseDir)
	paths := make(map[string]string, len(ws7SubdirectoryNames))
	for _, name := range ws7SubdirectoryNames {
		if base == "" {
			paths[name] = ""
			continue
		}
		paths[name] = filepath.Join(base, name)
	}
	return paths
}

func createWS7Subdirectories(baseDir string) error {
	if err := validateWS7BaseDirectory(baseDir); err != nil {
		return err
	}
	base := normalizeWS7BaseDirectory(baseDir)
	for _, name := range ws7SubdirectoryNames {
		if err := os.MkdirAll(filepath.Join(base, name), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func configuredToolCandidatePaths(toolID, dir string) []string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	switch toolID {
	case settingOpenMSXExeKey:
		return []string{
			filepath.Join(dir, "openmsx.exe"),
			filepath.Join(dir, "openmsx"),
		}
	case settingMSXBas2RomExeKey:
		return []string{
			filepath.Join(dir, "msxbas2rom.exe"),
			filepath.Join(dir, "msxbas2rom"),
		}
	case settingBasicDignifiedExeKey:
		return []string{
			filepath.Join(dir, "badig.py"),
			filepath.Join(dir, "badig.exe"),
			filepath.Join(dir, "badig"),
		}
	case settingMSXEncodingExeKey:
		return []string{
			filepath.Join(dir, "dist", "extension.js"),
			filepath.Join(dir, "package.json"),
			filepath.Join(dir, "esbuild.js"),
		}
	default:
		return nil
	}
}

func detectConfiguredToolPath(toolID, dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}
	for _, candidate := range configuredToolCandidatePaths(toolID, dir) {
		if stat, err := os.Stat(candidate); err == nil && !stat.IsDir() {
			return filepath.Clean(candidate)
		}
	}
	return filepath.Clean(dir)
}

func openMSXExtensionSettingKeys() []string {
	return []string{settingOpenMSXExt1Key, settingOpenMSXExt2Key, settingOpenMSXExt3Key, settingOpenMSXExt4Key}
}

func normalizeOpenMSXResourceName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	if strings.EqualFold(filepath.Ext(name), ".xml") {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return strings.TrimSpace(name)
}

func detectOpenMSXResourceDir(configuredPath, subdir string) string {
	configuredPath = strings.TrimSpace(configuredPath)
	subdir = strings.TrimSpace(subdir)
	if configuredPath == "" || subdir == "" {
		return ""
	}

	rootCandidates := make([]string, 0, 4)
	clean := filepath.Clean(configuredPath)
	if st, err := os.Stat(clean); err == nil {
		if st.IsDir() {
			rootCandidates = append(rootCandidates, clean)
		} else {
			rootCandidates = append(rootCandidates, filepath.Dir(clean))
		}
	} else {
		rootCandidates = append(rootCandidates, filepath.Dir(clean))
	}

	seen := map[string]struct{}{}
	for _, root := range rootCandidates {
		for _, candidate := range []string{filepath.Join(root, "share", subdir), filepath.Join(filepath.Dir(root), "share", subdir)} {
			key := strings.ToLower(filepath.Clean(candidate))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			if st, err := os.Stat(candidate); err == nil && st.IsDir() {
				return filepath.Clean(candidate)
			}
		}
	}

	return ""
}

func listOpenMSXXMLResourceNames(configuredPath, subdir string) []string {
	resourceDir := detectOpenMSXResourceDir(configuredPath, subdir)
	if resourceDir == "" {
		return nil
	}

	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		return nil
	}

	names := make([]string, 0, len(entries))
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(filepath.Ext(name), ".xml") {
			continue
		}
		base := normalizeOpenMSXResourceName(name)
		if base == "" {
			continue
		}
		key := strings.ToLower(base)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, base)
	}
	sort.Strings(names)
	return names
}

func buildOpenMSXResourceOptions(detected []string, current string) []string {
	options := []string{""}
	seen := map[string]struct{}{"": {}}

	for _, name := range detected {
		normalized := normalizeOpenMSXResourceName(name)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		options = append(options, normalized)
	}

	if normalized := normalizeOpenMSXResourceName(current); normalized != "" {
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; !ok {
			options = append(options, normalized)
		}
	}

	return options
}

func updateOpenMSXResourceSelect(selectWidget *widget.Select, detected []string, selected string) {
	if selectWidget == nil {
		return
	}
	selectWidget.Options = buildOpenMSXResourceOptions(detected, selected)
	selectWidget.Refresh()
	selectWidget.SetSelected(normalizeOpenMSXResourceName(selected))
}

func configuredToolLabel(toolID string) string {
	switch toolID {
	case settingOpenMSXExeKey:
		return "openMSX"
	case settingMSXBas2RomExeKey:
		return "msxbas2rom"
	case settingBasicDignifiedExeKey:
		return "BASIC Dignified"
	case settingMSXEncodingExeKey:
		return "MSX Encoding"
	default:
		return "External Tool"
	}
}

func resolveConfiguredToolPath(toolID, configuredValue string) (string, error) {
	raw := strings.TrimSpace(configuredValue)
	if raw == "" {
		return "", fmt.Errorf("path is not configured")
	}

	clean := filepath.Clean(raw)
	if st, err := os.Stat(clean); err == nil {
		if !st.IsDir() {
			return clean, nil
		}
		detected := detectConfiguredToolPath(toolID, clean)
		if detected == "" || filepath.Clean(detected) == clean {
			return "", fmt.Errorf("no executable/script detected in directory: %s", clean)
		}
		if dst, derr := os.Stat(detected); derr == nil && !dst.IsDir() {
			return filepath.Clean(detected), nil
		}
		return "", fmt.Errorf("detected path is not a file: %s", detected)
	}

	parent := filepath.Dir(clean)
	if parent != "" && parent != "." {
		if st, err := os.Stat(parent); err == nil && st.IsDir() {
			detected := detectConfiguredToolPath(toolID, parent)
			if detected != "" && filepath.Clean(detected) != filepath.Clean(parent) {
				if dst, derr := os.Stat(detected); derr == nil && !dst.IsDir() {
					return filepath.Clean(detected), nil
				}
			}
		}
	}

	return "", fmt.Errorf("configured path does not exist: %s", clean)
}

func buildConfiguredToolCommand(toolID, resolvedPath string, extraArgs []string) (name string, args []string, workDir string) {
	resolvedPath = filepath.Clean(strings.TrimSpace(resolvedPath))
	ext := strings.ToLower(filepath.Ext(resolvedPath))
	base := strings.ToLower(filepath.Base(resolvedPath))

	if ext == ".py" {
		return "python", append([]string{"-u", resolvedPath}, extraArgs...), filepath.Dir(resolvedPath)
	}
	if ext == ".js" {
		return "node", append([]string{resolvedPath}, extraArgs...), filepath.Dir(resolvedPath)
	}
	if toolID == settingMSXEncodingExeKey && base == "package.json" {
		dir := filepath.Dir(resolvedPath)
		return "npm", []string{"--prefix", dir, "run", "compile"}, dir
	}

	return resolvedPath, extraArgs, filepath.Dir(resolvedPath)
}

type toolProbeSpec struct {
	name    string
	args    []string
	workDir string
	label   string
}

func buildConfiguredToolProbeSpecs(toolID, resolvedPath string) []toolProbeSpec {
	resolvedPath = filepath.Clean(strings.TrimSpace(resolvedPath))
	if resolvedPath == "" {
		return nil
	}

	dir := filepath.Dir(resolvedPath)
	ext := strings.ToLower(filepath.Ext(resolvedPath))
	base := strings.ToLower(filepath.Base(resolvedPath))

	specs := make([]toolProbeSpec, 0, 3)

	if ext == ".py" {
		specs = append(specs,
			toolProbeSpec{name: "python", args: []string{"-u", resolvedPath, "--help"}, workDir: dir, label: "python -u <script> --help"},
			toolProbeSpec{name: "python", args: []string{"-u", resolvedPath, "-h"}, workDir: dir, label: "python -u <script> -h"},
		)
		return specs
	}

	if ext == ".js" {
		specs = append(specs, toolProbeSpec{name: "node", args: []string{"--check", resolvedPath}, workDir: dir, label: "node --check <script>"})
		return specs
	}

	if toolID == settingMSXEncodingExeKey && base == "package.json" {
		specs = append(specs, toolProbeSpec{name: "npm", args: []string{"--prefix", dir, "--version"}, workDir: dir, label: "npm --prefix <dir> --version"})
		return specs
	}

	switch toolID {
	case settingOpenMSXExeKey:
		specs = append(specs,
			toolProbeSpec{name: resolvedPath, args: []string{"--version"}, workDir: dir, label: "<tool> --version"},
			toolProbeSpec{name: resolvedPath, args: []string{"-v"}, workDir: dir, label: "<tool> -v"},
		)
	case settingMSXBas2RomExeKey, settingBasicDignifiedExeKey:
		specs = append(specs,
			toolProbeSpec{name: resolvedPath, args: []string{"--help"}, workDir: dir, label: "<tool> --help"},
			toolProbeSpec{name: resolvedPath, args: []string{"-h"}, workDir: dir, label: "<tool> -h"},
		)
	default:
		specs = append(specs, toolProbeSpec{name: resolvedPath, args: []string{"--version"}, workDir: dir, label: "<tool> --version"})
	}

	return specs
}

func runToolProbeWithTimeout(spec toolProbeSpec, timeout time.Duration) (string, error) {
	if strings.TrimSpace(spec.name) == "" {
		return "", fmt.Errorf("empty command name")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, spec.name, spec.args...)
	if strings.TrimSpace(spec.workDir) != "" {
		cmd.Dir = spec.workDir
	}
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if ctx.Err() == context.DeadlineExceeded {
		return text, fmt.Errorf("timeout after %s", timeout)
	}
	return text, err
}

func (e *editorUI) cmdConfigure() {
	themeOptions := []struct {
		Label string
		ID    string
	}{
		{"Dark", editorThemeDarkID},
		{"Light", editorThemeLightID},
		{"One Dark", editorThemeOneDarkID},
		{"Monokai", editorThemeMonokaiID},
		{"Solarized Dark", editorThemeSolarizedID},
		{"Github Dark", editorThemeGithubID},
	}
	labels := make([]string, len(themeOptions))
	currentTheme := normalizeEditorThemeID(e.editorThemeID)
	initialLabel := "Dark"
	for i, opt := range themeOptions {
		labels[i] = opt.Label
		if opt.ID == currentTheme {
			initialLabel = opt.Label
		}
	}

	themeSelect := widget.NewSelect(labels, nil)
	themeSelect.SetSelected(initialLabel)

	loadSetting := func(key string) string {
		if e.store == nil {
			return ""
		}
		v, _ := e.store.GetSetting(context.Background(), key)
		return strings.TrimSpace(v)
	}

	lastDir := loadSetting("last_dir")

	testToolPath := func(title, toolID string, entry *widget.Entry) {
		raw := strings.TrimSpace(entry.Text)
		if raw == "" {
			dialog.ShowInformation(title, "Path is empty. Choose a folder or type a tool path first.", e.window)
			if e.status != nil {
				e.status.SetText(title + ": empty path")
			}
			return
		}

		resolved, err := resolveConfiguredToolPath(toolID, raw)
		if err != nil {
			dialog.ShowInformation(title, "Validation failed: "+err.Error(), e.window)
			if e.status != nil {
				e.status.SetText(title + ": invalid path")
			}
			return
		}

		probes := buildConfiguredToolProbeSpecs(toolID, resolved)
		if len(probes) == 0 {
			dialog.ShowInformation(title, "No lightweight probe available for this tool path.", e.window)
			if e.status != nil {
				e.status.SetText(title + ": no probe available")
			}
			return
		}

		var (
			successSpec *toolProbeSpec
			successOut  string
			lastErrSpec toolProbeSpec
			lastErrOut  string
			lastErr     error
		)

		for i := range probes {
			spec := probes[i]
			out, probeErr := runToolProbeWithTimeout(spec, 8*time.Second)
			if probeErr == nil {
				successSpec = &spec
				successOut = out
				break
			}
			lastErrSpec = spec
			lastErrOut = out
			lastErr = probeErr
		}

		if successSpec != nil {
			msg := "Resolved file: " + resolved + "\n" +
				"Probe: " + successSpec.name + " " + strings.Join(successSpec.args, " ") + "\n" +
				"Work dir: " + successSpec.workDir + "\n" +
				"Result: success"
			if strings.TrimSpace(successOut) != "" {
				msg += "\n\nOutput:\n" + successOut
			}
			dialog.ShowInformation(title, msg, e.window)
			if e.status != nil {
				e.status.SetText(title + ": test ok")
			}
			return
		}

		msg := "Resolved file: " + resolved + "\n" +
			"Probe: " + lastErrSpec.name + " " + strings.Join(lastErrSpec.args, " ") + "\n" +
			"Work dir: " + lastErrSpec.workDir + "\n" +
			"Result: failed - " + lastErr.Error()
		if strings.TrimSpace(lastErrOut) != "" {
			msg += "\n\nOutput:\n" + lastErrOut
		}
		dialog.ShowInformation(title, msg, e.window)
		if e.status != nil {
			e.status.SetText(title + ": test failed")
		}
	}

	bindDirectoryPicker := func(title, toolID string, entry *widget.Entry) *fyne.Container {
		test := widget.NewButton("Test", func() {
			testToolPath(title, toolID, entry)
		})
		browse := widget.NewButton("Browse...", func() {
			d := dialog.NewFolderOpen(func(listable fyne.ListableURI, err error) {
				if err != nil {
					if e.status != nil {
						e.status.SetText(title + ": " + err.Error())
					}
					return
				}
				if listable == nil {
					return
				}
				selectedDir := filepath.Clean(listable.Path())
				resolved := detectConfiguredToolPath(toolID, selectedDir)
				entry.SetText(resolved)
				if e.status != nil {
					if resolved != selectedDir {
						e.status.SetText(title + ": detected " + filepath.Base(resolved))
					} else {
						e.status.SetText(title + ": folder selected (no executable auto-detected)")
					}
				}
			}, e.window)

			initialDir := configureInitialDirectory(entry.Text, lastDir)
			if initialDir != "" {
				u, err := storage.ParseURI("file://" + filepath.ToSlash(initialDir))
				if err == nil {
					if lister, lErr := storage.ListerForURI(u); lErr == nil {
						d.SetLocation(lister)
					}
				}
			}
			d.Show()
		})
		actions := container.NewHBox(test, browse)
		return container.NewBorder(nil, nil, nil, actions, entry)
	}

	openMSXExe := widget.NewEntry()
	openMSXExe.SetPlaceHolder("Browse a folder or type the full openMSX path")
	openMSXExe.SetText(loadSetting(settingOpenMSXExeKey))

	configuredMachine := loadSetting(settingOpenMSXMachineKey)
	configuredExts := make([]string, len(openMSXExtensionSettingKeys()))
	for i, key := range openMSXExtensionSettingKeys() {
		configuredExts[i] = loadSetting(key)
	}

	machineSelect := widget.NewSelect(buildOpenMSXResourceOptions(listOpenMSXXMLResourceNames(openMSXExe.Text, "machines"), configuredMachine), nil)
	machineSelect.SetSelected(normalizeOpenMSXResourceName(configuredMachine))

	openMSXExtSelects := []*widget.Select{
		widget.NewSelect(buildOpenMSXResourceOptions(listOpenMSXXMLResourceNames(openMSXExe.Text, "extensions"), configuredExts[0]), nil),
		widget.NewSelect(buildOpenMSXResourceOptions(listOpenMSXXMLResourceNames(openMSXExe.Text, "extensions"), configuredExts[1]), nil),
		widget.NewSelect(buildOpenMSXResourceOptions(listOpenMSXXMLResourceNames(openMSXExe.Text, "extensions"), configuredExts[2]), nil),
		widget.NewSelect(buildOpenMSXResourceOptions(listOpenMSXXMLResourceNames(openMSXExe.Text, "extensions"), configuredExts[3]), nil),
	}
	for i := range openMSXExtSelects {
		openMSXExtSelects[i].SetSelected(normalizeOpenMSXResourceName(configuredExts[i]))
	}

	machineHint := widget.NewLabel("")
	machineHint.Wrapping = fyne.TextWrapWord
	extHint := widget.NewLabel("")
	extHint.Wrapping = fyne.TextWrapWord
	refreshOpenMSXHints := func() {
		machines := listOpenMSXXMLResourceNames(openMSXExe.Text, "machines")
		extensions := listOpenMSXXMLResourceNames(openMSXExe.Text, "extensions")

		updateOpenMSXResourceSelect(machineSelect, machines, machineSelect.Selected)
		for i := range openMSXExtSelects {
			updateOpenMSXResourceSelect(openMSXExtSelects[i], extensions, openMSXExtSelects[i].Selected)
		}

		if len(machines) == 0 {
			machineHint.SetText("Detected machines: none (check openMSX path/share/machines)")
		} else {
			sample := machines
			if len(sample) > 8 {
				sample = sample[:8]
			}
			machineHint.SetText(fmt.Sprintf("Detected machines (%d): %s", len(machines), strings.Join(sample, ", ")))
		}

		if len(extensions) == 0 {
			extHint.SetText("Detected extensions: none (check openMSX path/share/extensions)")
		} else {
			sample := extensions
			if len(sample) > 8 {
				sample = sample[:8]
			}
			extHint.SetText(fmt.Sprintf("Detected extensions (%d): %s", len(extensions), strings.Join(sample, ", ")))
		}
	}
	openMSXExe.OnChanged = func(string) {
		refreshOpenMSXHints()
	}
	refreshHintsBtn := widget.NewButton("Refresh openMSX Lists", func() {
		refreshOpenMSXHints()
		if e.status != nil {
			e.status.SetText("Configure: openMSX lists refreshed")
		}
	})
	refreshOpenMSXHints()

	ws7BaseDir := widget.NewEntry()
	ws7BaseDir.SetPlaceHolder("Directory that contains ws7.exe")
	ws7BaseDir.SetText(loadSetting(settingWS7BaseDirKey))
	ws7ValidationHint := widget.NewLabel("")
	ws7ValidationHint.Wrapping = fyne.TextWrapWord

	ws7SubdirEntries := map[string]*widget.Entry{}
	for _, name := range ws7SubdirectoryNames {
		entry := widget.NewEntry()
		entry.Disable()
		ws7SubdirEntries[name] = entry
	}
	refreshWS7SubdirFields := func() {
		paths := buildWS7SubdirectoryPaths(ws7BaseDir.Text)
		for _, name := range ws7SubdirectoryNames {
			ws7SubdirEntries[name].SetText(paths[name])
		}
	}
	var ws7CreateBtn *widget.Button
	applyWS7BaseValidation := func() {
		err := validateWS7BaseDirectory(ws7BaseDir.Text)
		ws7BaseDir.SetValidationError(err)
		if err != nil {
			ws7ValidationHint.SetText("WS7 directory invalid: " + err.Error())
			if ws7CreateBtn != nil {
				ws7CreateBtn.Disable()
			}
		} else {
			ws7ValidationHint.SetText("WS7 directory looks valid.")
			if ws7CreateBtn != nil {
				ws7CreateBtn.Enable()
			}
		}
		refreshWS7SubdirFields()
	}
	ws7BaseDir.OnChanged = func(string) {
		applyWS7BaseValidation()
	}

	ws7BrowseBtn := widget.NewButton("Browse...", func() {
		d := dialog.NewFolderOpen(func(listable fyne.ListableURI, err error) {
			if err != nil {
				if e.status != nil {
					e.status.SetText("ws7 directory: " + err.Error())
				}
				return
			}
			if listable == nil {
				return
			}
			ws7BaseDir.SetText(filepath.Clean(listable.Path()))
		}, e.window)

		initialDir := configureInitialDirectory(ws7BaseDir.Text, lastDir)
		if initialDir != "" {
			u, err := storage.ParseURI("file://" + filepath.ToSlash(initialDir))
			if err == nil {
				if lister, lErr := storage.ListerForURI(u); lErr == nil {
					d.SetLocation(lister)
				}
			}
		}
		d.Show()
	})
	ws7CreateBtn = widget.NewButton("Create WS7 Subdirs", func() {
		if err := validateWS7BaseDirectory(ws7BaseDir.Text); err != nil {
			dialog.ShowInformation("WS7 Directories", "Invalid WS7 directory: "+err.Error(), e.window)
			if e.status != nil {
				e.status.SetText("ws7 subdirs: invalid directory")
			}
			applyWS7BaseValidation()
			return
		}
		if err := createWS7Subdirectories(ws7BaseDir.Text); err != nil {
			dialog.ShowError(err, e.window)
			if e.status != nil {
				e.status.SetText("ws7 subdirs: failed")
			}
			return
		}
		refreshWS7SubdirFields()
		dialog.ShowInformation("WS7 Directories", "Created/validated directories: TEMP, DSKA, DSKB, RES, UTIL.", e.window)
		if e.status != nil {
			e.status.SetText("ws7 subdirs: ready")
		}
	})
	ws7DirField := container.NewBorder(nil, nil, nil, container.NewHBox(ws7BrowseBtn, ws7CreateBtn), ws7BaseDir)
	applyWS7BaseValidation()

	msxbas2romExe := widget.NewEntry()
	msxbas2romExe.SetPlaceHolder("Browse a folder or type the full msxbas2rom path")
	msxbas2romExe.SetText(loadSetting(settingMSXBas2RomExeKey))

	basicDignifiedExe := widget.NewEntry()
	basicDignifiedExe.SetPlaceHolder("Browse a folder or type the full BASIC Dignified path")
	basicDignifiedExe.SetText(loadSetting(settingBasicDignifiedExeKey))

	msxEncodingExe := widget.NewEntry()
	msxEncodingExe.SetPlaceHolder("Browse a folder or type the full MSX Encoding path")
	msxEncodingExe.SetText(loadSetting(settingMSXEncodingExeKey))

	form := widget.NewForm(
		widget.NewFormItem("Editor Theme", themeSelect),
		widget.NewFormItem("WS7 Directory", ws7DirField),
		widget.NewFormItem("", ws7ValidationHint),
		widget.NewFormItem("WS7 TEMP", ws7SubdirEntries["TEMP"]),
		widget.NewFormItem("WS7 DSKA", ws7SubdirEntries["DSKA"]),
		widget.NewFormItem("WS7 DSKB", ws7SubdirEntries["DSKB"]),
		widget.NewFormItem("WS7 RES", ws7SubdirEntries["RES"]),
		widget.NewFormItem("WS7 UTIL", ws7SubdirEntries["UTIL"]),
		widget.NewFormItem("openMSX Path", bindDirectoryPicker("openMSX path", settingOpenMSXExeKey, openMSXExe)),
		widget.NewFormItem("", refreshHintsBtn),
		widget.NewFormItem("openMSX Machine", machineSelect),
		widget.NewFormItem("", machineHint),
		widget.NewFormItem("openMSX Ext 1", openMSXExtSelects[0]),
		widget.NewFormItem("openMSX Ext 2", openMSXExtSelects[1]),
		widget.NewFormItem("openMSX Ext 3", openMSXExtSelects[2]),
		widget.NewFormItem("openMSX Ext 4", openMSXExtSelects[3]),
		widget.NewFormItem("", extHint),
		widget.NewFormItem("msxbas2rom Path", bindDirectoryPicker("msxbas2rom path", settingMSXBas2RomExeKey, msxbas2romExe)),
		widget.NewFormItem("BASIC Dignified Path", bindDirectoryPicker("BASIC Dignified path", settingBasicDignifiedExeKey, basicDignifiedExe)),
		widget.NewFormItem("MSX Encoding Path", bindDirectoryPicker("MSX Encoding path", settingMSXEncodingExeKey, msxEncodingExe)),
	)

	var cfgDialog *dialog.CustomDialog
	saveConfig := func() {

		nextTheme := editorThemeDarkID
		for _, opt := range themeOptions {
			if opt.Label == themeSelect.Selected {
				nextTheme = opt.ID
				break
			}
		}
		e.editorThemeID = nextTheme
		e.applyCurrentEditorTheme()

		if e.store != nil {
			_ = e.store.SetSetting(context.Background(), settingEditorThemeKey, e.editorThemeID)
			_ = e.store.SetSetting(context.Background(), settingWS7BaseDirKey, normalizeWS7BaseDirectory(ws7BaseDir.Text))
			_ = e.store.SetSetting(context.Background(), settingOpenMSXExeKey, strings.TrimSpace(openMSXExe.Text))
			_ = e.store.SetSetting(context.Background(), settingOpenMSXMachineKey, normalizeOpenMSXResourceName(machineSelect.Selected))
			for i, key := range openMSXExtensionSettingKeys() {
				_ = e.store.SetSetting(context.Background(), key, normalizeOpenMSXResourceName(openMSXExtSelects[i].Selected))
			}
			_ = e.store.SetSetting(context.Background(), settingMSXBas2RomExeKey, strings.TrimSpace(msxbas2romExe.Text))
			_ = e.store.SetSetting(context.Background(), settingBasicDignifiedExeKey, strings.TrimSpace(basicDignifiedExe.Text))
			_ = e.store.SetSetting(context.Background(), settingMSXEncodingExeKey, strings.TrimSpace(msxEncodingExe.Text))
		}

		if e.status != nil {
			e.status.SetText("Configuration saved")
		}
		if cfgDialog != nil {
			cfgDialog.Hide()
		}
	}

	saveBtn := widget.NewButton("Save", saveConfig)
	cancelBtn := widget.NewButton("Cancel", func() {
		if cfgDialog != nil {
			cfgDialog.Hide()
		}
	})
	actions := container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn)
	content := container.NewBorder(nil, actions, nil, nil, container.NewVScroll(form))
	cfgDialog = dialog.NewCustomWithoutButtons("Configure", content, e.window)
	cfgDialog.Resize(fyne.NewSize(1120, 700))
	cfgDialog.Show()
}

func (e *editorUI) cmdBasicRenum() {
	if e.activeTab == nil || e.activeTab.entry == nil {
		return
	}

	startDefault, incDefault, fromDefault := e.renumDefaultsForActiveTab()

	startEntry := widget.NewEntry()
	startEntry.SetText(strconv.Itoa(startDefault))
	incEntry := widget.NewEntry()
	incEntry.SetText(strconv.Itoa(incDefault))
	fromEntry := widget.NewEntry()
	fromEntry.SetText(strconv.Itoa(fromDefault))
	strictCheck := widget.NewCheck("Strict MSX parity (fail on undefined flow references)", nil)

	dialog.ShowForm("BASIC RENUM", "Apply", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Start Line", startEntry),
		widget.NewFormItem("Increment", incEntry),
		widget.NewFormItem("Renumber From Line", fromEntry),
		widget.NewFormItem("Mode", strictCheck),
	}, func(ok bool) {
		if !ok {
			return
		}

		startLine, err := strconv.Atoi(strings.TrimSpace(startEntry.Text))
		if err != nil || startLine <= 0 {
			dialog.ShowInformation("BASIC RENUM", "Start Line must be an integer greater than zero.", e.window)
			return
		}
		increment, err := strconv.Atoi(strings.TrimSpace(incEntry.Text))
		if err != nil || increment <= 0 {
			dialog.ShowInformation("BASIC RENUM", "Increment must be an integer greater than zero.", e.window)
			return
		}
		fromLine, err := strconv.Atoi(strings.TrimSpace(fromEntry.Text))
		if err != nil || fromLine < 0 {
			dialog.ShowInformation("BASIC RENUM", "Renumber From Line must be an integer zero or greater.", e.window)
			return
		}

		opts := renum.Options{StartLine: startLine, Increment: increment, FromLine: fromLine, StrictMSXParity: strictCheck.Checked}
		result, renumErr := renum.Renumber(e.activeTab.entry.Text, opts)
		if renumErr != nil {
			dialog.ShowError(renumErr, e.window)
			if e.status != nil {
				e.status.SetText("BASIC RENUM failed (strict parity)")
			}
			return
		}
		e.activeTab.entry.SetText(result.Text)
		e.recordProgramSnapshot(e.activeTab, &opts)
		stats := renum.SummarizeWarnings(result.UndefinedRefs)
		if stats.Total > 0 {
			dialog.ShowInformation("BASIC RENUM Warnings", formatRenumWarnings(result.UndefinedRefs), e.window)
		}
		if e.status != nil {
			modeSuffix := ""
			if opts.StrictMSXParity {
				modeSuffix = " [strict parity]"
			}
			status := fmt.Sprintf("BASIC RENUM complete (%d line(s) renumbered)%s", result.RenumberedLines, modeSuffix)
			if stats.Total > 0 {
				parts := make([]string, 0, 2)
				if stats.Flow > 0 {
					parts = append(parts, fmt.Sprintf("%d flow warning(s)", stats.Flow))
				}
				if stats.Listing > 0 {
					parts = append(parts, fmt.Sprintf("%d listing info item(s)", stats.Listing))
				}
				status = fmt.Sprintf("BASIC RENUM complete (%d line(s) renumbered, %s)%s", result.RenumberedLines, strings.Join(parts, ", "), modeSuffix)
			}
			e.status.SetText(status)
		}
	}, e.window)
}

func (e *editorUI) cmdBasicDelete() {
	if e.activeTab == nil || e.activeTab.entry == nil {
		return
	}

	text := e.activeTab.entry.Text
	firstLine, lastLine, ok := basicProgramLineRange(text)
	if !ok {
		dialog.ShowInformation("BASIC DELETE", "The program has no numbered BASIC lines.", e.window)
		return
	}

	modeOptions := []string{
		"Current line only",
		"Entire program",
		"Cursor to end",
		"Cursor to beginning",
		"Line range",
	}
	modeSelect := widget.NewSelect(modeOptions, nil)
	modeSelect.SetSelected(modeOptions[0])

	startEntry := widget.NewEntry()
	endEntry := widget.NewEntry()
	if currentLine, ok := basicLineNumberAtRow(text, e.activeTab.entry.CursorRow); ok {
		startEntry.SetText(strconv.Itoa(currentLine))
		endEntry.SetText(strconv.Itoa(currentLine))
	}

	dialog.ShowForm("BASIC DELETE", "Apply", "Cancel", []*widget.FormItem{
		widget.NewFormItem("Mode", modeSelect),
		widget.NewFormItem("Start Line", startEntry),
		widget.NewFormItem("End Line", endEntry),
		widget.NewFormItem("Notes", widget.NewLabel("Use Start Line and End Line only for Line range.")),
	}, func(ok bool) {
		if !ok {
			return
		}

		deleteFrom, deleteTo, err := e.resolveBasicDeleteScope(modeSelect.Selected, strings.TrimSpace(startEntry.Text), strings.TrimSpace(endEntry.Text), firstLine, lastLine)
		if err != nil {
			dialog.ShowInformation("BASIC DELETE", err.Error(), e.window)
			return
		}

		result, deleteErr := renum.DeleteRange(text, deleteFrom, deleteTo)
		if deleteErr != nil {
			dialog.ShowError(deleteErr, e.window)
			return
		}
		if len(result.BlockingRefs) > 0 {
			dialog.ShowInformation("BASIC DELETE Blocked", formatBasicDeleteWarnings(deleteFrom, deleteTo, result.BlockingRefs), e.window)
			if e.status != nil {
				stats := renum.SummarizeReferences(result.BlockingRefs)
				parts := make([]string, 0, 2)
				if stats.Flow > 0 {
					parts = append(parts, fmt.Sprintf("%d flow reference(s)", stats.Flow))
				}
				if stats.Listing > 0 {
					parts = append(parts, fmt.Sprintf("%d listing reference(s)", stats.Listing))
				}
				e.status.SetText(fmt.Sprintf("BASIC DELETE blocked (%s)", strings.Join(parts, ", ")))
			}
			return
		}
		if result.DeletedLines == 0 {
			dialog.ShowInformation("BASIC DELETE", "No numbered BASIC lines matched the selected delete scope.", e.window)
			return
		}

		e.activeTab.entry.SetText(result.Text)
		e.recordProgramSnapshot(e.activeTab, nil)
		if e.status != nil {
			e.status.SetText(fmt.Sprintf("BASIC DELETE complete (%d line(s) deleted)", result.DeletedLines))
		}
	}, e.window)
}

func formatRenumWarnings(refs []renum.UndefinedReference) string {
	if len(refs) == 0 {
		return "No warnings."
	}
	stats := renum.SummarizeWarnings(refs)
	flow := make([]renum.UndefinedReference, 0, stats.Flow)
	listing := make([]renum.UndefinedReference, 0, stats.Listing)
	for _, ref := range refs {
		switch renumWarningCategory(ref) {
		case renum.WarningCategoryListing:
			listing = append(listing, ref)
		default:
			flow = append(flow, ref)
		}
	}

	const maxItems = 12
	shown := 0
	lines := []string{"Some references point to lines not found in the program:"}
	shown = appendRenumWarningSection(&lines, "Flow warnings (severity: warning)", flow, maxItems, shown)
	shown = appendRenumWarningSection(&lines, "Listing notices (severity: info)", listing, maxItems, shown)
	if len(refs) > shown {
		lines = append(lines, fmt.Sprintf("- ...and %d more item(s)", len(refs)-shown))
	}
	return strings.Join(lines, "\n")
}

func formatBasicDeleteWarnings(deleteFrom, deleteTo int, refs []renum.Reference) string {
	if len(refs) == 0 {
		return "No warnings."
	}
	stats := renum.SummarizeReferences(refs)
	flow := make([]renum.Reference, 0, stats.Flow)
	listing := make([]renum.Reference, 0, stats.Listing)
	for _, ref := range refs {
		switch ref.Category {
		case renum.WarningCategoryListing:
			listing = append(listing, ref)
		default:
			flow = append(flow, ref)
		}
	}

	const maxItems = 12
	shown := 0
	lines := []string{fmt.Sprintf("Deletion blocked: lines %d to %d are still referenced by remaining code.", deleteFrom, deleteTo)}
	shown = appendBasicDeleteWarningSection(&lines, "Flow references (severity: warning)", flow, maxItems, shown)
	shown = appendBasicDeleteWarningSection(&lines, "Listing references (severity: info)", listing, maxItems, shown)
	if len(refs) > shown {
		lines = append(lines, fmt.Sprintf("- ...and %d more item(s)", len(refs)-shown))
	}
	return strings.Join(lines, "\n")
}

func appendBasicDeleteWarningSection(lines *[]string, title string, refs []renum.Reference, maxItems, shown int) int {
	if len(refs) == 0 || shown >= maxItems {
		return shown
	}
	*lines = append(*lines, title)
	for _, ref := range refs {
		if shown >= maxItems {
			break
		}
		*lines = append(*lines, fmt.Sprintf("- Source %d: %s %d", ref.SourceLine, strings.ToUpper(strings.TrimSpace(ref.Command)), ref.Target))
		shown++
	}
	return shown
}

func appendRenumWarningSection(lines *[]string, title string, refs []renum.UndefinedReference, maxItems, shown int) int {
	if len(refs) == 0 || shown >= maxItems {
		return shown
	}
	*lines = append(*lines, title)
	for _, ref := range refs {
		if shown >= maxItems {
			break
		}
		*lines = append(*lines, fmt.Sprintf("- Source %d: %s %d", ref.SourceLine, strings.ToUpper(strings.TrimSpace(ref.Command)), ref.Target))
		shown++
	}
	return shown
}

func renumWarningCategory(ref renum.UndefinedReference) renum.WarningCategory {
	if ref.Category != "" {
		return ref.Category
	}
	switch strings.ToUpper(strings.TrimSpace(ref.Command)) {
	case "LIST", "LLIST":
		return renum.WarningCategoryListing
	default:
		return renum.WarningCategoryFlow
	}
}

func (e *editorUI) renumDefaultsForActiveTab() (int, int, int) {
	start := defaultRenumStartLine
	inc := defaultRenumIncrement
	from := defaultRenumFromLine
	if e.activeTab == nil || e.store == nil {
		return start, inc, from
	}
	fileName, filePath := tabProgramIdentity(e.activeTab)
	snapshot, err := e.store.GetLatestProgramSnapshot(context.Background(), fileName, filePath)
	if err != nil {
		return start, inc, from
	}
	if snapshot.RenumStart > 0 {
		start = snapshot.RenumStart
	}
	if snapshot.RenumIncrement > 0 {
		inc = snapshot.RenumIncrement
	}
	if snapshot.RenumFromLine >= 0 {
		from = snapshot.RenumFromLine
	}
	return start, inc, from
}

func (e *editorUI) resolveBasicDeleteScope(mode, startText, endText string, firstLine, lastLine int) (int, int, error) {
	text := e.activeTab.entry.Text
	row := e.activeTab.entry.CursorRow
	switch mode {
	case "Current line only":
		line, ok := basicLineNumberAtRow(text, row)
		if !ok {
			return 0, 0, fmt.Errorf("the cursor is not on a numbered BASIC line")
		}
		return line, line, nil
	case "Entire program":
		return firstLine, lastLine, nil
	case "Cursor to end":
		line, ok := basicLineNumberOnOrAfterRow(text, row)
		if !ok {
			return 0, 0, fmt.Errorf("no numbered BASIC line was found at or after the cursor")
		}
		return line, lastLine, nil
	case "Cursor to beginning":
		line, ok := basicLineNumberOnOrBeforeRow(text, row)
		if !ok {
			return 0, 0, fmt.Errorf("no numbered BASIC line was found at or before the cursor")
		}
		return firstLine, line, nil
	case "Line range":
		startLine, err := strconv.Atoi(startText)
		if err != nil || startLine <= 0 {
			return 0, 0, fmt.Errorf("Start Line must be an integer greater than zero")
		}
		endLine, err := strconv.Atoi(endText)
		if err != nil || endLine <= 0 {
			return 0, 0, fmt.Errorf("End Line must be an integer greater than zero")
		}
		if startLine > endLine {
			return 0, 0, fmt.Errorf("Start Line must be less than or equal to End Line")
		}
		return startLine, endLine, nil
	default:
		return 0, 0, fmt.Errorf("choose a delete mode")
	}
}

func (e *editorUI) recordProgramSnapshot(tab *editorTab, renumOpts *renum.Options) {
	if e.store == nil || tab == nil || tab.entry == nil {
		return
	}
	fileName, filePath := tabProgramIdentity(tab)
	sha := sha1.Sum([]byte(tab.entry.Text))
	snapshot := sqlite.ProgramSnapshot{
		FileName:     fileName,
		FilePath:     filePath,
		ContentSHA1:  hex.EncodeToString(sha[:]),
		ContentBytes: len(tab.entry.Text),
	}
	if renumOpts != nil {
		snapshot.RenumStart = renumOpts.StartLine
		snapshot.RenumIncrement = renumOpts.Increment
		snapshot.RenumFromLine = renumOpts.FromLine
	}
	_ = e.store.UpsertProgramSnapshot(context.Background(), snapshot)
}

// cmdUndo pops the most recent undo state for the active tab and restores it.
func (e *editorUI) cmdUndo() {
	tab := e.activeTab
	if tab == nil || tab.entry == nil {
		if e.status != nil {
			e.status.SetText("Undo: no active editor")
		}
		return
	}
	if len(tab.undoStack) == 0 {
		if e.status != nil {
			e.status.SetText("Undo: nothing more to undo")
		}
		return
	}
	state := tab.undoStack[len(tab.undoStack)-1]
	tab.undoStack = tab.undoStack[:len(tab.undoStack)-1]

	tab.undoing = true
	tab.entry.SetText(state.text)
	tab.lastKnownText = state.text
	tab.undoing = false

	e.applyCursorPosition(state.cursorRow, state.cursorCol)

	remaining := len(tab.undoStack)
	if e.status != nil {
		e.status.SetText(fmt.Sprintf("Undo: restored (%d level(s) remaining)", remaining))
	}
}

func tabProgramIdentity(tab *editorTab) (string, string) {
	fileName := tab.name
	if strings.TrimSpace(tab.filePath) != "" {
		fileName = filepath.Base(tab.filePath)
	}
	filePath := ""
	if strings.TrimSpace(tab.filePath) != "" {
		filePath = filepath.Clean(tab.filePath)
	}
	return fileName, filePath
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

func (e *editorUI) resolveConfiguredToolPathFromSettings(toolID string) (string, error) {
	if e.store == nil {
		return "", fmt.Errorf("configuration storage is unavailable")
	}
	raw, _ := e.store.GetSetting(context.Background(), toolID)
	resolved, err := resolveConfiguredToolPath(toolID, raw)
	if err != nil {
		return "", fmt.Errorf("%s: %w", configuredToolLabel(toolID), err)
	}
	return resolved, nil
}

func (e *editorUI) activeFileArgs() []string {
	if e.activeTab == nil || strings.TrimSpace(e.activeTab.filePath) == "" {
		return nil
	}
	return []string{filepath.Clean(e.activeTab.filePath)}
}

func (e *editorUI) runConfiguredTool(toolID string, args []string, detached bool) {
	label := configuredToolLabel(toolID)
	resolved, err := e.resolveConfiguredToolPathFromSettings(toolID)
	if err != nil {
		msg := err.Error() + "\n\nUse Utilities -> Configure... to set the tool path."
		dialog.ShowInformation(label, msg, e.window)
		if e.status != nil {
			e.status.SetText(label + ": not configured")
		}
		return
	}

	name, cmdArgs, workDir := buildConfiguredToolCommand(toolID, resolved, args)
	cmd := exec.Command(name, cmdArgs...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	if detached {
		if err := cmd.Start(); err != nil {
			dialog.ShowError(err, e.window)
			if e.status != nil {
				e.status.SetText(label + ": failed to start")
			}
			return
		}
		if e.status != nil {
			e.status.SetText(label + ": started")
		}
		return
	}

	output, runErr := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if text == "" {
		text = "(no output)"
	}
	if runErr != nil {
		text += "\n\nError: " + runErr.Error()
	}

	result := widget.NewMultiLineEntry()
	result.SetMinRowsVisible(20)
	result.SetText(text)
	result.Disable()
	dialog.ShowCustom(label+" Output", "Close", result, e.window)

	if e.status != nil {
		if runErr != nil {
			e.status.SetText(label + ": failed")
		} else {
			e.status.SetText(label + ": done")
		}
	}
}

func (e *editorUI) cmdLaunchOpenMSX() {
	e.openOpenMSXBridge()
}

func (e *editorUI) cmdLaunchMSXBas2Rom() {
	e.runConfiguredTool(settingMSXBas2RomExeKey, e.activeFileArgs(), false)
}

func (e *editorUI) cmdLaunchBasicDignified() {
	e.runConfiguredTool(settingBasicDignifiedExeKey, e.activeFileArgs(), false)
}

func (e *editorUI) cmdLaunchMSXEncoding() {
	e.runConfiguredTool(settingMSXEncodingExeKey, e.activeFileArgs(), false)
}

func (e *editorUI) cmdExitMSXide() {
	e.requestAppExit()
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

func basicProgramLineRange(text string) (int, int, bool) {
	lines := strings.Split(text, "\n")
	first := 0
	last := 0
	for _, raw := range lines {
		lineNumber, ok := parseBasicLineNumber(raw)
		if !ok {
			continue
		}
		if first == 0 {
			first = lineNumber
		}
		last = lineNumber
	}
	if first == 0 {
		return 0, 0, false
	}
	return first, last, true
}

func basicLineNumberAtRow(text string, row int) (int, bool) {
	lines := strings.Split(text, "\n")
	if row < 0 || row >= len(lines) {
		return 0, false
	}
	return parseBasicLineNumber(lines[row])
}

func basicLineNumberOnOrBeforeRow(text string, row int) (int, bool) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0, false
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	for i := row; i >= 0; i-- {
		if lineNumber, ok := parseBasicLineNumber(lines[i]); ok {
			return lineNumber, true
		}
	}
	return 0, false
}

func basicLineNumberOnOrAfterRow(text string, row int) (int, bool) {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return 0, false
	}
	if row < 0 {
		row = 0
	}
	for i := row; i < len(lines); i++ {
		if lineNumber, ok := parseBasicLineNumber(lines[i]); ok {
			return lineNumber, true
		}
	}
	return 0, false
}

func parseBasicLineNumber(raw string) (int, bool) {
	parts := basicLineNumberRE.FindStringSubmatch(raw)
	if len(parts) < 2 {
		return 0, false
	}
	lineNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}
	return lineNumber, true
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

func (e *editorUI) cmdConvertUppercase() {
	e.cmdConvertCase("Uppercase", "Ctrl+K,\"", strings.ToUpper)
}

func (e *editorUI) cmdConvertLowercase() {
	e.cmdConvertCase("Lowercase", "Ctrl+K,'", strings.ToLower)
}

func (e *editorUI) cmdConvertCapitalize() {
	e.cmdConvertCase("Capitalize", "Ctrl+K,.", capitalizeText)
}

func (e *editorUI) composeEditorTextStyle(tab *editorTab) fyne.TextStyle {
	if tab == nil {
		return fyne.TextStyle{}
	}
	return fyne.TextStyle{Bold: tab.isBold, Italic: e.editorFontItalic}
}

func (e *editorUI) applyEditorStyleToTab(tab *editorTab) {
	if tab == nil {
		return
	}
	style := e.composeEditorTextStyle(tab)
	if tab.entry != nil {
		tab.entry.TextStyle = style
		tab.entry.Refresh()
	}
	if tab.ruler != nil {
		tab.ruler.SetTextStyle(style.Bold, style.Italic)
	}
	if tab.lineNums != nil {
		tab.lineNums.SetTextStyle(style.Bold, style.Italic)
	}
	if tab.floatingRuler != nil {
		tab.floatingRuler.SetTextStyle(style.Bold, style.Italic)
	}
}

// cmdStyleBold toggles the bold font style on the active editor tab.
// The entry text style, column ruler, line-number gutter and floating ruler
// are all updated so that their character-size measurements stay in sync.
func (e *editorUI) cmdStyleBold() {
	tab := e.activeTab
	if tab == nil {
		return
	}
	tab.isBold = !tab.isBold
	b := tab.isBold
	e.applyEditorStyleToTab(tab)

	if e.status != nil {
		state := "off"
		if b {
			state = "on"
		}
		e.status.SetText("Bold: " + state + "  (Ctrl+P,B to toggle)")
	}
}

func (e *editorUI) cmdStyleFont() {
	if e.window == nil {
		return
	}

	families := availableEditorFontFamilies()
	familySelect := widget.NewSelect(families, nil)
	familySelect.SetSelected(normalizeEditorFontFamily(e.editorFontFamily))

	weightSelect := widget.NewSelect(editorFontWeightsForFamily(familySelect.Selected), nil)
	weightSelect.SetSelected(normalizeEditorFontWeight(familySelect.Selected, e.editorFontWeight))

	sizeEntry := widget.NewEntry()
	sizeEntry.SetText(strconv.Itoa(int(normalizeEditorFontSize(e.editorFontSize))))

	italicCheck := widget.NewCheck("Italic", nil)
	italicCheck.SetChecked(e.editorFontItalic)
	if !editorFontFamilySupportsItalic(familySelect.Selected) {
		italicCheck.SetChecked(false)
		italicCheck.Disable()
	}

	widthSelect := widget.NewSelect([]string{"Normal", "Narrow (not available)"}, nil)
	widthSelect.SetSelected("Normal")

	familySelect.OnChanged = func(next string) {
		next = normalizeEditorFontFamily(next)
		weightOptions := editorFontWeightsForFamily(next)
		weightSelect.Options = weightOptions
		weightSelect.SetSelected(normalizeEditorFontWeight(next, weightSelect.Selected))
		weightSelect.Refresh()

		supportsItalic := editorFontFamilySupportsItalic(next)
		if supportsItalic {
			italicCheck.Enable()
		} else {
			italicCheck.Disable()
			italicCheck.SetChecked(false)
		}
	}

	note := widget.NewLabel("Only bundled fixed-width/programming fonts are listed. Narrow variants are currently unavailable.")
	note.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Font Family", familySelect),
			widget.NewFormItem("Size", sizeEntry),
			widget.NewFormItem("Weight", weightSelect),
			widget.NewFormItem("Style", italicCheck),
			widget.NewFormItem("Width", widthSelect),
		),
		note,
	)

	dlg := dialog.NewCustomConfirm("Font", "Apply", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}

		nextFamily := normalizeEditorFontFamily(familySelect.Selected)
		nextWeight := normalizeEditorFontWeight(nextFamily, weightSelect.Selected)
		nextItalic := italicCheck.Checked && editorFontFamilySupportsItalic(nextFamily)

		sizeNum, err := strconv.ParseFloat(strings.TrimSpace(sizeEntry.Text), 32)
		if err != nil {
			dialog.ShowInformation("Font", "Size must be a valid number (8..48).", e.window)
			return
		}
		nextSize := normalizeEditorFontSize(float32(sizeNum))
		if nextSize != float32(sizeNum) {
			dialog.ShowInformation("Font", "Size must be between 8 and 48.", e.window)
			return
		}

		e.editorFontFamily = nextFamily
		e.editorFontWeight = nextWeight
		e.editorFontSize = nextSize
		e.editorFontItalic = nextItalic

		e.applyCurrentEditorTheme()

		for _, tab := range e.tabState {
			e.applyEditorStyleToTab(tab)
		}

		if e.store != nil {
			_ = e.store.SetSetting(context.Background(), settingEditorFontFamilyKey, e.editorFontFamily)
			_ = e.store.SetSetting(context.Background(), settingEditorFontWeightKey, e.editorFontWeight)
			_ = e.store.SetSetting(context.Background(), settingEditorFontSizeKey, fmt.Sprintf("%.0f", e.editorFontSize))
			_ = e.store.SetSetting(context.Background(), settingEditorFontItalicKey, strconv.FormatBool(e.editorFontItalic))
		}

		if e.status != nil {
			status := fmt.Sprintf("Font: %s %s %.0fpt", e.editorFontFamily, e.editorFontWeight, e.editorFontSize)
			if e.editorFontItalic {
				status += " Italic"
			}
			if widthSelect.Selected != "Normal" {
				status += " (narrow unavailable)"
			}
			e.status.SetText(status)
		}
	}, e.window)
	dlg.Resize(fyne.NewSize(560, 360))
	dlg.Show()
}

func (e *editorUI) cmdConvertCase(mode, chord string, transform func(string) string) {
	if e.entry == nil {
		if e.status != nil {
			e.status.SetText("Convert Case: no active editor")
		}
		return
	}
	if transform == nil {
		return
	}

	text := e.entry.Text
	start, end, scope, reason, ok := e.resolveConvertCaseRange(text)
	if !ok {
		if e.status != nil {
			switch reason {
			case "empty_document":
				e.status.SetText(chord + ": document is empty")
			case "empty_block":
				e.status.SetText(chord + ": empty block (B and K at same position)")
			case "empty_line":
				e.status.SetText(chord + ": current line is empty")
			default:
				e.status.SetText(chord + ": nothing to convert")
			}
		}
		return
	}

	oldCursor := e.cursorByteOffset()
	if oldCursor < 0 {
		oldCursor = cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
	}
	oldCursor = clampOffset(oldCursor, len(text))

	chunk := text[start:end]
	converted := transform(chunk)
	if converted == chunk {
		if e.status != nil {
			e.status.SetText(fmt.Sprintf("Convert Case (%s): no changes in %s", mode, scope))
		}
		return
	}

	newText := text[:start] + converted + text[end:]
	e.entry.SetText(newText)

	delta := len(converted) - (end - start)
	newCursor := oldCursor
	if oldCursor >= end {
		newCursor = oldCursor + delta
	} else if oldCursor > start {
		newCursor = start + len(converted)
	}
	newCursor = clampOffset(newCursor, len(newText))
	row, col := offsetToRowCol(newText, newCursor)
	e.applyCursorPosition(row, col)

	if scope == "block" && e.activeTab != nil && e.activeTab.hasBlockBegin && e.activeTab.hasBlockEnd {
		e.activeTab.blockBegin = start
		e.activeTab.blockEnd = start + len(converted)
	}

	if e.status != nil {
		e.status.SetText(fmt.Sprintf("Convert Case (%s): applied to %s", mode, scope))
	}
}

func (e *editorUI) resolveConvertCaseRange(text string) (start, end int, scope, reason string, ok bool) {
	if text == "" {
		return 0, 0, "", "empty_document", false
	}

	if bStart, bEnd, bOk := e.activeBlockRange(); bOk {
		return bStart, bEnd, "block", "", true
	}
	if e.activeTab != nil && e.activeTab.hasBlockBegin && e.activeTab.hasBlockEnd {
		return 0, 0, "", "empty_block", false
	}

	if e.entry != nil {
		selected := e.entry.SelectedText()
		if selected != "" {
			cursor := e.cursorByteOffset()
			if cursor < 0 {
				cursor = cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
			}
			if sStart, sEnd, sOk := findSelectionRange(text, selected, cursor); sOk {
				return sStart, sEnd, "selection", "", true
			}
		}
	}

	if e.entry != nil {
		cursor := e.cursorByteOffset()
		if cursor < 0 {
			cursor = cursorOffset(text, e.entry.CursorRow, e.entry.CursorColumn)
		}
		lStart, lEnd := currentLineBounds(text, cursor)
		if lEnd <= lStart {
			return 0, 0, "", "empty_line", false
		}
		if lStart >= 0 {
			return lStart, lEnd, "current line", "", true
		}
	}

	return 0, 0, "", "", false
}

func currentLineBounds(text string, cursor int) (start, end int) {
	if text == "" {
		return 0, 0
	}
	cursor = clampOffset(cursor, len(text))
	start = strings.LastIndex(text[:cursor], "\n") + 1
	lineEndRel := strings.Index(text[cursor:], "\n")
	end = len(text)
	if lineEndRel >= 0 {
		end = cursor + lineEndRel
	}
	return start, end
}

func currentLineRange(text string, cursor int) (start, end int, ok bool) {
	start, end = currentLineBounds(text, cursor)
	if end <= start {
		return 0, 0, false
	}
	return start, end, true
}

func findSelectionRange(text, selected string, cursor int) (start, end int, ok bool) {
	if selected == "" || text == "" {
		return 0, 0, false
	}
	cursor = clampOffset(cursor, len(text))

	bestStart, bestEnd := 0, 0
	bestScore := 99
	bestDist := int(^uint(0) >> 1)

	for from := 0; from <= len(text); {
		idx := strings.Index(text[from:], selected)
		if idx < 0 {
			break
		}
		s := from + idx
		e := s + len(selected)

		score := 2
		if s == cursor || e == cursor {
			score = 0
		} else if cursor > s && cursor < e {
			score = 1
		}
		dist := absInt(((s + e) / 2) - cursor)
		if score < bestScore || (score == bestScore && dist < bestDist) {
			bestStart, bestEnd = s, e
			bestScore = score
			bestDist = dist
		}
		from = s + 1
	}

	if bestScore == 99 {
		return 0, 0, false
	}
	return bestStart, bestEnd, true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func capitalizeText(text string) string {
	if text == "" {
		return text
	}
	b := strings.Builder{}
	b.Grow(len(text))
	startWord := true
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if startWord {
				b.WriteRune(unicode.ToUpper(r))
				startWord = false
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
			continue
		}
		startWord = true
		b.WriteRune(r)
	}
	return b.String()
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

func (e *editorUI) insertTextAtCursor(text string, label string) {
	if e.entry == nil {
		return
	}
	all := e.entry.Text
	pos := cursorOffset(all, e.entry.CursorRow, e.entry.CursorColumn)
	if pos < 0 {
		pos = len(all)
	}
	e.entry.SetText(all[:pos] + text + all[pos:])
	if e.status != nil {
		e.status.SetText("Inserted " + label)
	}
}

func (e *editorUI) cmdInsertTodayDate() {
	dateStr := time.Now().Format("January 02, 2006")
	e.insertTextAtCursor(dateStr, "Today's Date")
}

func (e *editorUI) cmdInsertCurrentTime() {
	timeStr := time.Now().Format("15:04:05")
	e.insertTextAtCursor(timeStr, "Current Time")
}

func (e *editorUI) cmdInsertExtendedChar() {
	// Re-implemented in extended_char.go
	e.showExtendedCharPicker()
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
