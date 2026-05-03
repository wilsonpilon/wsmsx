package input

// KeybindDefinition describes a command exposed in the keyboard reference.
type KeybindDefinition struct {
	ID           string
	Label        string
	Shortcut     string
	Context      string
	Implemented  bool
	Configurable bool
}

const (
	KeybindContextEditor  = "editor"
	KeybindContextGlobal  = "global"
	KeybindContextOpening = "opening"
)

var defaultKeybindDefinitions = []KeybindDefinition{
	{ID: string(CmdCursorLeft), Label: "Cursor Left", Shortcut: "Ctrl+S", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCursorRight), Label: "Cursor Right", Shortcut: "Ctrl+D", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCursorUp), Label: "Cursor Up", Shortcut: "Ctrl+E", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCursorDown), Label: "Cursor Down", Shortcut: "Ctrl+X", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdPageUp), Label: "Page Up", Shortcut: "Ctrl+R", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdPageDown), Label: "Page Down", Shortcut: "Ctrl+C", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdNewTab), Label: "New File", Shortcut: "Ctrl+N", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdClose), Label: "Close", Shortcut: "Ctrl+W", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdDeleteLine), Label: "Delete Line", Shortcut: "Ctrl+Y", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdDeleteWord), Label: "Delete Word", Shortcut: "Ctrl+T", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdUndo), Label: "Undo", Shortcut: "Ctrl+U", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdRepeatFind), Label: "Repeat Find", Shortcut: "Ctrl+L", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdInsertMode), Label: "Insert/Overtype Mode", Shortcut: "Ctrl+V", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdSave), Label: "Save", Shortcut: "Ctrl+K,S", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdSaveAs), Label: "Save As", Shortcut: "Ctrl+K,T", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdSaveAndClose), Label: "Save and Close", Shortcut: "Ctrl+K,D", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdOpenSwitch), Label: "Open/Switch", Shortcut: "Ctrl+O,K", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdPrint), Label: "Print", Shortcut: "Ctrl+K,P", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdChangePrinter), Label: "Change Printer", Shortcut: "Ctrl+P,?", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdFileCopy), Label: "Copy File", Shortcut: "Ctrl+K,O", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdFileDelete), Label: "Delete File", Shortcut: "Ctrl+K,J", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdFileRename), Label: "Rename File", Shortcut: "Ctrl+K,E", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdChangeDirectory), Label: "Change Drive/Directory", Shortcut: "Ctrl+K,L", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdRunPSCommand), Label: "Run PS Command", Shortcut: "Ctrl+K,F", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdStatus), Label: "Status", Shortcut: "Ctrl+O,?", Context: KeybindContextGlobal, Implemented: true, Configurable: true},
	{ID: string(CmdExit), Label: "Exit", Shortcut: "Ctrl+K,Q,X", Context: KeybindContextGlobal, Implemented: true, Configurable: true},
	{ID: string(CmdMarkBlockBegin), Label: "Mark Block Begin", Shortcut: "Ctrl+K,B", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdMarkBlockEnd), Label: "Mark Block End", Shortcut: "Ctrl+K,K", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdMoveBlock), Label: "Move Block", Shortcut: "Ctrl+K,V", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdMoveBlockOtherWin), Label: "Move Block from Other Window", Shortcut: "Ctrl+K,G", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdCopyBlock), Label: "Copy Block", Shortcut: "Ctrl+K,C", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCopyBlockOtherWin), Label: "Copy Block from Other Window", Shortcut: "Ctrl+K,A", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdCopyFromClipboard), Label: "Paste from Clipboard", Shortcut: "Ctrl+K,[", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCopyToClipboard), Label: "Copy to Clipboard", Shortcut: "Ctrl+K,]", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCopyToFile), Label: "Copy to Another File", Shortcut: "Ctrl+K,W", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdIncludeFile), Label: "Include File", Shortcut: "Ctrl+K,R", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdConvertUppercase), Label: "Convert Uppercase", Shortcut: "Ctrl+K,\"", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdConvertLowercase), Label: "Convert Lowercase", Shortcut: "Ctrl+K,'", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdConvertCapitalize), Label: "Convert Capitalize", Shortcut: "Ctrl+K,.", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdDeleteBlock), Label: "Delete Block", Shortcut: "Ctrl+K,Y", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdMarkPreviousBlock), Label: "Mark Previous Block", Shortcut: "Ctrl+K,U", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdColumnBlockMode), Label: "Column Block Mode", Shortcut: "Ctrl+K,N", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdColumnReplaceMode), Label: "Column Replace Mode", Shortcut: "Ctrl+K,I", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdFind), Label: "Find", Shortcut: "Ctrl+Q,F", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdFindReplace), Label: "Find and Replace", Shortcut: "Ctrl+Q,A", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdGoToChar), Label: "Go to Character", Shortcut: "Ctrl+Q,G", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdGoToPage), Label: "Go to Page", Shortcut: "Ctrl+Q,I", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdGoToFontTag), Label: "Go to Font Tag", Shortcut: "Ctrl+Q,=", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoToStyleTag), Label: "Go to Style Tag", Shortcut: "Ctrl+Q,<", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoToNote), Label: "Go to Note", Shortcut: "Ctrl+Q,N,G", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoPrevPosition), Label: "Go to Previous Position", Shortcut: "Ctrl+Q,P", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoLastFindReplace), Label: "Go to Last Find/Replace", Shortcut: "Ctrl+Q,V", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoBlockBegin), Label: "Go to Beginning of Block", Shortcut: "Ctrl+Q,B", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoBlockEnd), Label: "Go to End of Block", Shortcut: "Ctrl+Q,K", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdGoDocBegin), Label: "Go to Document Beginning", Shortcut: "Ctrl+O,L", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdGoDocEnd), Label: "Go to Document End", Shortcut: "Ctrl+Q,C", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdScrollContUp), Label: "Scroll Continuously Up", Shortcut: "Ctrl+Q,W", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdScrollContDown), Label: "Scroll Continuously Down", Shortcut: "Ctrl+Q,Z", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdDeleteLineLeft), Label: "Delete Left of Cursor", Shortcut: "Ctrl+Q,DEL", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdDeleteLineRight), Label: "Delete Right of Cursor", Shortcut: "Ctrl+Q,Y", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdBasicDelete), Label: "BASIC DELETE", Shortcut: "Ctrl+Q,D", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdBasicRenum), Label: "BASIC RENUM", Shortcut: "Ctrl+Q,E", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdEditNote), Label: "Edit Note", Shortcut: "Ctrl+O,N,D", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdConvertNote), Label: "Convert Note", Shortcut: "Ctrl+O,N,V", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdAutoAlign), Label: "Auto Align", Shortcut: "Ctrl+O,A", Context: KeybindContextEditor, Implemented: false, Configurable: true},
	{ID: string(CmdRule), Label: "Rule", Shortcut: "Ctrl+Q,R", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCalculator), Label: "Calculator", Shortcut: "Ctrl+Q,M", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdWordCount), Label: "Word Count", Shortcut: "", Context: KeybindContextEditor, Implemented: true, Configurable: false},
	{ID: string(CmdStyleBold), Label: "Style Bold", Shortcut: "Ctrl+P,B", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdStyleFont), Label: "Style Font", Shortcut: "Ctrl+P,=", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdInsertExtendedChar), Label: "Insert Extended Character", Shortcut: "Ctrl+M,G", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdCloseDialog), Label: "Close Dialog", Shortcut: "Ctrl+O,Enter", Context: KeybindContextGlobal, Implemented: true, Configurable: false},
	{ID: "set_marker_digit", Label: "Set Marker (0-9)", Shortcut: "Ctrl+K,[0-9]", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: "go_to_marker_digit", Label: "Go to Marker (0-9)", Shortcut: "Ctrl+Q,[0-9]", Context: KeybindContextEditor, Implemented: true, Configurable: true},
	{ID: string(CmdScrollUp), Label: "Scroll Up (Legacy)", Shortcut: "", Context: KeybindContextEditor, Implemented: false, Configurable: false},
	{ID: string(CmdInsertLine), Label: "Insert Line (Legacy)", Shortcut: "", Context: KeybindContextEditor, Implemented: true, Configurable: false},
}

// DefaultKeybindDefinitions returns the built-in command/keybind catalog.
func DefaultKeybindDefinitions() []KeybindDefinition {
	defs := make([]KeybindDefinition, len(defaultKeybindDefinitions))
	copy(defs, defaultKeybindDefinitions)
	return defs
}
