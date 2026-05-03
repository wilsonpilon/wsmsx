package msxbasic

import (
	"strings"
	"unicode"

	"ws7/internal/syntax/core"
)

const dialectID = "msx-basic"

// Keywords sourced from basic-dignified/msx/badig_msx.py.
var instructionKeywords = map[string]struct{}{
	"AS": {}, "BASE": {}, "BEEP": {}, "BLOAD": {}, "BSAVE": {}, "CALL": {}, "CIRCLE": {},
	"CLEAR": {}, "CLOAD": {}, "CLOSE": {}, "CLS": {}, "CMD": {}, "COLOR": {}, "CONT": {},
	"COPY": {}, "CSAVE": {}, "CSRLIN": {}, "DEF": {}, "DEFDBL": {}, "DEFINT": {}, "MAXFILES": {},
	"DEFSNG": {}, "DEFSTR": {}, "DIM": {}, "DRAW": {}, "DSKI": {}, "END": {}, "EQV": {},
	"ERASE": {}, "ERR": {}, "ERROR": {}, "FIELD": {}, "FILES": {}, "FN": {}, "FOR": {}, "GET": {},
	"IF": {}, "INPUT": {}, "INTERVAL": {}, "IMP": {}, "IPL": {}, "KILL": {}, "LET": {},
	"LFILES": {}, "LINE": {}, "LOAD": {}, "LOCATE": {}, "LPRINT": {}, "LSET": {}, "MAX": {},
	"MERGE": {}, "MOTOR": {}, "NAME": {}, "NEW": {}, "NEXT": {}, "OFF": {}, "ON": {}, "OPEN": {},
	"OUT": {}, "OUTPUT": {}, "PAINT": {}, "POINT": {}, "POKE": {}, "PRESET": {}, "PRINT": {},
	"PSET": {}, "PUT": {}, "READ": {}, "RSET": {}, "SAVE": {}, "SCREEN": {}, "SET": {},
	"SOUND": {}, "STEP": {}, "STOP": {}, "SWAP": {}, "TIME": {}, "TO": {}, "TROFF": {},
	"TRON": {}, "USING": {}, "VPOKE": {}, "WAIT": {}, "WIDTH": {}, "?": {}, "DATA": {},
}

var functionKeywords = map[string]struct{}{
	"ATTR$": {}, "BIN$": {}, "CHR$": {}, "DSKO$": {}, "HEX$": {}, "INKEY$": {}, "INPUT$": {},
	"LEFT$": {}, "MID$": {}, "MKD$": {}, "MKI$": {}, "MKS$": {}, "OCT$": {}, "RIGHT$": {},
	"SPACE$": {}, "SPRITE$": {}, "STR$": {}, "STRING$": {},
	"ABS": {}, "ASC": {}, "ATN": {}, "CDBL": {}, "CINT": {}, "COS": {}, "CSNG": {}, "CVD": {},
	"CVI": {}, "CVS": {}, "DSKF": {}, "EOF": {}, "EXP": {}, "FIX": {}, "FPOS": {}, "FRE": {},
	"INP": {}, "INSTR": {}, "INT": {}, "KEY": {}, "LEN": {}, "LOC": {}, "LOF": {}, "LOG": {},
	"LPOS": {}, "PAD": {}, "PDL": {}, "PEEK": {}, "PLAY": {}, "POS": {}, "RND": {}, "SGN": {},
	"SIN": {}, "SPC": {}, "SPRITE": {}, "SQR": {}, "STICK": {}, "STRIG": {}, "TAB": {}, "TAN": {},
	"VAL": {}, "VARPTR": {}, "VDP": {}, "VPEEK": {},
}

var jumpKeywords = map[string]struct{}{
	"RESTORE": {}, "AUTO": {}, "RENUM": {}, "DELETE": {}, "RESUME": {}, "ERL": {}, "ELSE": {},
	"RUN": {}, "LIST": {}, "LLIST": {}, "GOTO": {}, "RETURN": {}, "THEN": {}, "GOSUB": {},
}

var wordOperators = map[string]struct{}{
	"AND": {}, "MOD": {}, "NOT": {}, "OR": {}, "XOR": {},
}

var operators = map[rune]struct{}{
	'>': {}, '=': {}, '<': {}, '+': {}, '-': {}, '*': {}, '/': {}, '^': {}, '\\': {}, ':': {}, ',': {}, ';': {}, '(': {}, ')': {}, '#': {},
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

		if ch == '?' {
			tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: "?"})
			i++
			lineStart = false
			continue
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

		if unicode.IsDigit(ch) || (ch == '&' && i+1 < len(line) && (line[i+1] == 'H' || line[i+1] == 'h' || line[i+1] == 'O' || line[i+1] == 'o' || line[i+1] == 'B' || line[i+1] == 'b')) {
			start := i
			i = consumeNumber(line, i)
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
			if _, ok := wordOperators[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenOperator, Value: word})
			} else if _, ok := jumpKeywords[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenJump, Value: word})
			} else if _, ok := instructionKeywords[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: word})
			} else if isDefUSRInstruction(upper) {
				tokens = append(tokens, core.Token{Kind: core.TokenKeyword, Value: word})
			} else if _, ok := functionKeywords[upper]; ok {
				tokens = append(tokens, core.Token{Kind: core.TokenFunction, Value: word})
			} else if isUSRFunction(upper) {
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

func consumeNumber(line string, i int) int {
	i++
	for i < len(line) {
		r := rune(line[i])
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '.' || r == '+' || r == '-' || r == '%' || r == '#' || r == '!' {
			i++
			continue
		}
		break
	}
	return i
}

func isDefUSRInstruction(word string) bool {
	if !strings.HasPrefix(word, "DEFUSR") {
		return false
	}
	for _, ch := range word[len("DEFUSR"):] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func isUSRFunction(word string) bool {
	if !strings.HasPrefix(word, "USR") {
		return false
	}
	for _, ch := range word[len("USR"):] {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
