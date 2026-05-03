package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// transparentEntryTextTheme hides entry glyphs while keeping caret/selection behavior.
type transparentEntryTextTheme struct{}

func (transparentEntryTextTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground, theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	default:
		return fyne.CurrentApp().Settings().Theme().Color(name, variant)
	}
}

func (transparentEntryTextTheme) Font(style fyne.TextStyle) fyne.Resource {
	return fyne.CurrentApp().Settings().Theme().Font(style)
}

func (transparentEntryTextTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return fyne.CurrentApp().Settings().Theme().Icon(name)
}

func (transparentEntryTextTheme) Size(name fyne.ThemeSizeName) float32 {
	return fyne.CurrentApp().Settings().Theme().Size(name)
}
