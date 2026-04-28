package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func toNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func TestSourceCodeProThemeSyntaxColorFallback(t *testing.T) {
	th, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", "vscode-dark-plus", defaultCustomSyntaxPalette(), editorThemeDarkID)
	if err != nil {
		t.Fatalf("expected fallback theme without error, got: %v", err)
	}

	cases := []struct {
		name string
		want color.NRGBA
	}{
		{name: string(theme.ColorNamePrimary), want: color.NRGBA{R: 0x56, G: 0x9C, B: 0xD6, A: 0xFF}},
		{name: string(theme.ColorNameSuccess), want: color.NRGBA{R: 0xDC, G: 0xDC, B: 0xAA, A: 0xFF}},
		{name: string(theme.ColorNameWarning), want: color.NRGBA{R: 0xCE, G: 0x91, B: 0x78, A: 0xFF}},
		{name: string(theme.ColorNameError), want: color.NRGBA{R: 0xB5, G: 0xCE, B: 0xA8, A: 0xFF}},
		{name: string(theme.ColorNameDisabled), want: color.NRGBA{R: 0x6A, G: 0x99, B: 0x55, A: 0xFF}},
		{name: string(colorNameSyntaxLiteral), want: color.NRGBA{R: 0xD4, G: 0xD4, B: 0xD4, A: 0xFF}},
	}

	for _, tc := range cases {
		got := toNRGBA(th.Color(fyne.ThemeColorName(tc.name), theme.VariantDark))
		if got != tc.want {
			t.Fatalf("unexpected fallback color for %s: got=%v want=%v", tc.name, got, tc.want)
		}
	}
}

func TestSourceCodeProThemeMonokaiPalette(t *testing.T) {
	th, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", "sublime-monokai", defaultCustomSyntaxPalette(), editorThemeDarkID)
	if err != nil {
		t.Fatalf("expected fallback theme without error, got: %v", err)
	}

	got := toNRGBA(th.Color(theme.ColorNamePrimary, theme.VariantDark))
	want := color.NRGBA{R: 0xF9, G: 0x26, B: 0x72, A: 0xFF}
	if got != want {
		t.Fatalf("unexpected monokai keyword color: got=%v want=%v", got, want)
	}
}

func TestSourceCodeProThemeCustomPalette(t *testing.T) {
	custom := syntaxPalette{
		Keyword:  color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xFF},
		Function: color.NRGBA{R: 0x22, G: 0x33, B: 0x44, A: 0xFF},
		String:   color.NRGBA{R: 0x33, G: 0x44, B: 0x55, A: 0xFF},
		Number:   color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xFF},
		Comment:  color.NRGBA{R: 0x55, G: 0x66, B: 0x77, A: 0xFF},
		Literal:  color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF},
	}

	th, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", customSyntaxThemeID, custom, editorThemeDarkID)
	if err != nil {
		t.Fatalf("expected fallback theme without error, got: %v", err)
	}

	got := toNRGBA(th.Color(theme.ColorNamePrimary, theme.VariantDark))
	if got != custom.Keyword {
		t.Fatalf("unexpected custom keyword color: got=%v want=%v", got, custom.Keyword)
	}
	gotLiteral := toNRGBA(th.Color(colorNameSyntaxLiteral, theme.VariantDark))
	if gotLiteral != custom.Literal {
		t.Fatalf("unexpected custom literal color: got=%v want=%v", gotLiteral, custom.Literal)
	}
}

