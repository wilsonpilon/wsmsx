package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	lineNumMaxVisible = 150 // max pre-created label rows
	lineNumCharWidth  = 5   // digits wide (covers up to 99 999 lines)
)

// ── Colours ───────────────────────────────────────────────────────────────────

var (
	lineNumBgColor     = color.NRGBA{R: 0x1e, G: 0x1e, B: 0x28, A: 0xff} // gutter background
	lineNumCurBgColor  = color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0x22} // cursor-row tint
	lineNumNormalColor = color.NRGBA{R: 0x66, G: 0x66, B: 0x88, A: 0xff} // dimmed number
	lineNumCursorColor = color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0xff} // highlighted number
	lineNumSepColor    = color.NRGBA{R: 0x44, G: 0x44, B: 0x55, A: 0xff} // right-edge separator
)

// ── lineNumbersWidget ─────────────────────────────────────────────────────────

// lineNumbersWidget renders a column of line numbers aligned with the editor.
// Callers update it via Update(totalLines, topLine, cursorLine) whenever the
// cursor moves or the text changes.
// For pixel-perfect alignment during partial scrolls, use UpdateWithOffset.
type lineNumbersWidget struct {
	widget.BaseWidget
	lineCount  int     // total lines in document
	topLine    int     // 0-based first visible line
	topOffset  float32 // pixel offset within the first visible line
	cursorLine int     // 0-based cursor line (highlighted)
	bold       bool
	italic     bool
}

func newLineNumbersWidget() *lineNumbersWidget {
	w := &lineNumbersWidget{lineCount: 1}
	w.ExtendBaseWidget(w)
	return w
}

// Update refreshes the widget with new state; redraws only when anything changed.
func (w *lineNumbersWidget) Update(lineCount, topLine, cursorLine int) {
	w.UpdateWithOffset(lineCount, topLine, 0, cursorLine)
}

// UpdateWithOffset refreshes the widget with viewport state including scroll offset.
// topOffset is the pixel offset within the first visible line.
func (w *lineNumbersWidget) UpdateWithOffset(lineCount, topLine int, topOffset float32, cursorLine int) {
	if topOffset < 0 {
		topOffset = 0
	}
	if w.lineCount == lineCount && w.topLine == topLine && w.topOffset == topOffset && w.cursorLine == cursorLine {
		return
	}
	w.lineCount = lineCount
	w.topLine = topLine
	w.topOffset = topOffset
	w.cursorLine = cursorLine
	w.Refresh()
}

// SetBold updates the bold state so that gutter width and row height match the
// active editor font weight. Triggers a redraw when the value changes.
func (w *lineNumbersWidget) SetBold(b bool) {
	w.SetTextStyle(b, w.italic)
}

// SetTextStyle updates the text style used by gutter labels.
func (w *lineNumbersWidget) SetTextStyle(bold, italic bool) {
	if w.bold == bold && w.italic == italic {
		return
	}
	w.bold = bold
	w.italic = italic
	w.Refresh()
}

func (w *lineNumbersWidget) CreateRenderer() fyne.WidgetRenderer {
	r := &lineNumbersRenderer{w: w}
	r.init()
	return r
}

// ── lineNumbersRenderer ───────────────────────────────────────────────────────

type lineNumbersRenderer struct {
	w        *lineNumbersWidget
	bg       *canvas.Rectangle
	cursorBg *canvas.Rectangle
	sep      *canvas.Rectangle // right-edge separator line
	texts    []*canvas.Text
	objects  []fyne.CanvasObject
}

func (r *lineNumbersRenderer) init() {
	r.bg = canvas.NewRectangle(lineNumBgColor)
	r.cursorBg = canvas.NewRectangle(lineNumCurBgColor)
	r.sep = canvas.NewRectangle(lineNumSepColor)

	r.objects = []fyne.CanvasObject{r.bg, r.cursorBg, r.sep}

	r.texts = make([]*canvas.Text, lineNumMaxVisible)
	for i := range r.texts {
		t := canvas.NewText("", lineNumNormalColor)
		t.TextStyle = fyne.TextStyle{Monospace: true, Bold: r.w.bold, Italic: r.w.italic}
		t.TextSize = theme.TextSize()
		r.texts[i] = t
		r.objects = append(r.objects, t)
	}
}

// textStyle returns the monospace text style matching the current bold state.
func (r *lineNumbersRenderer) textStyle() fyne.TextStyle {
	return fyne.TextStyle{Monospace: true, Bold: r.w.bold, Italic: r.w.italic}
}

// lh returns the height of one line (matches editor/syntax calculations).
// We measure regular monospace to keep behavior stable across themes.
func (r *lineNumbersRenderer) lh() float32 {
	sz := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})
	return sz.Height
}

// gutterWidth returns the width of the gutter in pixels.
func (r *lineNumbersRenderer) gutterWidth() float32 {
	sz := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})
	// (lineNumCharWidth + 2 spaces padding) characters + 1-px separator
	return float32(lineNumCharWidth+2)*sz.Width + 2
}

func (r *lineNumbersRenderer) Layout(size fyne.Size) {
	lh := r.lh()
	offsetY := -r.w.topOffset

	// Background fills entire gutter
	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)

	// Right-edge separator (1 px wide, full height)
	r.sep.Move(fyne.NewPos(size.Width-1, 0))
	r.sep.Resize(fyne.NewSize(1, size.Height))

	// Cursor-line background highlight: position and height adjusted by pixel offset
	curVisIdx := r.w.cursorLine - r.w.topLine
	if curVisIdx >= 0 && curVisIdx < lineNumMaxVisible {
		cursorY := float32(curVisIdx)*lh + offsetY

		// Clamp to visible area
		if cursorY+lh > 0 && cursorY < size.Height {
			if cursorY < 0 {
				// Cursor is partially visible at top: show only the visible part
				visibleHeight := lh + cursorY
				r.cursorBg.Move(fyne.NewPos(0, 0))
				r.cursorBg.Resize(fyne.NewSize(size.Width, visibleHeight))
			} else if cursorY+lh > size.Height {
				// Cursor is partially visible at bottom: show only the visible part
				visibleHeight := size.Height - cursorY
				r.cursorBg.Move(fyne.NewPos(0, cursorY))
				r.cursorBg.Resize(fyne.NewSize(size.Width, visibleHeight))
			} else {
				// Cursor is fully visible
				r.cursorBg.Move(fyne.NewPos(0, cursorY))
				r.cursorBg.Resize(fyne.NewSize(size.Width, lh))
			}
			r.cursorBg.Show()
		} else {
			r.cursorBg.Hide()
		}
	} else {
		r.cursorBg.Hide()
	}

	r.updateRows(size, lh, offsetY)
}

func (r *lineNumbersRenderer) updateRows(size fyne.Size, lh, offsetY float32) {
	ts := theme.TextSize()
	style := r.textStyle()

	// Line number labels, positioned with pixel scroll offset
	for i, t := range r.texts {
		lineNum := r.w.topLine + i + 1 // 1-based for display
		y := float32(i)*lh + offsetY

		// Only position/show if within visible bounds
		if y+lh > 0 && y < size.Height {
			t.Move(fyne.NewPos(0, y))
			t.Resize(fyne.NewSize(size.Width-2, lh))
			t.TextSize = ts
			t.TextStyle = style

			if lineNum < 1 || lineNum > r.w.lineCount {
				t.Text = ""
				t.Hide()
				continue
			}

			t.Text = fmt.Sprintf("%*d ", lineNumCharWidth, lineNum)

			isCursor := (r.w.topLine + i) == r.w.cursorLine
			if isCursor {
				t.Color = lineNumCursorColor
			} else {
				t.Color = lineNumNormalColor
			}
			t.Show()
			t.Refresh()
		} else {
			t.Hide()
		}
	}
}

func (r *lineNumbersRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.gutterWidth(), r.lh())
}

func (r *lineNumbersRenderer) Refresh() {
	sz := r.w.Size()
	if sz.Width <= 0 || sz.Height <= 0 {
		sz = r.MinSize()
	}
	// Recompute row content on every widget refresh so text/count updates are visible
	// even when layout size does not change.
	r.Layout(sz)

	r.bg.Refresh()
	r.sep.Refresh()
	r.cursorBg.Refresh()
	for _, t := range r.texts {
		t.Refresh()
	}
}

func (r *lineNumbersRenderer) Destroy() {}

func (r *lineNumbersRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
