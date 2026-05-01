package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/theme"
)

func toNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}

func TestSourceCodeProThemeDarkFallback(t *testing.T) {
	th, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", editorThemeDarkID)
	if err != nil {
		t.Fatalf("expected fallback theme without error, got: %v", err)
	}

	gotPrimary := toNRGBA(th.Color(theme.ColorNamePrimary, theme.VariantDark))
	wantPrimary := color.NRGBA{R: 0x00, G: 0x7A, B: 0xCC, A: 0xFF}
	if gotPrimary != wantPrimary {
		t.Fatalf("unexpected dark primary color: got=%v want=%v", gotPrimary, wantPrimary)
	}
}

func TestSourceCodeProThemeMonokaiPrimary(t *testing.T) {
	th, err := newSourceCodeProTheme("Z:\\not-found-font.ttf", editorThemeMonokaiID)
	if err != nil {
		t.Fatalf("expected fallback theme without error, got: %v", err)
	}

	got := toNRGBA(th.Color(theme.ColorNamePrimary, theme.VariantDark))
	want := color.NRGBA{R: 0xA6, G: 0xE2, B: 0x2E, A: 0xFF}
	if got != want {
		t.Fatalf("unexpected monokai primary color: got=%v want=%v", got, want)
	}
}

func TestNormalizeEditorFontFamilyFallback(t *testing.T) {
	got := normalizeEditorFontFamily("not-a-real-family")
	if got != defaultEditorFontFamily {
		t.Fatalf("unexpected family fallback: got=%q want=%q", got, defaultEditorFontFamily)
	}
}

func TestNormalizeEditorFontWeightByFamily(t *testing.T) {
	if got := normalizeEditorFontWeight("MSX Screen 0", "Black"); got != "Regular" {
		t.Fatalf("unexpected weight for MSX Screen 0: got=%q want=%q", got, "Regular")
	}
	if got := normalizeEditorFontWeight(defaultEditorFontFamily, "Black"); got != "Black" {
		t.Fatalf("unexpected weight for Source Code Pro: got=%q want=%q", got, "Black")
	}
}
