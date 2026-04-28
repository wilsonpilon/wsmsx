package msxbasic

import (
	"strings"
	"unicode"

	"ws7/internal/syntax/core"
)

const dialectID = "msx-basic"

// Commands sourced from msxWrite/msx_basic_decoder.py TOKEN_MAP.
var commands = map[string]struct{}{
	"END": {}, "FOR": {}, "NEXT": {}, "DATA": {}, "INPUT": {}, "DIM": {}, "READ": {}, "LET": {}, "GOTO": {}, "RUN": {},
	"IF": {}, "RESTORE": {}, "GOSUB": {}, "RETURN": {}, "REM": {}, "STOP": {}, "PRINT": {}, "CLEAR": {}, "LIST": {},
	"NEW": {}, "ON": {}, "WAIT": {}, "DEF": {}, "POKE": {}, "CONT": {}, "CSAVE": {}, "CLOAD": {}, "OUT": {}, "LPRINT": {},
	"LLIST": {}, "CLS": {}, "WIDTH": {}, "ELSE": {}, "TRON": {}, "TROFF": {}, "SWAP": {}, "ERASE": {}, "ERROR": {},
	"RESUME": {}, "DELETE": {}, "AUTO": {}, "RENUM": {}, "DEFSTR": {}, "DEFINT": {}, "DEFSNG": {}, "DEFDBL": {},
	"LINE": {}, "OPEN": {}, "FIELD": {}, "GET": {}, "PUT": {}, "CLOSE": {}, "LOAD": {}, "MERGE": {}, "FILES": {},
	"LSET": {}, "RSET": {}, "SAVE": {}, "LFILES": {}, "CIRCLE": {}, "COLOR": {}, "DRAW": {}, "PAINT": {}, "BEEP": {},
	"PLAY": {}, "PSET": {}, "PRESET": {}, "SOUND": {}, "SCREEN": {}, "VPOKE": {}, "SPRITE": {}, "VDP": {}, "BASE": {},
	"CALL": {}, "TIME": {}, "KEY": {}, "MAX": {}, "MOTOR": {}, "BLOAD": {}, "BSAVE": {}, "DSKO$": {},
	"SET": {}, "NAME": {}, "KILL": {}, "IPL": {}, "COPY": {}, "CMD": {}, "LOCATE": {},
	"TO": {}, "THEN": {}, "TAB(": {}, "STEP": {}, "USR": {}, "FN": {}, "SPC(": {}, "NOT": {}, "ERL": {}, "ERR": {},
	"STRING$": {}, "USING": {}, "INSTR": {}, "VARPTR": {}, "CSRLIN": {}, "ATTR$": {}, "DSKI$": {}, "OFF": {},
	"INKEY$": {}, "POINT": {}, "AND": {}, "OR": {}, "XOR": {}, "EQV": {}, "IMP": {}, "MOD": {},
	"'": {},
}

// Functions sourced from msxWrite/msx_basic_decoder.py TOKEN_MAP_FF.
var functions = map[string]struct{}{
	"LEFT$": {}, "RIGHT$": {}, "MID$": {}, "SGN": {}, "INT": {}, "ABS": {}, "SQR": {}, "RND": {}, "SIN": {}, "LOG": {},
	"EXP": {}, "COS": {}, "TAN": {}, "ATN": {}, "FRE": {}, "INP": {}, "POS": {}, "LEN": {}, "STR$": {}, "VAL": {}, "ASC": {},
	"CHR$": {}, "PEEK": {}, "VPEEK": {}, "SPACES$": {}, "OCT$": {}, "HEX$": {}, "LPOS": {}, "BIN$": {}, "CINT": {},
	"CSNG": {}, "CDBL": {}, "FIX": {}, "STICK": {}, "STRIG": {}, "PDL": {}, "PAD": {}, "DSKF": {}, "FPOS": {}, "CVI": {},
	"CVS": {}, "CVD": {}, "EOF": {}, "LOC": {}, "LOF": {}, "MKI$": {}, "MK$": {}, "MKD$": {},
}

var operators = map[rune]struct{}{
	'>': {}, '=': {}, '<': {}, '+': {}, '-': {}, '*': {}, '/': {}, '^': {}, '\\': {}, ':': {}, ',': {}, ';': {}, '(': {}, ')': {},
}

type Highlighter struct{}

func NewHighlighter() *Highlighter { return &Highlighter{} }

func (h *Highlighter) ID() string   { return dialectID }
func (h *Highlighter) Name() string { return "MSX-BASIC Official" }

func (h *Highlighter) HighlightLine(line string) []core.Token {
	tokens := make([]core.Token, 0, len(line)/3+1)
	i := 0
	lineStart := true

	appendPlain := func(text string) {
		if text == "" {
			return
		}
		tokens = append(tokens, core.Token{Kind: core.TokenPlain, Value: text})
	}

	for i < len(line) {
		ch := rune(line[i])

		if ch == '\'' {
			tokens = append(tokens, core.Token{Kind: core.TokenComment, Value: line[i:]})
			break
		}

		if ch == '"' {
			start := i
			i++
			for i < len(line) {
				if line[i] == '"' {
					i++
					break
				}
				i++
			}
			tokens = append(tokens, core.Token{Kind: core.TokenString, Value: line[start:i]})
			lineStart = false
			continue
		}

		if unicode.IsSpace(ch) {
			start := i
			for i < len(line) && unicode.IsSpace(rune(line[i])) {
				i++
			}
			appendPlain(line[start:i])
			continue
		}

		if lineStart && unicode.IsDigit(ch) {
			start := i
			for i < len(line) && unicode.IsDigit(rune(line[i])) {
				i++
			}
			tokens = append(tokens, core.Token{Kind: core.TokenNumber, Value: line[start:i]})
			lineStart = false
			continue
		}

		if unicode.IsDigit(ch) || (ch == '&' && i+1 < len(line) && (line[i+1] == 'H' || line[i+1] == 'h' || line[i+1] == 'O' || line[i+1] == 'o')) {
			start := i
			i++
			for i < len(line) {
				r := rune(line[i])
				if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '.' {
					i++
					continue
				}
				break
			}
			tokens = append(tokens, core.Token{Kind: core.TokenNumber, Value: line[start:i]})
			lineStart = false
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			start := i
			i++
			for i < len(line) {
				r := rune(line[i])
				if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$' {
					i++
					continue
				}
				break
			}
			word := line[start:i]
			upper := strings.ToUpper(word)
			if upper == "REM" {
				tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: word})
				if i < len(line) {
					tokens = append(tokens, core.Token{Kind: core.TokenComment, Value: line[i:]})
				}
				break
			}
			if i < len(line) && line[i] == '(' {
				if _, ok := commands[upper+"("]; ok {
					tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: word})
					lineStart = false
					continue
				}
			}
			if _, ok := commands[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: word})
			} else if _, ok := functions[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenFunction, Value: word})
			} else {
				tokens = append(tokens, core.Token{Kind: core.TokenIdent, Value: word})
			}
			lineStart = false
			continue
		}

		if _, ok := operators[ch]; ok {
			tokens = append(tokens, core.Token{Kind: core.TokenOperator, Value: string(ch)})
			i++
			lineStart = false
			continue
		}

		appendPlain(string(ch))
		i++
		lineStart = false
	}

	if len(tokens) == 0 {
		return []core.Token{{Kind: core.TokenPlain, Value: ""}}
	}
	return tokens
}
