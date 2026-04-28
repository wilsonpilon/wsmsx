package ui

import (
	"image/color"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const defaultSyntaxThemeID = "vscode-dark-plus"
const customSyntaxThemeID = "custom"
const defaultEditorThemeID = "dark"

const (
	editorThemeDarkID  = "dark"
	editorThemeLightID = "light"
)

const (
	colorNameSyntaxKeyword  fyne.ThemeColorName = "ws7.syntax.keyword"
	colorNameSyntaxFunction fyne.ThemeColorName = "ws7.syntax.function"
	colorNameSyntaxString   fyne.ThemeColorName = "ws7.syntax.string"
	colorNameSyntaxNumber   fyne.ThemeColorName = "ws7.syntax.number"
	colorNameSyntaxComment  fyne.ThemeColorName = "ws7.syntax.comment"
	colorNameSyntaxLiteral  fyne.ThemeColorName = "ws7.syntax.literal"
)

type syntaxThemeOption struct {
	ID    string
	Label string
}

type syntaxPalette struct {
	Keyword  color.NRGBA
	Function color.NRGBA
	String   color.NRGBA
	Number   color.NRGBA
	Comment  color.NRGBA
	Literal  color.NRGBA
}

var syntaxThemeOptions = []syntaxThemeOption{
	{ID: "vscode-dark-plus", Label: "VS Code Dark+"},
	{ID: "vscode-light-plus", Label: "VS Code Light+"},
	{ID: "sublime-monokai", Label: "Sublime Monokai"},
	{ID: "sublime-mariana", Label: "Sublime Mariana"},
	{ID: customSyntaxThemeID, Label: "Custom"},
}

var syntaxPalettes = map[string]syntaxPalette{
	"vscode-dark-plus": {
		Keyword:  color.NRGBA{R: 0x56, G: 0x9C, B: 0xD6, A: 0xFF},
		Function: color.NRGBA{R: 0xDC, G: 0xDC, B: 0xAA, A: 0xFF},
		String:   color.NRGBA{R: 0xCE, G: 0x91, B: 0x78, A: 0xFF},
		Number:   color.NRGBA{R: 0xB5, G: 0xCE, B: 0xA8, A: 0xFF},
		Comment:  color.NRGBA{R: 0x6A, G: 0x99, B: 0x55, A: 0xFF},
		Literal:  color.NRGBA{R: 0xD4, G: 0xD4, B: 0xD4, A: 0xFF},
	},
	"vscode-light-plus": {
		Keyword:  color.NRGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF},
		Function: color.NRGBA{R: 0x79, G: 0x50, B: 0xE8, A: 0xFF},
		String:   color.NRGBA{R: 0xA3, G: 0x15, B: 0x15, A: 0xFF},
		Number:   color.NRGBA{R: 0x09, G: 0x8A, B: 0x00, A: 0xFF},
		Comment:  color.NRGBA{R: 0x00, G: 0x80, B: 0x00, A: 0xFF},
		Literal:  color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF},
	},
	"sublime-monokai": {
		Keyword:  color.NRGBA{R: 0xF9, G: 0x26, B: 0x72, A: 0xFF},
		Function: color.NRGBA{R: 0xA6, G: 0xE2, B: 0x2E, A: 0xFF},
		String:   color.NRGBA{R: 0xE6, G: 0xDB, B: 0x74, A: 0xFF},
		Number:   color.NRGBA{R: 0xAE, G: 0x81, B: 0xFF, A: 0xFF},
		Comment:  color.NRGBA{R: 0x75, G: 0x71, B: 0x5E, A: 0xFF},
		Literal:  color.NRGBA{R: 0xF8, G: 0xF8, B: 0xF2, A: 0xFF},
	},
	"sublime-mariana": {
		Keyword:  color.NRGBA{R: 0xC7, G: 0x92, B: 0xEA, A: 0xFF},
		Function: color.NRGBA{R: 0x82, G: 0xAA, B: 0xFF, A: 0xFF},
		String:   color.NRGBA{R: 0xC3, G: 0xE8, B: 0x8D, A: 0xFF},
		Number:   color.NRGBA{R: 0xF7, G: 0x8C, B: 0x6C, A: 0xFF},
		Comment:  color.NRGBA{R: 0x67, G: 0x6E, B: 0x95, A: 0xFF},
		Literal:  color.NRGBA{R: 0xD8, G: 0xDE, B: 0xE9, A: 0xFF},
	},
}

func defaultCustomSyntaxPalette() syntaxPalette {
	return syntaxPalettes[defaultSyntaxThemeID]
}

func normalizeSyntaxThemeID(id string) string {
	if id == customSyntaxThemeID {
		return id
	}
	if _, ok := syntaxPalettes[id]; ok {
		return id
	}
	return defaultSyntaxThemeID
}

func syntaxThemeLabel(id string) string {
	id = normalizeSyntaxThemeID(id)
	for _, opt := range syntaxThemeOptions {
		if opt.ID == id {
			return opt.Label
		}
	}
	return "VS Code Dark+"
}

func resolveSyntaxPalette(themeID string, custom syntaxPalette) syntaxPalette {
	themeID = normalizeSyntaxThemeID(themeID)
	if themeID == customSyntaxThemeID {
		return custom
	}
	if p, ok := syntaxPalettes[themeID]; ok {
		return p
	}
	return syntaxPalettes[defaultSyntaxThemeID]
}

func normalizeEditorThemeID(id string) string {
	id = strings.TrimSpace(strings.ToLower(id))
	if id == editorThemeLightID {
		return editorThemeLightID
	}
	return editorThemeDarkID
}

type sourceCodeProTheme struct {
	fyne.Theme
	font    fyne.Resource
	palette syntaxPalette
}

func newSourceCodeProTheme(fontPath, syntaxThemeID string, customPalette syntaxPalette, editorThemeID string) (fyne.Theme, error) {
	syntaxThemeID = normalizeSyntaxThemeID(syntaxThemeID)
	editorThemeID = normalizeEditorThemeID(editorThemeID)
	palette := resolveSyntaxPalette(syntaxThemeID, customPalette)
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
	case theme.ColorNamePrimary, colorNameSyntaxKeyword:
		return t.palette.Keyword
	case theme.ColorNameSuccess, colorNameSyntaxFunction:
		return t.palette.Function
	case theme.ColorNameWarning, colorNameSyntaxString:
		return t.palette.String
	case theme.ColorNameError, colorNameSyntaxNumber:
		return t.palette.Number
	case theme.ColorNameDisabled, colorNameSyntaxComment:
		return t.palette.Comment
	case colorNameSyntaxLiteral:
		return t.palette.Literal
	}
	return t.Theme.Color(name, variant)
}

