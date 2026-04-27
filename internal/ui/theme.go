package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type sourceCodeProTheme struct {
	fyne.Theme
	font fyne.Resource
}

func newSourceCodeProTheme(fontPath string) (fyne.Theme, error) {
	bytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}
	base := theme.DefaultTheme()
	return &sourceCodeProTheme{
		Theme: base,
		font:  fyne.NewStaticResource("SourceCodePro-Bold.ttf", bytes),
	}, nil
}

func (t *sourceCodeProTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.font
}

