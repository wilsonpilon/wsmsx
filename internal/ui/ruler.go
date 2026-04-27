package ui

import (
	"fmt"
	"image/color"
	"reflect"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ── Special column marks ──────────────────────────────────────────────────────

var rulerMarkCols = []int{32, 40, 80, 132}

func isMarkCol(col int) bool {
	for _, m := range rulerMarkCols {
		if col == m {
			return true
		}
	}
	return false
}

// ── cursorEntry ───────────────────────────────────────────────────────────────

// cursorEntry extends widget.Entry and fires callbacks whenever the cursor
// position or viewport scroll changes.
type cursorEntry struct {
	widget.Entry
	onCursorMoved    func(row, col int)
	onViewportOffset func(offsetY float32)
	onRuneBeforeInput func(r rune) bool
	onShortcut       func(shortcut fyne.Shortcut) bool
}

func newCursorEntry() *cursorEntry {
	e := &cursorEntry{}
	e.MultiLine = true
	e.Wrapping = fyne.TextWrapOff
	e.ExtendBaseWidget(e)
	return e
}

func (e *cursorEntry) notify() {
	if e.onCursorMoved != nil {
		e.onCursorMoved(e.CursorRow, e.CursorColumn)
	}
}

func (e *cursorEntry) CreateRenderer() fyne.WidgetRenderer {
	r := e.Entry.CreateRenderer()
	e.attachInternalScrollHook()
	return r
}

// attachInternalScrollHook installs OnScrolled on Entry's internal scroll
// container. Fyne does not expose this field publicly, so we bridge via
// reflection to keep the gutter synchronized with the real viewport.
func (e *cursorEntry) attachInternalScrollHook() {
	if e.onViewportOffset == nil {
		return
	}
	field := reflect.ValueOf(&e.Entry).Elem().FieldByName("scroll")
	if !field.IsValid() || field.IsNil() {
		return
	}
	scroll := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	if scroll.IsNil() {
		return
	}
	scrollElem := scroll.Elem()
	onScrolled := scrollElem.FieldByName("OnScrolled")
	if onScrolled.IsValid() {
		fn := reflect.MakeFunc(onScrolled.Type(), func(args []reflect.Value) []reflect.Value {
			if len(args) > 0 {
				if pos, ok := args[0].Interface().(fyne.Position); ok {
					e.onViewportOffset(pos.Y)
				}
			}
			return nil
		})
		onScrolled.Set(fn)
	}
	if offset := scrollElem.FieldByName("Offset"); offset.IsValid() {
		if y := offset.FieldByName("Y"); y.IsValid() {
			e.onViewportOffset(float32(y.Float()))
		}
	}
}

func (e *cursorEntry) TypedKey(key *fyne.KeyEvent) {
	e.attachInternalScrollHook()
	e.Entry.TypedKey(key)
	e.notify()
}

func (e *cursorEntry) TypedRune(r rune) {
	if e.onRuneBeforeInput != nil && e.onRuneBeforeInput(r) {
		return
	}
	e.attachInternalScrollHook()
	e.Entry.TypedRune(r)
	e.notify()
}

func (e *cursorEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if e.onShortcut != nil && e.onShortcut(shortcut) {
		return
	}
	e.Entry.TypedShortcut(shortcut)
	e.notify()
}

// Tapped handles mouse clicks; Fyne updates CursorRow/CursorColumn after the
// event is processed, so we wait one frame before reading the position.
func (e *cursorEntry) Tapped(ev *fyne.PointEvent) {
	e.attachInternalScrollHook()
	e.Entry.Tapped(ev)
	go func() {
		time.Sleep(32 * time.Millisecond)
		e.notify()
	}()
}

// ── rulerWidget ───────────────────────────────────────────────────────────────

// rulerWidget draws a three-row column ruler:
//
//	Row 0  – decade numbers   (  1         2  …)
//	Row 1  – unit digits      (12345678901234…) with T marks at special cols
//	Row 2  – cursor indicator (     ^  Col:6  Ln:12)
//
// Columns 32, 40, 80 and 132 are highlighted with a semi-transparent overlay.
// The current cursor column is highlighted in yellow.
type rulerWidget struct {
	widget.BaseWidget
	cursorCol int // 0-based
	cursorRow int // 0-based (for Ln: display)
}

func newRulerWidget() *rulerWidget {
	r := &rulerWidget{}
	r.ExtendBaseWidget(r)
	return r
}

// SetCursorColumn updates only the column (0-based).
func (r *rulerWidget) SetCursorColumn(col int) {
	r.UpdateCursor(r.cursorRow, col)
}

// UpdateCursor updates both row and column (both 0-based) and redraws if changed.
func (r *rulerWidget) UpdateCursor(row, col int) {
	if r.cursorRow == row && r.cursorCol == col {
		return
	}
	r.cursorRow = row
	r.cursorCol = col
	r.Refresh()
}

func (r *rulerWidget) CreateRenderer() fyne.WidgetRenderer {
	rnd := &rulerRenderer{w: r}
	rnd.init()
	return rnd
}

// ── rulerRenderer ─────────────────────────────────────────────────────────────

const rulerMaxCols = 200 // pre-built string length (trimmed at render time)

type rulerRenderer struct {
	w *rulerWidget

	// Background: mark-column highlights
	markRects []*canvas.Rectangle // one per rulerMarkCol
	// Background: cursor column highlight
	cursorRect *canvas.Rectangle

	// Foreground text rows
	rowDecades *canvas.Text
	rowUnits   *canvas.Text
	rowCursor  *canvas.Text

	objects []fyne.CanvasObject
}

func (r *rulerRenderer) init() {
	colorDecade := color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff}
	colorUnit := color.NRGBA{R: 0xbb, G: 0xbb, B: 0xbb, A: 0xff}
	colorCursor := color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0xff}
	colorMark := color.NRGBA{R: 0x44, G: 0x88, B: 0xff, A: 0x55}   // blue, semi-transparent
	colorCurRect := color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0x50} // yellow, semi-transparent

	style := fyne.TextStyle{Monospace: true}
	ts := theme.TextSize()

	r.rowDecades = canvas.NewText("", colorDecade)
	r.rowDecades.TextStyle = style
	r.rowDecades.TextSize = ts

	r.rowUnits = canvas.NewText("", colorUnit)
	r.rowUnits.TextStyle = style
	r.rowUnits.TextSize = ts

	r.rowCursor = canvas.NewText("", colorCursor)
	r.rowCursor.TextStyle = style
	r.rowCursor.TextSize = ts

	r.markRects = make([]*canvas.Rectangle, len(rulerMarkCols))
	for i := range r.markRects {
		r.markRects[i] = canvas.NewRectangle(colorMark)
	}
	r.cursorRect = canvas.NewRectangle(colorCurRect)

	// Objects order: backgrounds first, text on top
	r.objects = []fyne.CanvasObject{}
	for _, rect := range r.markRects {
		r.objects = append(r.objects, rect)
	}
	r.objects = append(r.objects, r.cursorRect)
	r.objects = append(r.objects, r.rowDecades, r.rowUnits, r.rowCursor)

	r.updateText()
}

// charSize returns the pixel width and height of one monospace character at
// the current theme text size.
func (r *rulerRenderer) charSize() (w, h float32) {
	sz := fyne.MeasureText("M", r.rowDecades.TextSize, r.rowDecades.TextStyle)
	return sz.Width, sz.Height
}

func (r *rulerRenderer) updateText() {
	n := rulerMaxCols

	// ── Row 0: decade numbers ────────────────────────────────────────────────
	// Right-align each decade label so its last digit sits above its column.
	decades := make([]byte, n)
	for i := range decades {
		decades[i] = ' '
	}
	for col := 10; col <= n; col += 10 {
		label := fmt.Sprintf("%d", col)
		end := col - 1 // 0-based index of the last digit
		for j := len(label) - 1; j >= 0; j-- {
			pos := end - (len(label) - 1 - j)
			if pos >= 0 && pos < n {
				decades[pos] = label[j]
			}
		}
	}
	r.rowDecades.Text = string(decades)

	// ── Row 1: unit digits 1–0, T at special columns ─────────────────────────
	units := make([]byte, n)
	for i := range units {
		col1 := i + 1 // 1-based column number
		if isMarkCol(col1) {
			units[i] = 'T'
		} else {
			d := col1 % 10
			units[i] = byte('0' + d) // '0' at multiples of 10
		}
	}
	r.rowUnits.Text = string(units)

	// ── Row 2: cursor indicator ───────────────────────────────────────────────
	cursor := r.w.cursorCol // 0-based
	col1 := cursor + 1      // 1-based for display
	row1 := r.w.cursorRow + 1
	if cursor >= 0 && cursor < n {
		pad := make([]byte, cursor+1)
		for i := range pad {
			pad[i] = ' '
		}
		pad[cursor] = '^'
		r.rowCursor.Text = string(pad) + fmt.Sprintf("  Col:%-4d Ln:%-4d", col1, row1)
	} else {
		r.rowCursor.Text = fmt.Sprintf("  Col:%-4d Ln:%-4d", col1, row1)
	}

	r.rowDecades.Refresh()
	r.rowUnits.Refresh()
	r.rowCursor.Refresh()
}


func (r *rulerRenderer) Layout(size fyne.Size) {
	cw, ch := r.charSize()
	lh := ch + 2 // a touch of padding per row

	// Position the three text rows
	r.rowDecades.Move(fyne.NewPos(0, 0))
	r.rowDecades.Resize(fyne.NewSize(size.Width, lh))

	r.rowUnits.Move(fyne.NewPos(0, lh))
	r.rowUnits.Resize(fyne.NewSize(size.Width, lh))

	r.rowCursor.Move(fyne.NewPos(0, lh*2))
	r.rowCursor.Resize(fyne.NewSize(size.Width, lh))

	totalH := lh * 3

	// Position mark-column highlight rectangles
	for i, col := range rulerMarkCols {
		x := float32(col-1) * cw
		r.markRects[i].Move(fyne.NewPos(x, 0))
		r.markRects[i].Resize(fyne.NewSize(cw, totalH))
	}

	// Position cursor column highlight rectangle
	curX := float32(r.w.cursorCol) * cw
	r.cursorRect.Move(fyne.NewPos(curX, 0))
	r.cursorRect.Resize(fyne.NewSize(cw, totalH))
}

func (r *rulerRenderer) MinSize() fyne.Size {
	_, ch := r.charSize()
	lh := ch + 2
	return fyne.NewSize(200, lh*3)
}

func (r *rulerRenderer) Refresh() {
	r.rowDecades.TextSize = theme.TextSize()
	r.rowUnits.TextSize = theme.TextSize()
	r.rowCursor.TextSize = theme.TextSize()
	r.updateText()
	canvas.Refresh(r.w)
}

func (r *rulerRenderer) Destroy() {}

func (r *rulerRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

