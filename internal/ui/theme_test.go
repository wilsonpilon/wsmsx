package ui

import (
	"image/color"
	"strings"
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

func TestNormalizeSyntaxThemeFallback(t *testing.T) {
	got := normalizeSyntaxThemeID("not-a-theme")
	if got != defaultSyntaxThemeID {
		t.Fatalf("unexpected syntax theme fallback: got=%q want=%q", got, defaultSyntaxThemeID)
	}
}

func TestDefaultSyntaxThemeForEditor(t *testing.T) {
	if got := defaultSyntaxThemeForEditor(editorThemeLightID); got != syntaxThemeLightID {
		t.Fatalf("unexpected light mapping: got=%q want=%q", got, syntaxThemeLightID)
	}
	if got := defaultSyntaxThemeForEditor(editorThemeDarkID); got != syntaxThemeDarkID {
		t.Fatalf("unexpected dark mapping: got=%q want=%q", got, syntaxThemeDarkID)
	}
}

func TestNormalizeSyntaxThemeRecognizesExtraPresets(t *testing.T) {
	for _, id := range []string{syntaxThemeGreenScreenID, syntaxThemeCobaltID, syntaxThemeAmberID} {
		if got := normalizeSyntaxThemeID(id); got != id {
			t.Fatalf("unexpected syntax preset normalization: got=%q want=%q", got, id)
		}
	}
}

func TestSerializeAndParseCustomSyntaxThemesRoundTrip(t *testing.T) {
	original := map[string]syntaxThemeDefinition{
		"custom-ocean": {
			ID:      "custom-ocean",
			Name:    "Custom Ocean",
			Palette: syntaxPalettes[syntaxThemeDarkID],
			Builtin: false,
		},
	}

	raw, err := serializeCustomSyntaxThemes(original)
	if err != nil {
		t.Fatalf("serialize custom syntax themes: %v", err)
	}
	parsed, err := parseCustomSyntaxThemesJSON(raw)
	if err != nil {
		t.Fatalf("parse custom syntax themes: %v", err)
	}
	got, ok := parsed["custom-ocean"]
	if !ok {
		t.Fatalf("expected custom-ocean theme, got %#v", parsed)
	}
	if got.Name != "Custom Ocean" {
		t.Fatalf("unexpected theme name: got=%q", got.Name)
	}
	if got.Palette.String != syntaxPalettes[syntaxThemeDarkID].String {
		t.Fatalf("unexpected string color: got=%v want=%v", got.Palette.String, syntaxPalettes[syntaxThemeDarkID].String)
	}
}

func TestExportAndImportSyntaxThemeJSON(t *testing.T) {
	def := syntaxThemeDefinition{
		ID:      "custom-copper",
		Name:    "Copper",
		Palette: syntaxPalettes[syntaxThemeAmberID],
		Builtin: false,
	}

	data, err := exportSyntaxThemeJSON(def)
	if err != nil {
		t.Fatalf("export syntax theme: %v", err)
	}
	imported, err := importSyntaxThemeJSON(data, map[string]syntaxThemeDefinition{})
	if err != nil {
		t.Fatalf("import syntax theme: %v", err)
	}
	if imported.Name != def.Name {
		t.Fatalf("unexpected imported name: got=%q want=%q", imported.Name, def.Name)
	}
	if imported.Palette.Number != def.Palette.Number {
		t.Fatalf("unexpected imported number color: got=%v want=%v", imported.Palette.Number, def.Palette.Number)
	}
	if imported.Builtin {
		t.Fatal("expected imported theme to be custom")
	}
}

func TestImportSyntaxThemeJSONRenamesDuplicates(t *testing.T) {
	data := []byte(`{"version":1,"name":"Cobalt","colors":{"plain":"#FFFFFF","instruction":"#9ED6FF","jump":"#FF9DDC","function":"#FFE06B","operator":"#E1F2FF","number":"#FFB06B","string":"#A5FF90","comment":"#7F9FBF","identifier":"#D7F0FF"}}`)
	imported, err := importSyntaxThemeJSON(data, map[string]syntaxThemeDefinition{})
	if err != nil {
		t.Fatalf("import duplicate name theme: %v", err)
	}
	if imported.Name == "Cobalt" {
		t.Fatalf("expected duplicate imported name to be adjusted, got=%q", imported.Name)
	}
	if !strings.HasPrefix(imported.Name, "Cobalt") {
		t.Fatalf("expected imported duplicate to keep base name, got=%q", imported.Name)
	}
}

func TestFallbackSyntaxThemeSelectionUsesExistingOrDefault(t *testing.T) {
	custom := map[string]syntaxThemeDefinition{
		"custom-ocean": {ID: "custom-ocean", Name: "Custom Ocean", Palette: syntaxPalettes[syntaxThemeDarkID]},
	}
	if got := fallbackSyntaxThemeSelection("custom-ocean", editorThemeDarkID, custom); got != "custom-ocean" {
		t.Fatalf("unexpected existing fallback selection: got=%q", got)
	}
	if got := fallbackSyntaxThemeSelection("missing", editorThemeLightID, custom); got != syntaxThemeLightID {
		t.Fatalf("unexpected default fallback selection: got=%q want=%q", got, syntaxThemeLightID)
	}
}

func TestDeleteCustomSyntaxThemeRemovesThemeAndReturnsFallback(t *testing.T) {
	custom := map[string]syntaxThemeDefinition{
		"custom-ocean": {ID: "custom-ocean", Name: "Custom Ocean", Palette: syntaxPalettes[syntaxThemeDarkID]},
	}
	nextID, err := deleteCustomSyntaxTheme("custom-ocean", editorThemeLightID, custom)
	if err != nil {
		t.Fatalf("delete custom syntax theme: %v", err)
	}
	if _, ok := custom["custom-ocean"]; ok {
		t.Fatal("expected custom theme to be removed")
	}
	if nextID != syntaxThemeLightID {
		t.Fatalf("unexpected fallback selection: got=%q want=%q", nextID, syntaxThemeLightID)
	}
}

func TestDeleteCustomSyntaxThemeRejectsBuiltin(t *testing.T) {
	custom := map[string]syntaxThemeDefinition{}
	nextID, err := deleteCustomSyntaxTheme(syntaxThemeDarkID, editorThemeDarkID, custom)
	if err == nil {
		t.Fatal("expected builtin delete to fail")
	}
	if nextID != syntaxThemeDarkID {
		t.Fatalf("unexpected selected theme after builtin delete rejection: got=%q", nextID)
	}
}

func TestResetCustomSyntaxThemeToBuiltinPreservesIdentity(t *testing.T) {
	custom := map[string]syntaxThemeDefinition{
		"custom-ocean": {ID: "custom-ocean", Name: "Custom Ocean", Palette: syntaxPalettes[syntaxThemeDarkID]},
	}
	updated, err := resetCustomSyntaxThemeToBuiltin("custom-ocean", syntaxThemeAmberID, custom)
	if err != nil {
		t.Fatalf("reset custom syntax theme: %v", err)
	}
	if updated.ID != "custom-ocean" || updated.Name != "Custom Ocean" {
		t.Fatalf("expected custom identity to be preserved, got=%#v", updated)
	}
	if updated.Palette.String != syntaxPalettes[syntaxThemeAmberID].String {
		t.Fatalf("unexpected preset-applied string color: got=%v want=%v", updated.Palette.String, syntaxPalettes[syntaxThemeAmberID].String)
	}
	if custom["custom-ocean"].Palette.Comment != syntaxPalettes[syntaxThemeAmberID].Comment {
		t.Fatalf("expected map to be updated with builtin preset palette")
	}
}

func TestResetCustomSyntaxThemeToBuiltinRejectsBuiltinTheme(t *testing.T) {
	custom := map[string]syntaxThemeDefinition{}
	if _, err := resetCustomSyntaxThemeToBuiltin(syntaxThemeDarkID, syntaxThemeAmberID, custom); err == nil {
		t.Fatal("expected builtin reset to fail")
	}
}
