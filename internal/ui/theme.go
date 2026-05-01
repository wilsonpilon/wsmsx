package ui

import (
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const defaultEditorThemeID = "dark"

const (
	editorThemeDarkID      = "dark"
	editorThemeLightID     = "light"
	editorThemeOneDarkID   = "one-dark"
	editorThemeMonokaiID   = "monokai"
	editorThemeSolarizedID = "solarized-dark"
	editorThemeGithubID    = "github-dark"
)

type editorPalette struct {
	Background      color.NRGBA
	Foreground      color.NRGBA
	MenuBackground  color.NRGBA
	InputBackground color.NRGBA
	Overlay         color.NRGBA
	Primary         color.NRGBA
}

var editorPalettes = map[string]editorPalette{
	editorThemeDarkID: {
		Background:      color.NRGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x12, G: 0x12, B: 0x12, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3D, G: 0x3D, B: 0x3D, A: 0xFF},
		Primary:         color.NRGBA{R: 0x00, G: 0x7A, B: 0xCC, A: 0xFF},
	},
	editorThemeLightID: {
		Background:      color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		Foreground:      color.NRGBA{R: 0x20, G: 0x20, B: 0x20, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0xF3, G: 0xF3, B: 0xF3, A: 0xFF},
		InputBackground: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
		Overlay:         color.NRGBA{R: 0xE5, G: 0xE5, B: 0xE5, A: 0xFF},
		Primary:         color.NRGBA{R: 0x00, G: 0x54, B: 0x99, A: 0xFF},
	},
	editorThemeOneDarkID: {
		Background:      color.NRGBA{R: 0x28, G: 0x2C, B: 0x34, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xAB, G: 0xB2, B: 0xBF, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x21, G: 0x25, B: 0x2B, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x1E, G: 0x22, B: 0x27, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3B, G: 0x40, B: 0x48, A: 0xFF},
		Primary:         color.NRGBA{R: 0x61, G: 0xAF, B: 0xEF, A: 0xFF},
	},
	editorThemeMonokaiID: {
		Background:      color.NRGBA{R: 0x27, G: 0x28, B: 0x22, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xF8, G: 0xF8, B: 0xF2, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x19, G: 0x19, B: 0x19, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x3E, G: 0x3D, B: 0x32, A: 0xFF},
		Primary:         color.NRGBA{R: 0xA6, G: 0xE2, B: 0x2E, A: 0xFF},
	},
	editorThemeSolarizedID: {
		Background:      color.NRGBA{R: 0x00, G: 0x2B, B: 0x36, A: 0xFF},
		Foreground:      color.NRGBA{R: 0x83, G: 0x94, B: 0x96, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x07, G: 0x36, B: 0x42, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x00, G: 0x1E, B: 0x26, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x58, G: 0x6E, B: 0x75, A: 0xFF},
		Primary:         color.NRGBA{R: 0x26, G: 0x8B, B: 0xD2, A: 0xFF},
	},
	editorThemeGithubID: {
		Background:      color.NRGBA{R: 0x0D, G: 0x11, B: 0x17, A: 0xFF},
		Foreground:      color.NRGBA{R: 0xC9, G: 0xD1, B: 0xD9, A: 0xFF},
		MenuBackground:  color.NRGBA{R: 0x16, G: 0x1B, B: 0x22, A: 0xFF},
		InputBackground: color.NRGBA{R: 0x01, G: 0x04, B: 0x09, A: 0xFF},
		Overlay:         color.NRGBA{R: 0x30, G: 0x36, B: 0x3D, A: 0xFF},
		Primary:         color.NRGBA{R: 0x58, G: 0xA6, B: 0xFF, A: 0xFF},
	},
}

func normalizeEditorThemeID(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	if _, ok := editorPalettes[id]; ok {
		return id
	}
	return editorThemeDarkID
}

type sourceCodeProTheme struct {
	fyne.Theme
	font    fyne.Resource
	palette editorPalette
}

func newSourceCodeProTheme(fontPath, editorThemeID string) (fyne.Theme, error) {
	editorThemeID = normalizeEditorThemeID(editorThemeID)
	palette := editorPalettes[editorThemeID]

	base := theme.DarkTheme()
	if editorThemeID == editorThemeLightID {
		base = theme.LightTheme()
	}

	bytes, err := os.ReadFile(fontPath)
	if err != nil {
		return &sourceCodeProTheme{Theme: base, palette: palette}, nil
	}
	return &sourceCodeProTheme{
		Theme:   base,
		font:    fyne.NewStaticResource("SourceCodePro-Bold.ttf", bytes),
		palette: palette,
	}, nil
}

func (t *sourceCodeProTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.font == nil {
		return t.Theme.Font(style)
	}
	return t.font
}

func (t *sourceCodeProTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.palette.Background
	case theme.ColorNameInputBackground:
		return t.palette.InputBackground
	case theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground:
		return t.palette.MenuBackground
	case theme.ColorNameForeground:
		return t.palette.Foreground
	case theme.ColorNamePrimary:
		return t.palette.Primary
	case theme.ColorNameSelection, theme.ColorNameFocus, theme.ColorNameHover:
		c := t.palette.Primary
		return color.NRGBA{R: c.R, G: c.G, B: c.B, A: 0x44}
	}
	return t.Theme.Color(name, variant)
}
