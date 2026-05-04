package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
)

func TestLineNumbersCursorHighlightZeroBased(t *testing.T) {
	tests := []struct {
		name       string
		lineCount  int
		topLine    int
		cursorLine int
		expectedI  int
	}{
		{
			name:       "cursor on first visible line",
			lineCount:  5,
			topLine:    0,
			cursorLine: 0,
			expectedI:  0,
		},
		{
			name:       "cursor on second visible line",
			lineCount:  5,
			topLine:    0,
			cursorLine: 1,
			expectedI:  1,
		},
		{
			name:       "cursor row 12 when top line is 11",
			lineCount:  20,
			topLine:    11,
			cursorLine: 12,
			expectedI:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newLineNumbersWidget()
			w.UpdateWithOffset(tt.lineCount, tt.topLine, 0, tt.cursorLine)
			r := &lineNumbersRenderer{w: w}
			r.init()
			r.Layout(fyne.NewSize(120, 300))

			curVisIdx := w.cursorLine - w.topLine
			if curVisIdx != tt.expectedI {
				t.Fatalf("cursor visible index = %d, want %d", curVisIdx, tt.expectedI)
			}

			if curVisIdx < 0 || curVisIdx >= len(r.texts) {
				t.Fatalf("cursor visible index out of range: %d", curVisIdx)
			}
			if got := r.texts[curVisIdx].Color; got != (color.Color)(lineNumCursorColor) {
				t.Fatalf("cursor row color = %#v, want %#v", got, lineNumCursorColor)
			}
		})
	}
}

func TestLineNumbersTopOffsetUsesPixels(t *testing.T) {
	w := newLineNumbersWidget()
	w.UpdateWithOffset(30, 10, 7.5, 11)
	if w.topOffset != 7.5 {
		t.Fatalf("topOffset = %.2f, want 7.50", w.topOffset)
	}
}
