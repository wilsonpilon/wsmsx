package ui

import (
	"testing"

	"fyne.io/fyne/v2/theme"

	"ws7/internal/syntax"
)

func TestSyntaxTextStyle(t *testing.T) {
	kw := syntaxTextStyle(syntax.TokenKeyword)
	if kw.ColorName != colorNameSyntaxKeyword || !kw.TextStyle.Bold {
		t.Fatalf("expected keyword style primary+bold, got color=%q bold=%v", kw.ColorName, kw.TextStyle.Bold)
	}

	comment := syntaxTextStyle(syntax.TokenComment)
	if comment.ColorName != colorNameSyntaxComment || !comment.TextStyle.Italic {
		t.Fatalf("expected comment style disabled+italic, got color=%q italic=%v", comment.ColorName, comment.TextStyle.Italic)
	}

	plain := syntaxTextStyle(syntax.TokenPlain)
	if plain.ColorName != theme.ColorNameForeground {
		t.Fatalf("expected plain style foreground, got color=%q", plain.ColorName)
	}

	literal := syntaxTextStyle(syntax.TokenIdent)
	if literal.ColorName != colorNameSyntaxLiteral {
		t.Fatalf("expected literal style dedicated literal color, got color=%q", literal.ColorName)
	}
}

func TestSyntaxPreviewSegments(t *testing.T) {
	lines := [][]syntax.Token{
		{
			{Kind: syntax.TokenNumber, Value: "10"},
			{Kind: syntax.TokenPlain, Value: " "},
			{Kind: syntax.TokenKeyword, Value: "PRINT"},
		},
		{
			{Kind: syntax.TokenComment, Value: "' test"},
		},
	}

	segments := syntaxPreviewSegments(lines)
	if len(segments) != 5 {
		t.Fatalf("expected 5 segments (3 + newline + 1), got %d", len(segments))
	}
}
