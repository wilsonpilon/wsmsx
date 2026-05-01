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
type Resolver struct{ prefix string }

// NewResolver creates a fresh Resolver.
func NewResolver() *Resolver { return &Resolver{} }

// Resolve processes one Ctrl key (letter or symbol, single rune).
// Returns (Command, pending, error); pending=true means waiting for next key.
func (r *Resolver) Resolve(ctrlKey string) (Command, bool, error) {
	if len(ctrlKey) != 1 {
		return "", false, fmt.Errorf("ctrlKey must contain exactly 1 character")
	}

	switch r.prefix {

	case "K":
		r.prefix = ""
		switch ctrlKey {
		case "S":
			return CmdSave, false, nil
		case "T":
			return CmdSaveAs, false, nil
		case "D":
			return CmdSaveAndClose, false, nil
		case "P":
			return CmdPrint, false, nil
		case "O":
			return CmdFileCopy, false, nil
		case "J":
			return CmdFileDelete, false, nil
		case "E":
			return CmdFileRename, false, nil
		case "L":
			return CmdChangeDirectory, false, nil
		case "F":
			return CmdRunPSCommand, false, nil
		case "B":
			return CmdMarkBlockBegin, false, nil
		case "K":
			return CmdMarkBlockEnd, false, nil
		case "V":
			return CmdMoveBlock, false, nil
		case "G":
			return CmdMoveBlockOtherWin, false, nil
		case "C":
			return CmdCopyBlock, false, nil
		case "A":
			return CmdCopyBlockOtherWin, false, nil
		case "[":
			return CmdCopyFromClipboard, false, nil
		case "]":
			return CmdCopyToClipboard, false, nil
		case "W":
			return CmdCopyToFile, false, nil
		case "R":
			return CmdIncludeFile, false, nil
		case "\"":
			return CmdConvertUppercase, false, nil
		case "'":
			return CmdConvertLowercase, false, nil
		case ".":
			return CmdConvertCapitalize, false, nil
		case "Y":
			return CmdDeleteBlock, false, nil
		case "U":
			return CmdMarkPreviousBlock, false, nil
		case "N":
			return CmdColumnBlockMode, false, nil
		case "I":
			return CmdColumnReplaceMode, false, nil
		case "Q":
			r.prefix = "KQ"
			return CmdPrefixKQ, true, nil
		default:
			if ctrlKey >= "0" && ctrlKey <= "9" {
				return MarkerSetCmd(ctrlKey), false, nil
			}
			return "", false, fmt.Errorf("unsupported sequence Ctrl+K+%s", ctrlKey)
		}

	case "KQ":
		r.prefix = ""
		switch ctrlKey {
		case "X":
			return CmdExit, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+K,Q+%s", ctrlKey)
		}

	case "O":
		r.prefix = ""
		switch ctrlKey {
		case "K":
			return CmdOpenSwitch, false, nil
		case "L":
			return CmdGoDocBegin, false, nil
		case "?":
			return CmdStatus, false, nil
		case "A":
			return CmdAutoAlign, false, nil
		case "N":
			r.prefix = "ON"
			return CmdPrefixON, true, nil
		case "\r", "\n":
			return CmdCloseDialog, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+O+%s", ctrlKey)
		}

	case "ON":
		r.prefix = ""
		switch ctrlKey {
		case "D":
			return CmdEditNote, false, nil
		case "V":
			return CmdConvertNote, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+O,N+%s", ctrlKey)
		}

	case "P":
		r.prefix = ""
		switch ctrlKey {
		case "B":
			return CmdStyleBold, false, nil
		case "=":
			return CmdStyleFont, false, nil
		case "?":
			return CmdChangePrinter, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+P+%s", ctrlKey)
		}

	case "Q":
		r.prefix = ""
		switch ctrlKey {
		case "F":
			return CmdFind, false, nil
		case "A":
			return CmdFindReplace, false, nil
		case "E":
			return CmdBasicRenum, false, nil
		case "G":
			return CmdGoToChar, false, nil
		case "I":
			return CmdGoToPage, false, nil
		case "M":
			return CmdCalculator, false, nil
		case "=":
			return CmdGoToFontTag, false, nil
		case "<":
			return CmdGoToStyleTag, false, nil
		case "N":
			r.prefix = "QN"
			return CmdPrefixQN, true, nil
		case "P":
			return CmdGoPrevPosition, false, nil
		case "V":
			return CmdGoLastFindReplace, false, nil
		case "B":
			return CmdGoBlockBegin, false, nil
		case "K":
			return CmdGoBlockEnd, false, nil
		case "R":
			return CmdRule, false, nil
		case "C":
			return CmdGoDocEnd, false, nil
		case "D":
			return CmdBasicDelete, false, nil
		case "W":
			return CmdScrollContUp, false, nil
		case "Z":
			return CmdScrollContDown, false, nil
		case "Y":
			return CmdDeleteLineRight, false, nil
		case "\b", "\x7f":
			return CmdDeleteLineLeft, false, nil
		default:
			if ctrlKey >= "0" && ctrlKey <= "9" {
				return MarkerGoCmd(ctrlKey), false, nil
			}
			return "", false, fmt.Errorf("unsupported sequence Ctrl+Q+%s", ctrlKey)
		}

	case "QN":
		r.prefix = ""
		switch ctrlKey {
		case "G":
			return CmdGoToNote, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+Q,N+%s", ctrlKey)
		}

	case "M":
		r.prefix = ""
		switch ctrlKey {
		case "G":
			return CmdInsertExtendedChar, false, nil
		default:
			return "", false, fmt.Errorf("unsupported sequence Ctrl+M+%s", ctrlKey)
		}
	}

	// No prefix — single-key Ctrl commands
	switch ctrlKey {
	case "S":
		return CmdCursorLeft, false, nil
	case "D":
		return CmdCursorRight, false, nil
	case "E":
		return CmdCursorUp, false, nil
	case "X":
		return CmdCursorDown, false, nil
	case "R":
		return CmdPageUp, false, nil
	case "C":
		return CmdPageDown, false, nil
	case "W":
		return CmdClose, false, nil
	case "Y":
		return CmdDeleteLine, false, nil
	case "T":
		return CmdDeleteWord, false, nil
	case "N":
		return CmdNewTab, false, nil
	case "U":
		return CmdUndo, false, nil
	case "L":
		return CmdRepeatFind, false, nil
	case "V":
		return CmdInsertMode, false, nil
	case "K":
		r.prefix = "K"
		return CmdPrefixK, true, nil
	case "O":
		r.prefix = "O"
		return CmdPrefixO, true, nil
	case "P":
		r.prefix = "P"
		return CmdPrefixP, true, nil
	case "Q":
		r.prefix = "Q"
		return CmdPrefixQ, true, nil
	case "M":
		r.prefix = "M"
		return CmdPrefixM, true, nil
	default:
		return "", false, fmt.Errorf("unsupported Ctrl+%s", ctrlKey)
	}
}

// ClearPrefix resets any accumulated prefix state.
func (r *Resolver) ClearPrefix() { r.prefix = "" }

// HasPrefix reports whether a prefix is being accumulated.
func (r *Resolver) HasPrefix() bool { return r.prefix != "" }

// CurrentPrefix returns the raw prefix string currently accumulated (e.g. "K", "KQ", "O", "Q", "P").
func (r *Resolver) CurrentPrefix() string { return r.prefix }
