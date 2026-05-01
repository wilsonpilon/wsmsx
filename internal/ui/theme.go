package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const defaultEditorThemeID = "dark"

const (
	defaultEditorFontFamily = "Source Code Pro"
	defaultEditorFontWeight = "Regular"
	defaultEditorFontSize   = float32(14)
)

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
	font         fyne.Resource // regular weight
	fontBold     fyne.Resource // bold weight
	fontItalic   fyne.Resource // italic style
	fontBoldItal fyne.Resource // bold+italic style
	palette      editorPalette
	textSize     float32
}

func availableEditorFontFamilies() []string {
	families := []string{defaultEditorFontFamily, "MSX Screen 0", "MSX Screen 1"}
	sort.Strings(families)
	return families
}

func editorFontWeightsForFamily(family string) []string {
	switch normalizeEditorFontFamily(family) {
	case "MSX Screen 0", "MSX Screen 1":
		return []string{"Regular"}
	default:
		return []string{"ExtraLight", "Light", "Regular", "Medium", "SemiBold", "Bold", "ExtraBold", "Black"}
	}
}

func editorFontFamilySupportsItalic(family string) bool {
	return normalizeEditorFontFamily(family) == defaultEditorFontFamily
}

func normalizeEditorFontFamily(family string) string {
	f := strings.TrimSpace(family)
	for _, known := range availableEditorFontFamilies() {
		if strings.EqualFold(known, f) {
			return known
		}
	}
	return defaultEditorFontFamily
}

func normalizeEditorFontWeight(family, weight string) string {
	family = normalizeEditorFontFamily(family)
	w := strings.TrimSpace(weight)
	for _, known := range editorFontWeightsForFamily(family) {
		if strings.EqualFold(known, w) {
			return known
		}
	}
	return defaultEditorFontWeight
}

func normalizeEditorFontSize(size float32) float32 {
	if size < 8 || size > 48 {
		return defaultEditorFontSize
	}
	return size
}

func nextHeavierWeight(weight string) string {
	order := []string{"ExtraLight", "Light", "Regular", "Medium", "SemiBold", "Bold", "ExtraBold", "Black"}
	for i, w := range order {
		if strings.EqualFold(w, weight) {
			if i+1 < len(order) {
				return order[i+1]
			}
			return order[i]
		}
	}
	return "Bold"
}

func fontFileName(family, weight string, italic bool) string {
	family = normalizeEditorFontFamily(family)
	weight = normalizeEditorFontWeight(family, weight)
	suffix := ""
	if italic {
		suffix = "Italic"
	}
	switch family {
	case "MSX Screen 0":
		return "MSX-Screen0.ttf"
	case "MSX Screen 1":
		return "MSX-Screen1.ttf"
	default:
		return "SourceCodePro-" + weight + suffix + ".ttf"
	}
}

func loadFontResource(path string) fyne.Resource {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(filepath.Base(path), b)
}

func newConfiguredEditorTheme(resDir, editorThemeID, family, weight string, textSize float32) (fyne.Theme, error) {
	editorThemeID = normalizeEditorThemeID(editorThemeID)
	palette := editorPalettes[editorThemeID]

	base := theme.DarkTheme()
	if editorThemeID == editorThemeLightID {
		base = theme.LightTheme()
	}

	family = normalizeEditorFontFamily(family)
	weight = normalizeEditorFontWeight(family, weight)
	textSize = normalizeEditorFontSize(textSize)

	regular := loadFontResource(filepath.Join(resDir, fontFileName(family, weight, false)))
	italic := loadFontResource(filepath.Join(resDir, fontFileName(family, weight, true)))
	bold := loadFontResource(filepath.Join(resDir, fontFileName(family, nextHeavierWeight(weight), false)))
	boldItalic := loadFontResource(filepath.Join(resDir, fontFileName(family, nextHeavierWeight(weight), true)))

	if bold == nil {
		bold = regular
	}
	if italic == nil {
		italic = regular
	}
	if boldItalic == nil {
		if italic != nil {
			boldItalic = italic
		} else {
			boldItalic = bold
		}
	}

	if regular == nil && bold == nil && italic == nil && boldItalic == nil {
		return &sourceCodeProTheme{Theme: base, palette: palette, textSize: textSize}, nil
	}
	return &sourceCodeProTheme{
		Theme:        base,
		font:         regular,
		fontBold:     bold,
		fontItalic:   italic,
		fontBoldItal: boldItalic,
		palette:      palette,
		textSize:     textSize,
	}, nil
}

func newSourceCodeProTheme(fontPath, editorThemeID string) (fyne.Theme, error) {
	return newConfiguredEditorTheme(filepath.Dir(fontPath), editorThemeID, defaultEditorFontFamily, defaultEditorFontWeight, defaultEditorFontSize)
}

func (t *sourceCodeProTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold && style.Italic && t.fontBoldItal != nil {
		return t.fontBoldItal
	}
	if style.Bold && t.fontBold != nil {
		return t.fontBold
	}
	if style.Italic && t.fontItalic != nil {
		return t.fontItalic
	}
	if t.font != nil {
		return t.font
	}
	return t.Theme.Font(style)
}

func (t *sourceCodeProTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText && t.textSize > 0 {
		return t.textSize
	}
	return t.Theme.Size(name)
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
