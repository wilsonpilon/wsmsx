package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// MSXInternationalDecodeTable is the 256-entry character map for MSX International.
// Sourced from msx-encoding/src/charsets/international.ts and common.ts
var MSXInternationalDecodeTable = [256]string{
	// 0x00-0x0F
	"\000", "\u263A", "\u263B", "\u2665", "\u2666", "\u2663", "\u2660", "\u2022",
	"\u25D8", "\u25CB", "\u25D9", "\u2642", "\u2640", "\u266A", "\u266B", "\u263C",
	// 0x10-0x1F
	"\u25BA", "\u2534", "\u252C", "\u2524", "\u251C", "\u253C", "\u2502", "\u2500",
	"\u250C", "\u2510", "\u2514", "\u2518", "\u2573", "\u2571", "\u2572", "\u1FBAF",
	// 0x20-0x2F
	" ", "!", "\"", "#", "$", "%", "&", "'", "(", ")", "*", "+", ",", "-", ".", "/",
	// 0x30-0x3F
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ":", ";", "<", "=", ">", "?",
	// 0x40-0x4F
	"@", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
	// 0x50-0x5F
	"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "[", "\\", "]", "^", "_",
	// 0x60-0x6F
	"`", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o",
	// 0x70-0x7F
	"p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "{", "|", "}", "~", "\u2302",
	// 0x80-0x8F
	"\u00C7", "\u00FC", "\u00E9", "\u00E2", "\u00E4", "\u00E0", "\u00E5", "\u00E7",
	"\u00EA", "\u00EB", "\u00E8", "\u00EF", "\u00EE", "\u00EC", "\u00C4", "\u00C5",
	// 0x90-0x9F
	"\u00C9", "\u00E6", "\u00C6", "\u00F4", "\u00F6", "\u00F2", "\u00FB", "\u00F9",
	"\u00FF", "\u00D6", "\u00DC", "\u00A2", "\u00A3", "\u00A5", "\u20A7", "\u0192",
	// 0xA0-0xAF
	"\u00E1", "\u00ED", "\u00F3", "\u00FA", "\u00F1", "\u00D1", "\u00AA", "\u00BA",
	"\u00BF", "\u2310", "\u00AC", "\u00BD", "\u00BC", "\u00A1", "\u00AB", "\u00BB",
	// 0xB0-0xBF
	"\u00C3", "\u00E3", "\u0128", "\u0129", "\u00D5", "\u00F5", "\u0168", "\u0169",
	"\u0132", "\u0133", "\u00BE", "\u223D", "\u25C7", "\u2030", "\u00B6", "\u00A7",
	// 0xC0-0xCF
	"\u2582", "\u259A", "\u2586", "\u1FB82", "\u25AC", "\u1FB85", "\u258E", "\u259E",
	"\u258A", "\u1FB87", "\u1FB8A", "\u1FB99", "\u1FB98", "\u1FB6D", "\u1FB6F", "\u1FB6C",
	// 0xD0-0xDF
	"\u1FB6E", "\u1FB9A", "\u1FB9B", "\u2598", "\u2597", "\u259D", "\u2596", "\u1FB96",
	"\u0394", "\u2021", "\u03C9", "\u2588", "\u2584", "\u258C", "\u2590", "\u2580",
	// 0xE0-0xEF
	"\u03B1", "\u00DF", "\u0393", "\u03C0", "\u03A3", "\u03C3", "\u00B5", "\u03C4",
	"\u03A6", "\u0398", "\u03A9", "\u03B4", "\u221E", "\u2205", "\u2208", "\u2229",
	// 0xF0-0xFF
	"\u2261", "\u00B1", "\u2265", "\u2264", "\u2320", "\u2321", "\u00F7", "\u2248",
	"\u00B0", "\u2219", "\u00B7", "\u221A", "\u207F", "\u00B2", "\u25A0", "\uFFFD",
}

func (e *editorUI) showExtendedCharPicker() {
	if e.window == nil {
		return
	}

	var dlg dialog.Dialog

	// Create a grid of characters
	grid := container.New(layout.NewGridLayout(16))

	infoLabel := widget.NewLabel("Select a character to insert")
	infoLabel.Alignment = fyne.TextAlignCenter

	for i := 0; i < 256; i++ {
		idx := i
		char := MSXInternationalDecodeTable[idx]
		if idx == 0 {
			char = " " // Null char as space for visibility
		}

		btn := widget.NewButton(char, func() {
			toInsert := MSXInternationalDecodeTable[idx]
			if idx == 0 {
				toInsert = "\x00"
			}
			e.insertTextAtCursor(toInsert, fmt.Sprintf("Character 0x%02X", idx))
			dlg.Hide()
		})
		btn.Importance = widget.LowImportance
		grid.Add(btn)
	}

	content := container.NewBorder(
		infoLabel,
		nil,
		nil,
		nil,
		container.NewVScroll(grid),
	)

	dlg = dialog.NewCustomWithoutButtons("Extended Characters (MSX International)", content, e.window)
	dlg.Resize(fyne.NewSize(640, 480))
	dlg.Show()
}
