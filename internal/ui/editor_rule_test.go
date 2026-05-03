package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/input"
)

func makeSplitViewTestTab(name string) *editorTab {
	tab := &editorTab{
		entry:         newCursorEntry(),
		ruler:         newRulerWidget(),
		floatingRuler: newFloatingRulerWidget(),
		lineNums:      newLineNumbersWidget(),
		status:        widget.NewLabel(""),
		blockTag:      widget.NewLabel(""),
		clipTag:       widget.NewLabel(""),
	}
	tab.item = container.NewTabItem(name, widget.NewLabel("placeholder"))
	return tab
}

func containsCanvasObject(objects []fyne.CanvasObject, target fyne.CanvasObject) bool {
	for _, obj := range objects {
		if obj == target {
			return true
		}
	}
	return false
}

func TestSetRuleModeShowsAndHidesRuler(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		inEditor:  true,
		status:    widget.NewLabel(""),
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		activeTab: tab,
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	tab.item.Content = ui.tabEditorContent(tab)

	border, ok := tab.item.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("expected border content, got %T", tab.item.Content)
	}
	if containsCanvasObject(border.Objects, tab.floatingRuler) {
		t.Fatalf("expected floating ruler hidden by default")
	}

	ui.setRuleMode(tab, true)
	stack, ok := tab.item.Content.(*fyne.Container)
	if !ok {
		t.Fatalf("expected stacked content after enabling rule, got %T", tab.item.Content)
	}
	if !containsCanvasObject(stack.Objects, tab.floatingRuler) {
		t.Fatalf("expected floating ruler visible when RULE is enabled")
	}
	if tab.floatingRuler.originCharPos != tab.floatingRuler.cursorCharPos {
		t.Fatalf("expected origin and cursor aligned on activation, got origin=%d cursor=%d", tab.floatingRuler.originCharPos, tab.floatingRuler.cursorCharPos)
	}

	handled := tab.entry.onKeyBeforeInput(&fyne.KeyEvent{Name: fyne.KeyEscape})
	if !handled {
		t.Fatalf("expected ESC to be handled while RULE is active")
	}
	if tab.ruleMode {
		t.Fatalf("expected RULE mode to be disabled after ESC")
	}
	if ui.status.Text != "RULE: off" {
		t.Fatalf("status = %q, want %q", ui.status.Text, "RULE: off")
	}
}

func TestAbsoluteCharPos(t *testing.T) {
	text := "Hello World\nGoodbye"
	if got := absoluteCharPos(text, 0, 3); got != 3 {
		t.Fatalf("row0/col3 = %d, want 3", got)
	}
	if got := absoluteCharPos(text, 1, 2); got != 14 {
		t.Fatalf("row1/col2 = %d, want 14", got)
	}
	if got := absoluteCharPos(text, 20, 200); got != 19 {
		t.Fatalf("clamped out-of-range = %d, want 19", got)
	}
}

func TestBMarksBlockWhenRuleActive(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	tab.entry.Text = "ABC\nDE"
	ui := &editorUI{
		inEditor:  true,
		status:    widget.NewLabel(""),
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		activeTab: tab,
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	ui.setRuleMode(tab, true)

	// First B: mark block start at absolute char 1 (B)
	tab.cursorRow = 0
	tab.cursorCol = 1
	handled := tab.entry.onRuneBeforeInput('B')
	if !handled {
		t.Fatal("expected B to be handled while RULE is active")
	}
	if !tab.floatingRuler.hasBlockStart || tab.floatingRuler.hasBlockEnd {
		t.Fatal("expected only block start after first B")
	}

	// Second B: mark block end at absolute char 5 (E)
	tab.cursorRow = 1
	tab.cursorCol = 1
	handled = tab.entry.onRuneBeforeInput('B')
	if !handled {
		t.Fatal("expected second B to be handled while RULE is active")
	}
	if !tab.floatingRuler.hasBlockEnd {
		t.Fatal("expected block end after second B")
	}

	if got := tab.floatingRuler.blockStartPos; got != 1 {
		t.Fatalf("block start = %d, want 1", got)
	}
	if got := tab.floatingRuler.blockEndPos; got != 5 {
		t.Fatalf("block end = %d, want 5", got)
	}
	if got := ui.status.Text; got != "RULE: block 1..5 (5 chars)" {
		t.Fatalf("status = %q, want %q", got, "RULE: block 1..5 (5 chars)")
	}
}

func TestEditorUtilitiesMenuContainsRule(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 6 {
		t.Fatalf("expected editor main menu with Utilities")
	}
	utilities := menu.Items[5]
	if utilities == nil || utilities.Label != "Utilities" {
		t.Fatalf("expected Utilities menu at top-level index 5")
	}
	if len(utilities.Items) < 1 {
		t.Fatalf("expected Utilities items")
	}
	if got := utilities.Items[0].Label; got != "RULE                       Ctrl+Q,R  ESC to exit" {
		t.Fatalf("first Utilities item = %q, want RULE item", got)
	}
}

func TestEditorUtilitiesMenuContainsCalculator(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 6 {
		t.Fatalf("expected editor main menu with Utilities")
	}
	utilities := menu.Items[5]
	if utilities == nil || utilities.Label != "Utilities" {
		t.Fatalf("expected Utilities menu at top-level index 5")
	}
	if len(utilities.Items) < 2 {
		t.Fatalf("expected Utilities to include calculator item")
	}
	if got := utilities.Items[1].Label; got != "Calculator                 Ctrl+Q,M" {
		t.Fatalf("second Utilities item = %q, want Calculator item", got)
	}
}

func TestEditorRunMenuContainsExpectedItems(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 5 {
		t.Fatalf("expected editor main menu with Run")
	}
	runMenu := menu.Items[4]
	if runMenu == nil || runMenu.Label != "Run" {
		t.Fatalf("expected Run menu at top-level index 4")
	}
	if len(runMenu.Items) < 4 {
		t.Fatalf("expected Run menu items")
	}
	if got := runMenu.Items[0].Label; got != "Execute on openMSX [NI]" {
		t.Fatalf("first Run item = %q, want Execute on openMSX [NI]", got)
	}
	if got := runMenu.Items[1].Label; got != "Make a Disk [NI]" {
		t.Fatalf("second Run item = %q, want Make a Disk [NI]", got)
	}
	if got := runMenu.Items[2].Label; got != "Transpile on BADIG [NI]" {
		t.Fatalf("third Run item = %q, want Transpile on BADIG [NI]", got)
	}
	if got := runMenu.Items[3].Label; got != "Compile on msxbas2rom [NI]" {
		t.Fatalf("fourth Run item = %q, want Compile on msxbas2rom [NI]", got)
	}
	for i, item := range runMenu.Items {
		if item.Disabled {
			t.Fatalf("expected Run item %d to remain visible/enabled for the [NI] placeholder dialog", i)
		}
	}
}

func TestOpeningMenuPlaceholderItemsShowNI(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeOpeningMenu()
	if menu == nil || len(menu.Items) < 3 {
		t.Fatalf("expected opening main menu")
	}
	fileMenu := menu.Items[0]
	if fileMenu == nil || fileMenu.Label != "File" {
		t.Fatalf("expected File menu at top-level index 0")
	}
	if got := fileMenu.Items[4].Label; got != "Print... [NI]             P" {
		t.Fatalf("opening File print item = %q, want Print... [NI]             P", got)
	}
	if got := fileMenu.Items[5].Label; got != "Print from keyboard... [NI] K" {
		t.Fatalf("opening File print-from-keyboard item = %q, want Print from keyboard... [NI] K", got)
	}
	utilities := menu.Items[1]
	if utilities == nil || utilities.Label != "Utilities" {
		t.Fatalf("expected Utilities menu at top-level index 1")
	}
	macros := utilities.Items[0]
	if macros == nil || macros.Label != "Macros" || macros.ChildMenu == nil {
		t.Fatalf("expected Macros submenu")
	}
	if got := macros.ChildMenu.Items[0].Label; got != "Play... [NI]               MP" {
		t.Fatalf("macros first item = %q, want Play... [NI]               MP", got)
	}
	additional := menu.Items[2]
	if additional == nil || additional.Label != "Additional" {
		t.Fatalf("expected Additional menu at top-level index 2")
	}
	if got := additional.Items[0].Label; got != "Character Editor... [NI]   AC" {
		t.Fatalf("additional first item = %q, want Character Editor... [NI]   AC", got)
	}
}

func TestEditorEditMenuPlaceholderItemsShowNI(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 2 {
		t.Fatalf("expected editor main menu with Edit")
	}
	edit := menu.Items[1]
	if edit == nil || edit.Label != "Edit" {
		t.Fatalf("expected Edit menu at top-level index 1")
	}
	if got := edit.Items[9].Label; got != "Mark Previous Block [NI]      Ctrl+K,U" {
		t.Fatalf("mark previous block item = %q, want Mark Previous Block [NI]      Ctrl+K,U", got)
	}
	goToOther := edit.Items[17]
	if goToOther == nil || goToOther.Label != "Go to Other" || goToOther.ChildMenu == nil {
		t.Fatalf("expected Go to Other submenu")
	}
	if got := goToOther.ChildMenu.Items[0].Label; got != "Font Tag [NI]                 Ctrl+Q,=" {
		t.Fatalf("go to other first item = %q, want Font Tag [NI]                 Ctrl+Q,=", got)
	}
	noteOptions := edit.Items[21]
	if noteOptions == nil || noteOptions.Label != "Note Options" || noteOptions.ChildMenu == nil {
		t.Fatalf("expected Note Options submenu")
	}
	if got := noteOptions.ChildMenu.Items[0].Label; got != "Starting Number for Note... [NI]" {
		t.Fatalf("note options first item = %q, want Starting Number for Note... [NI]", got)
	}
	editSettings := edit.Items[23]
	if editSettings == nil || editSettings.Label != "Edit Settings" || editSettings.ChildMenu == nil {
		t.Fatalf("expected Edit Settings submenu")
	}
	if got := editSettings.ChildMenu.Items[0].Label; got != "Column Block Mode [NI]        Ctrl+K,N" {
		t.Fatalf("edit settings first item = %q, want Column Block Mode [NI]        Ctrl+K,N", got)
	}
}

func TestEditorStyleMenuContainsExpectedItems(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 4 {
		t.Fatalf("expected editor main menu with Style")
	}
	style := menu.Items[3]
	if style == nil || style.Label != "Style" {
		t.Fatalf("expected Style menu at top-level index 3")
	}
	if len(style.Items) < 4 {
		t.Fatalf("expected Style menu items")
	}
	if got := style.Items[0].Label; got != "Bold                     Ctrl+P,B" {
		t.Fatalf("first Style item = %q, want Bold", got)
	}
	if got := style.Items[1].Label; got != "Font...                  Ctrl+P,=" {
		t.Fatalf("second Style item = %q, want Font", got)
	}
	if got := style.Items[2].Label; got != "Tokenized" {
		t.Fatalf("third Style item = %q, want Tokenized", got)
	}
	if got := style.Items[3].Label; got != "Convert Case" {
		t.Fatalf("fourth Style item = %q, want Convert Case", got)
	}
}

func TestSuggestMSXSaveFileNameUsesBASWhenTokenized(t *testing.T) {
	if got := suggestMSXSaveFileName("", "", true); got != "untitled.bas" {
		t.Fatalf("suggested = %q, want untitled.bas", got)
	}
	if got := suggestMSXSaveFileName("", "demo", true); got != "demo.bas" {
		t.Fatalf("suggested = %q, want demo.bas", got)
	}
	if got := suggestMSXSaveFileName("", "demo.asc", true); got != "demo.bas" {
		t.Fatalf("suggested = %q, want demo.bas", got)
	}
}

func TestEditorStyleMenuTokenizedItemCheckedWhenEnabled(t *testing.T) {
	ui := &editorUI{saveTokenized: true}
	menu := ui.makeEditorMenu()
	style := menu.Items[3]
	if len(style.Items) < 3 {
		t.Fatalf("expected Style menu items")
	}
	if !style.Items[2].Checked {
		t.Fatal("expected Tokenized menu item to be checked")
	}
}

func TestBindActiveTabSyncsTokenizedSaveState(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	tab.tokenizedSave = true
	ui := &editorUI{
		inEditor:  true,
		status:    widget.NewLabel(""),
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		activeTab: tab,
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	ui.bindActiveTab(tab)
	if !ui.saveTokenized {
		t.Fatal("expected saveTokenized=true when active tab is tokenized")
	}
}

func TestEditorInsertMenuContainsIncludeFile(t *testing.T) {
	ui := &editorUI{}
	menu := ui.makeEditorMenu()
	if menu == nil || len(menu.Items) < 3 {
		t.Fatalf("expected editor main menu with Insert")
	}
	insert := menu.Items[2]
	if insert == nil || insert.Label != "Insert" {
		t.Fatalf("expected Insert menu at top-level index 2")
	}
	if len(insert.Items) < 1 {
		t.Fatalf("expected Insert items")
	}
	if got := insert.Items[0].Label; got != "Include File            Ctrl+K,R" {
		t.Fatalf("first Insert item = %q, want Include File item", got)
	}
}

func TestHandleEditorShortcutTriggersRuleChordToggle(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		inEditor:  true,
		resolver:  input.NewResolver(),
		status:    widget.NewLabel(""),
		activeTab: tab,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	tab.item.Content = ui.tabEditorContent(tab)

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+Q prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q,R to be handled")
	}
	if tab.ruleMode != true {
		t.Fatal("expected RULE mode enabled after Ctrl+Q,R")
	}
	if got := ui.status.Text; got != "RULE: on (ESC to exit)" {
		t.Fatalf("status = %q, want %q", got, "RULE: on (ESC to exit)")
	}

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q prefix to be handled (second time)")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q,R to be handled (second time)")
	}
	if tab.ruleMode {
		t.Fatal("expected RULE mode disabled after second Ctrl+Q,R")
	}
	if got := ui.status.Text; got != "RULE: off" {
		t.Fatalf("status = %q, want %q", got, "RULE: off")
	}
}

func TestHandleEditorShortcutTriggersCalculatorChord(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		inEditor:  true,
		resolver:  input.NewResolver(),
		status:    widget.NewLabel(""),
		activeTab: tab,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	tab.item.Content = ui.tabEditorContent(tab)

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+Q prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+Q,M to be handled")
	}
}

func TestHandleEditorShortcutTriggersIncludeFileChord(t *testing.T) {
	tab := makeSplitViewTestTab("A")
	ui := &editorUI{
		inEditor:  true,
		resolver:  input.NewResolver(),
		status:    widget.NewLabel(""),
		activeTab: tab,
		tabState:  map[*container.TabItem]*editorTab{tab.item: tab},
		entry:     tab.entry,
	}
	ui.bindTabEntry(tab)
	tab.item.Content = ui.tabEditorContent(tab)

	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+K prefix to be handled")
	}
	if !ui.resolver.HasPrefix() {
		t.Fatal("expected Ctrl+K prefix state to remain active")
	}
	if handled := ui.handleEditorShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyR, Modifier: fyne.KeyModifierControl}); !handled {
		t.Fatal("expected Ctrl+K,R to be handled")
	}
}
