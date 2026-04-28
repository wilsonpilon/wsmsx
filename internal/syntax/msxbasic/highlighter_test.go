package msxbasic

import (
	"testing"

	"ws7/internal/syntax/core"
)

func TestHighlightLineKeywordsNumbersAndString(t *testing.T) {
	h := NewHighlighter()
	line := "10 IF A=1 THEN PRINT \"OK\""
	tokens := h.HighlightLine(line)

	assertHasToken(t, tokens, core.TokenNumber, "10")
	assertHasToken(t, tokens, core.TokenKeyword, "IF")
	assertHasToken(t, tokens, core.TokenKeyword, "THEN")
	assertHasToken(t, tokens, core.TokenKeyword, "PRINT")
	assertHasToken(t, tokens, core.TokenString, "\"OK\"")
}

func TestHighlightLineRemComment(t *testing.T) {
	h := NewHighlighter()
	tokens := h.HighlightLine("100 REM comentario")
	assertHasToken(t, tokens, core.TokenKeyword, "REM")
	assertHasToken(t, tokens, core.TokenComment, " comentario")
}

func TestHighlightLineSingleQuoteComment(t *testing.T) {
	h := NewHighlighter()
	tokens := h.HighlightLine("PRINT A ' teste")
	assertHasToken(t, tokens, core.TokenKeyword, "PRINT")
	assertHasToken(t, tokens, core.TokenComment, "' teste")
}

func TestHighlightLineFunctionToken(t *testing.T) {
	h := NewHighlighter()
	tokens := h.HighlightLine("A$=LEFT$(B$):PRINT A$")
	assertHasToken(t, tokens, core.TokenIdent, "A$")
	assertHasToken(t, tokens, core.TokenFunction, "LEFT$")
	assertHasToken(t, tokens, core.TokenIdent, "B$")
	assertHasToken(t, tokens, core.TokenKeyword, "PRINT")
}

func TestHighlightLineCommandWithParenStaysKeyword(t *testing.T) {
	h := NewHighlighter()
	tokens := h.HighlightLine("X=TAB(10):Y=SPC(2)")
	assertHasToken(t, tokens, core.TokenKeyword, "TAB")
	assertHasToken(t, tokens, core.TokenKeyword, "SPC")
}

func TestHighlightLineNumericLiteralFormats(t *testing.T) {
	h := NewHighlighter()
	tokens := h.HighlightLine("10 A=&HFF:B=&O377:C=42")
	assertHasToken(t, tokens, core.TokenNumber, "10")
	assertHasToken(t, tokens, core.TokenNumber, "&HFF")
	assertHasToken(t, tokens, core.TokenNumber, "&O377")
	assertHasToken(t, tokens, core.TokenNumber, "42")
}

func assertHasToken(t *testing.T, tokens []core.Token, kind core.TokenKind, value string) {
	t.Helper()
	for _, tok := range tokens {
		if tok.Kind == kind && tok.Value == value {
			return
		}
	}
	t.Fatalf("expected token kind=%q value=%q, got=%v", kind, value, tokens)
}
