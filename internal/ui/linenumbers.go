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
	lineNumBgColor       = color.NRGBA{R: 0x1e, G: 0x1e, B: 0x28, A: 0xff} // gutter background
	lineNumCurBgColor    = color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0x22} // cursor-row tint
	lineNumNormalColor   = color.NRGBA{R: 0x66, G: 0x66, B: 0x88, A: 0xff} // dimmed number
	lineNumCursorColor   = color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0xff} // highlighted number
	lineNumSepColor      = color.NRGBA{R: 0x44, G: 0x44, B: 0x55, A: 0xff} // right-edge separator
)

// ── lineNumbersWidget ─────────────────────────────────────────────────────────

// lineNumbersWidget renders a column of line numbers aligned with the editor.
// Callers update it via Update(totalLines, topLine, cursorLine) whenever the
// cursor moves or the text changes.
type lineNumbersWidget struct {
	widget.BaseWidget
	lineCount  int // total lines in document
	topLine    int // 0-based first visible line
	cursorLine int // 0-based cursor line (highlighted)
}

func newLineNumbersWidget() *lineNumbersWidget {
	w := &lineNumbersWidget{lineCount: 1}
	w.ExtendBaseWidget(w)
	return w
}

// Update refreshes the widget with new state; redraws only when anything changed.
func (w *lineNumbersWidget) Update(lineCount, topLine, cursorLine int) {
	if w.lineCount == lineCount && w.topLine == topLine && w.cursorLine == cursorLine {
		return
	}
	w.lineCount = lineCount
	w.topLine = topLine
	w.cursorLine = cursorLine
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
	sep      *canvas.Rectangle  // right-edge separator line
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
		t.TextStyle = fyne.TextStyle{Monospace: true}
		t.TextSize = theme.TextSize()
		r.texts[i] = t
		r.objects = append(r.objects, t)
	}
}

// lh returns the height of one line (matches the ruler row height).
func (r *lineNumbersRenderer) lh() float32 {
	sz := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})
	return sz.Height + 2
}

// gutterWidth returns the width of the gutter in pixels.
func (r *lineNumbersRenderer) gutterWidth() float32 {
	sz := fyne.MeasureText("M", theme.TextSize(), fyne.TextStyle{Monospace: true})
	// (lineNumCharWidth + 2 spaces padding) characters + 1-px separator
	return float32(lineNumCharWidth+2)*sz.Width + 2
}

func (r *lineNumbersRenderer) Layout(size fyne.Size) {
	lh := r.lh()
	ts := theme.TextSize()

	// Background fills entire gutter
	r.bg.Move(fyne.NewPos(0, 0))
	r.bg.Resize(size)

	// Right-edge separator (1 px wide, full height)
	r.sep.Move(fyne.NewPos(size.Width-1, 0))
	r.sep.Resize(fyne.NewSize(1, size.Height))

	// Cursor-line background highlight
	curVisIdx := r.w.cursorLine - r.w.topLine
	if curVisIdx >= 0 && curVisIdx < lineNumMaxVisible {
		r.cursorBg.Move(fyne.NewPos(0, float32(curVisIdx)*lh))
		r.cursorBg.Resize(fyne.NewSize(size.Width, lh))
		r.cursorBg.Show()
	} else {
		r.cursorBg.Hide()
	}

	// Line number labels
	for i, t := range r.texts {
		lineNum := r.w.topLine + i + 1 // 1-based for display
		y := float32(i) * lh

		t.Move(fyne.NewPos(0, y))
		t.Resize(fyne.NewSize(size.Width-2, lh))
		t.TextSize = ts

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
	}
}

func (r *lineNumbersRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.gutterWidth(), r.lh())
}

func (r *lineNumbersRenderer) Refresh() {
	ts := theme.TextSize()
	for _, t := range r.texts {
		t.TextSize = ts
	}
	canvas.Refresh(r.w)
}

func (r *lineNumbersRenderer) Destroy() {}

func (r *lineNumbersRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

