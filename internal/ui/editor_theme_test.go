package ui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/syntax"
)

func findSegmentByColorName(segments []widget.RichTextSegment, colorName fyne.ThemeColorName) *widget.TextSegment {
	for _, seg := range segments {
		textSeg, ok := seg.(*widget.TextSegment)
		if !ok {
			continue
		}
		if textSeg.Style.ColorName == colorName {
			return textSeg
		}
	}
	return nil
}

func TestApplyCurrentSyntaxThemeRebuildsSyntaxEntrySegments(t *testing.T) {
	a := test.NewApp()
	a.Settings().SetTheme(theme.DefaultTheme())
	t.Cleanup(func() { a.Quit() })

	w := a.NewWindow("theme-test")
	t.Cleanup(w.Close)

	entry := newSyntaxHighlightEntry(syntax.DialectMSXBasicOfficial)
	entry.SetText("10 PRINT \"HELLO\"\n20 REM test comment")

	tab := &editorTab{
		entry:         entry.entry,
		syntaxEntry:   entry,
		syntaxDialect: syntax.DialectMSXBasicOfficial,
	}

	ui := &editorUI{
		fyneApp:           a,
		window:            w,
		tabState:          map[*container.TabItem]*editorTab{&container.TabItem{}: tab},
		syntaxThemeID:     "vscode-dark-plus",
		customSyntaxPalette: defaultCustomSyntaxPalette(),
	}

	kwBefore := findSegmentByColorName(entry.richText.Segments, colorNameSyntaxKeyword)
	if kwBefore == nil {
		t.Fatalf("expected keyword segment before theme switch")
	}
	commentBefore := findSegmentByColorName(entry.richText.Segments, colorNameSyntaxComment)
	if commentBefore == nil {
		t.Fatalf("expected comment segment before theme switch")
	}
	beforeSegPtr := entry.richText.Segments[0]

	ui.syntaxThemeID = "sublime-monokai"
	ui.applyCurrentSyntaxTheme()

	kwAfter := findSegmentByColorName(entry.richText.Segments, colorNameSyntaxKeyword)
	if kwAfter == nil {
		t.Fatalf("expected keyword segment after theme switch")
	}
	commentAfter := findSegmentByColorName(entry.richText.Segments, colorNameSyntaxComment)
	if commentAfter == nil {
		t.Fatalf("expected comment segment after theme switch")
	}
	if beforeSegPtr == entry.richText.Segments[0] {
		t.Fatalf("expected syntax segments to be rebuilt on theme switch")
	}
	if kwAfter.Style.ColorName != colorNameSyntaxKeyword {
		t.Fatalf("expected keyword segment color name to stay mapped to syntax keyword, got %q", kwAfter.Style.ColorName)
	}
	if commentAfter.Style.ColorName != colorNameSyntaxComment {
		t.Fatalf("expected comment segment color name to stay mapped to syntax comment, got %q", commentAfter.Style.ColorName)
	}

	thMonokai, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", "sublime-monokai", defaultCustomSyntaxPalette(), editorThemeDarkID)
	if err != nil {
		t.Fatalf("failed to build monokai theme for assertion: %v", err)
	}

	// Keyword color
	wantKeywordColor := toNRGBA(thMonokai.Color(colorNameSyntaxKeyword, theme.VariantDark))
	gotKeywordColor := toNRGBA(a.Settings().Theme().Color(colorNameSyntaxKeyword, theme.VariantDark))
	if gotKeywordColor != wantKeywordColor {
		t.Fatalf("expected app keyword color to follow selected theme: got=%v want=%v", gotKeywordColor, wantKeywordColor)
	}

	// Comment color
	wantCommentColor := toNRGBA(thMonokai.Color(colorNameSyntaxComment, theme.VariantDark))
	gotCommentColor := toNRGBA(a.Settings().Theme().Color(colorNameSyntaxComment, theme.VariantDark))
	if gotCommentColor != wantCommentColor {
		t.Fatalf("expected app comment color to follow selected theme: got=%v want=%v", gotCommentColor, wantCommentColor)
	}
}

