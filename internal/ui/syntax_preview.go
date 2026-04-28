package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/syntax"
)

func syntaxTextStyle(kind syntax.TokenKind) widget.RichTextStyle {
	style := widget.RichTextStyle{
		Inline: true,
		TextStyle: fyne.TextStyle{
			Monospace: true,
		},
		ColorName: theme.ColorNameForeground,
	}

	switch kind {
	case syntax.TokenKeyword:
		style.ColorName = colorNameSyntaxKeyword
		style.TextStyle.Bold = true
	case syntax.TokenFunction:
		style.ColorName = colorNameSyntaxFunction
	case syntax.TokenComment:
		style.ColorName = colorNameSyntaxComment
		style.TextStyle.Italic = true
	case syntax.TokenString:
		style.ColorName = colorNameSyntaxString
	case syntax.TokenNumber:
		style.ColorName = colorNameSyntaxNumber
	case syntax.TokenOperator:
		style.ColorName = theme.ColorNameForeground
	case syntax.TokenIdent:
		style.ColorName = colorNameSyntaxLiteral
	default:
		style.ColorName = theme.ColorNameForeground
	}
	return style
}

func syntaxPreviewSegments(lines [][]syntax.Token) []widget.RichTextSegment {
	segments := make([]widget.RichTextSegment, 0, 64)
	if len(lines) == 0 {
		return []widget.RichTextSegment{&widget.TextSegment{Text: "", Style: syntaxTextStyle(syntax.TokenPlain)}}
	}
	for row, line := range lines {
		if len(line) == 0 {
			segments = append(segments, &widget.TextSegment{Text: "", Style: syntaxTextStyle(syntax.TokenPlain)})
		} else {
			for _, tok := range line {
				segments = append(segments, &widget.TextSegment{Text: tok.Value, Style: syntaxTextStyle(tok.Kind)})
			}
		}
		if row < len(lines)-1 {
			segments = append(segments, &widget.TextSegment{Text: "\n", Style: syntaxTextStyle(syntax.TokenPlain)})
		}
	}
	return segments
}
