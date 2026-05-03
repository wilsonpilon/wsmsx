package input

import (
	"fmt"
	"strings"
)

// Command identifies a named editor action.
type Command string

const (
	// ── Cursor / single-key Ctrl ─────────────────────────────────────────────
	CmdCursorLeft  Command = "cursor_left"  // Ctrl+S
	CmdCursorRight Command = "cursor_right" // Ctrl+D
	CmdCursorUp    Command = "cursor_up"    // Ctrl+E
	CmdCursorDown  Command = "cursor_down"  // Ctrl+X
	CmdPageUp      Command = "page_up"      // Ctrl+R
	CmdPageDown    Command = "page_down"    // Ctrl+C
	CmdNewTab      Command = "new_tab"      // Ctrl+N
	CmdScrollUp    Command = "scroll_up"    // legacy
	CmdInsertLine  Command = "insert_line"  // legacy
	CmdDeleteWord  Command = "delete_word"  // Ctrl+T
	CmdDeleteLine  Command = "delete_line"  // Ctrl+Y
	CmdUndo        Command = "undo"         // Ctrl+U
	CmdRepeatFind  Command = "repeat_find"  // Ctrl+L
	CmdInsertMode  Command = "insert_mode"  // Ctrl+V

	// ── Prefixes ─────────────────────────────────────────────────────────────
	CmdPrefixK  Command = "prefix_k"
	CmdPrefixKQ Command = "prefix_kq"
	CmdPrefixO  Command = "prefix_o"
	CmdPrefixON Command = "prefix_on"
	CmdPrefixP  Command = "prefix_p"
	CmdPrefixQ  Command = "prefix_q"
	CmdPrefixQN Command = "prefix_qn"
	CmdPrefixM  Command = "prefix_m"

	// ── File (Ctrl+K / Ctrl+O / Ctrl+P / Ctrl+K,Q) ───────────────────────────
	CmdSave            Command = "save"
	CmdSaveAs          Command = "save_as"
	CmdSaveAndClose    Command = "save_and_close"
	CmdOpenSwitch      Command = "open_switch"
	CmdClose           Command = "close"
	CmdPrint           Command = "print"
	CmdChangePrinter   Command = "change_printer"
	CmdFileCopy        Command = "file_copy"
	CmdFileDelete      Command = "file_delete"
	CmdFileRename      Command = "file_rename"
	CmdChangeDirectory Command = "change_directory"
	CmdRunPSCommand    Command = "run_ps_command"
	CmdStatus          Command = "status"
	CmdExit            Command = "exit"

	// ── Block / Edit (Ctrl+K prefix) ─────────────────────────────────────────
	CmdMarkBlockBegin    Command = "mark_block_begin"
	CmdMarkBlockEnd      Command = "mark_block_end"
	CmdMoveBlock         Command = "move_block"
	CmdMoveBlockOtherWin Command = "move_block_other_win"
	CmdCopyBlock         Command = "copy_block"
	CmdCopyBlockOtherWin Command = "copy_block_other_win"
	CmdCopyFromClipboard Command = "copy_from_clipboard"
	CmdCopyToClipboard   Command = "copy_to_clipboard"
	CmdCopyToFile        Command = "copy_to_file"
	CmdIncludeFile       Command = "include_file"
	CmdConvertUppercase  Command = "convert_uppercase"
	CmdConvertLowercase  Command = "convert_lowercase"
	CmdConvertCapitalize Command = "convert_capitalize"
	CmdDeleteBlock       Command = "delete_block"
	CmdMarkPreviousBlock Command = "mark_previous_block"
	CmdColumnBlockMode   Command = "column_block_mode"
	CmdColumnReplaceMode Command = "column_replace_mode"

	// ── Navigation / Edit (Ctrl+Q prefix) ────────────────────────────────────
	CmdFind              Command = "find"
	CmdFindReplace       Command = "find_replace"
	CmdGoToChar          Command = "go_to_char"
	CmdGoToPage          Command = "go_to_page"
	CmdGoToFontTag       Command = "go_to_font_tag"
	CmdGoToStyleTag      Command = "go_to_style_tag"
	CmdGoToNote          Command = "go_to_note"
	CmdGoPrevPosition    Command = "go_prev_position"
	CmdGoLastFindReplace Command = "go_last_find_replace"
	CmdGoBlockBegin      Command = "go_block_begin"
	CmdGoBlockEnd        Command = "go_block_end"
	CmdGoDocBegin        Command = "go_doc_begin"
	CmdGoDocEnd          Command = "go_doc_end"
	CmdScrollContUp      Command = "scroll_cont_up"
	CmdScrollContDown    Command = "scroll_cont_down"
	CmdDeleteLineLeft    Command = "delete_line_left"
	CmdDeleteLineRight   Command = "delete_line_right"
	CmdBasicDelete       Command = "basic_delete"
	CmdBasicRenum        Command = "basic_renum"

	// ── Note (Ctrl+O,N) ───────────────────────────────────────────────────────
	CmdEditNote    Command = "edit_note"
	CmdConvertNote Command = "convert_note"

	// ── Settings (Ctrl+O) ─────────────────────────────────────────────────────
	CmdAutoAlign          Command = "auto_align"
	CmdRule               Command = "rule"
	CmdCalculator         Command = "calculator"
	CmdWordCount          Command = "word_count"
	CmdStyleBold          Command = "style_bold"
	CmdStyleFont          Command = "style_font"
	CmdInsertExtendedChar Command = "insert_extended_char"
	CmdCloseDialog        Command = "close_dialog"
)

// MarkerSetCmd returns the Command for setting marker digit (0–9).
func MarkerSetCmd(digit string) Command { return Command("set_marker_" + digit) }

// MarkerGoCmd returns the Command for jumping to marker digit (0–9).
func MarkerGoCmd(digit string) Command { return Command("go_to_marker_" + digit) }

// IsSetMarker reports whether cmd is a set-marker command and returns the digit.
func IsSetMarker(cmd Command) (string, bool) {
	s := string(cmd)
	if strings.HasPrefix(s, "set_marker_") {
		return s[len("set_marker_"):], true
	}
	return "", false
}

// IsGoToMarker reports whether cmd is a go-to-marker command and returns the digit.
func IsGoToMarker(cmd Command) (string, bool) {
	s := string(cmd)
	if strings.HasPrefix(s, "go_to_marker_") {
		return s[len("go_to_marker_"):], true
	}
	return "", false
}

// ─── Resolver ────────────────────────────────────────────────────────────────

// Resolver keeps WordStar-style multi-key Ctrl chord state.
type Resolver struct {
	prefix         string
	commandByChord map[string]Command
	chordByCommand map[Command]string
}

// NewResolver creates a fresh Resolver.
func NewResolver() *Resolver {
	r := &Resolver{}
	_ = r.ApplyKeybinds(DefaultKeybindDefinitions())
	return r
}

// ApplyKeybinds updates resolver mappings using the provided command catalog.
func (r *Resolver) ApplyKeybinds(defs []KeybindDefinition) error {
	r.commandByChord = map[string]Command{}
	r.chordByCommand = map[Command]string{}
	for _, def := range defs {
		id := strings.TrimSpace(def.ID)
		shortcut := strings.TrimSpace(def.Shortcut)
		if id == "" || shortcut == "" {
			continue
		}
		if id == "set_marker_digit" || id == "go_to_marker_digit" {
			continue
		}
		if _, err := NormalizeShortcut(shortcut); err != nil {
			continue
		}
		keys, err := ShortcutToResolverChord(shortcut)
		if err != nil || len(keys) == 0 {
			continue
		}
		chord := strings.Join(keys, ",")
		cmd := Command(id)
		if existing, ok := r.commandByChord[chord]; ok && existing != cmd {
			return fmt.Errorf("shortcut conflict: %s", shortcut)
		}
		r.commandByChord[chord] = cmd
		if _, exists := r.chordByCommand[cmd]; !exists {
			r.chordByCommand[cmd] = shortcut
		}
	}
	if _, ok := r.commandByChord["K"]; ok {
		delete(r.commandByChord, "K")
	}
	if _, ok := r.commandByChord["O"]; ok {
		delete(r.commandByChord, "O")
	}
	if _, ok := r.commandByChord["P"]; ok {
		delete(r.commandByChord, "P")
	}
	if _, ok := r.commandByChord["Q"]; ok {
		delete(r.commandByChord, "Q")
	}
	if _, ok := r.commandByChord["M"]; ok {
		delete(r.commandByChord, "M")
	}
	if _, ok := r.commandByChord["K,Q"]; ok {
		delete(r.commandByChord, "K,Q")
	}
	if _, ok := r.commandByChord["O,N"]; ok {
		delete(r.commandByChord, "O,N")
	}
	if _, ok := r.commandByChord["Q,N"]; ok {
		delete(r.commandByChord, "Q,N")
	}
	return nil
}

// ShortcutForCommand returns the currently mapped shortcut label for a command.
func (r *Resolver) ShortcutForCommand(cmd Command) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.chordByCommand[cmd])
}

// Resolve processes one Ctrl key (letter or symbol, single rune).
// Returns (Command, pending, error); pending=true means waiting for next key.
func (r *Resolver) Resolve(ctrlKey string) (Command, bool, error) {
	if len(ctrlKey) != 1 {
		return "", false, fmt.Errorf("ctrlKey must contain exactly 1 character")
	}
	if r.commandByChord == nil {
		_ = r.ApplyKeybinds(DefaultKeybindDefinitions())
	}

	next := ctrlKey
	if r.prefix != "" {
		next = r.prefix + "," + ctrlKey
	}

	if cmd, ok := r.commandByChord[next]; ok {
		r.prefix = ""
		return cmd, false, nil
	}

	if r.prefix == "K" && ctrlKey >= "0" && ctrlKey <= "9" {
		r.prefix = ""
		return MarkerSetCmd(ctrlKey), false, nil
	}
	if r.prefix == "Q" && ctrlKey >= "0" && ctrlKey <= "9" {
		r.prefix = ""
		return MarkerGoCmd(ctrlKey), false, nil
	}

	if hasChordPrefix(r.commandByChord, next) {
		r.prefix = next
		return prefixCommand(next), true, nil
	}

	r.prefix = ""
	return "", false, fmt.Errorf("unsupported sequence Ctrl+%s", strings.ReplaceAll(next, ",", "+"))
}

// ClearPrefix resets any accumulated prefix state.
func (r *Resolver) ClearPrefix() { r.prefix = "" }

// HasPrefix reports whether a prefix is being accumulated.
func (r *Resolver) HasPrefix() bool { return r.prefix != "" }

// CurrentPrefix returns the raw prefix string currently accumulated (e.g. "K", "KQ", "O", "Q", "P").
func (r *Resolver) CurrentPrefix() string { return r.prefix }

func hasChordPrefix(chords map[string]Command, candidate string) bool {
	prefix := candidate + ","
	for chord := range chords {
		if strings.HasPrefix(chord, prefix) {
			return true
		}
	}
	return false
}

func prefixCommand(prefix string) Command {
	switch prefix {
	case "K":
		return CmdPrefixK
	case "K,Q":
		return CmdPrefixKQ
	case "O":
		return CmdPrefixO
	case "O,N":
		return CmdPrefixON
	case "P":
		return CmdPrefixP
	case "Q":
		return CmdPrefixQ
	case "Q,N":
		return CmdPrefixQN
	case "M":
		return CmdPrefixM
	default:
		return ""
	}
}
