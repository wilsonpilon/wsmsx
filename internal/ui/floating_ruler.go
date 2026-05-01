package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// floatingRulerWidget is a floating measurement tool that can be repositioned
// over the editor. It shows distances in characters from a user-defined origin (0).
//
// Usage:
//   - Click and drag the ruler header to move it
//   - Right-click on any character in the editor to set it as the origin (0)
//   - The ruler displays character distances from that origin
type floatingRulerWidget struct {
	widget.BaseWidget

	// Position of the ruler origin (0) in the editor, in absolute character coordinates
	originCharPos int // 0-based character position in the text

	// Current cursor position for visual reference
	cursorCharPos int

	// Visual position offset from the editor start
	visualOffset float32

	// Current text being edited (for calculating positions)
	text string

	// Mouse interaction state
	isDragging bool
	dragStartX float32
	dragStartY float32
	posX       float32
	posY       float32

	// Callback for when origin is set via right-click
	onOriginSet func(charPos int)

	blockStartPos int
	blockEndPos   int
	hasBlockStart bool
	hasBlockEnd   bool

	bold   bool
	italic bool
}

const floatingRulerScaleCols = 132

func newFloatingRulerWidget() *floatingRulerWidget {
	r := &floatingRulerWidget{
		originCharPos: 0,
		cursorCharPos: 0,
		visualOffset:  0,
		posX:          50,
		posY:          50,
	}
	r.ExtendBaseWidget(r)
	return r
}

// SetBold updates the bold state so that character width measurements match
// the active editor font weight. Triggers a redraw when the value changes.
func (r *floatingRulerWidget) SetBold(b bool) {
	r.SetTextStyle(b, r.italic)
}

// SetTextStyle updates bold and italic flags used by floating ruler labels.
func (r *floatingRulerWidget) SetTextStyle(bold, italic bool) {
	if r.bold == bold && r.italic == italic {
		return
	}
	r.bold = bold
	r.italic = italic
	r.Refresh()
}

// SetOriginCharPos sets the character position that serves as the ruler's origin (0).
func (r *floatingRulerWidget) SetOriginCharPos(pos int) {
	if r.originCharPos == pos {
		return
	}
	r.originCharPos = pos
	r.Refresh()
}

func (r *floatingRulerWidget) ResetBlockSelection() {
	r.hasBlockStart = false
	r.hasBlockEnd = false
	r.blockStartPos = 0
	r.blockEndPos = 0
	r.Refresh()
}

// MarkBlockPoint toggles B-mark behavior: first B sets start, second B sets end.
func (r *floatingRulerWidget) MarkBlockPoint(pos int) string {
	if !r.hasBlockStart || r.hasBlockEnd {
		r.blockStartPos = pos
		r.hasBlockStart = true
		r.hasBlockEnd = false
		r.Refresh()
		return fmt.Sprintf("RULE: block start=%d (press B for end)", pos)
	}
	r.blockEndPos = pos
	r.hasBlockEnd = true
	r.Refresh()
	start, end := r.blockStartPos, r.blockEndPos
	if start > end {
		start, end = end, start
	}
	return fmt.Sprintf("RULE: block %d..%d (%d chars)", start, end, (end-start)+1)
}

// UpdateCursor updates the current cursor position for visual highlighting.
func (r *floatingRulerWidget) UpdateCursor(charPos int) {
	if r.cursorCharPos == charPos {
		return
	}
	r.cursorCharPos = charPos
	r.Refresh()
}

// SetText updates the text content (needed for accurate position calculations).
func (r *floatingRulerWidget) SetText(text string) {
	r.text = text
	r.Refresh()
}

// SetVisualOffset updates the visual offset for viewport scrolling.
func (r *floatingRulerWidget) SetVisualOffset(offset float32) {
	if r.visualOffset == offset {
		return
	}
	r.visualOffset = offset
	r.Refresh()
}

// SetPosition sets the visual position of the floating ruler on screen.
func (r *floatingRulerWidget) SetPosition(x, y float32) {
	r.posX = x
	r.posY = y
	r.Refresh()
}

// GetPosition returns the current visual position of the floating ruler.
func (r *floatingRulerWidget) GetPosition() (x, y float32) {
	return r.posX, r.posY
}

// SetOriginSetCallback sets a callback function for when the origin is changed via events.
func (r *floatingRulerWidget) SetOriginSetCallback(cb func(int)) {
	r.onOriginSet = cb
}

// Dragged handles mouse dragging for repositioning the ruler.
func (r *floatingRulerWidget) Dragged(ev *fyne.DragEvent) {
	if ev == nil {
		return
	}
	// Move the ruler based on drag delta
	r.posX += ev.Dragged.DX
	r.posY += ev.Dragged.DY
	r.Refresh()
}

// DragEnd is called when the drag operation finishes.
func (r *floatingRulerWidget) DragEnd() {}

// SecondaryTapped handles right-click events for setting origin via context menu
func (r *floatingRulerWidget) SecondaryTapped(ev *fyne.PointEvent) {
	// Could be used for context menu in the future
}

func (r *floatingRulerWidget) CreateRenderer() fyne.WidgetRenderer {
	rnd := &floatingRulerRenderer{w: r}
	rnd.init()
	return rnd
}

// ── floatingRulerRenderer ─────────────────────────────────────────────────────

type floatingRulerRenderer struct {
	w *floatingRulerWidget

	// Visual elements
	bgRect    *canvas.Rectangle
	header    *canvas.Rectangle
	headerTxt *canvas.Text
	rulerLine *canvas.Rectangle
	scaleTop  []*canvas.Text
	scaleBot  []*canvas.Text
	markText  *canvas.Text
	distText  *canvas.Text
	blockText *canvas.Text

	objects []fyne.CanvasObject
}

func (r *floatingRulerRenderer) init() {
	// Background (slightly transparent dark)
	bgColor := color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xf0}
	r.bgRect = canvas.NewRectangle(bgColor)

	// Footer/title bar
	headerColor := color.NRGBA{R: 0x0d, G: 0x0d, B: 0x0d, A: 0xff}
	r.header = canvas.NewRectangle(headerColor)

	// Header text
	headerTextColor := color.NRGBA{R: 0xff, G: 0xe0, B: 0x00, A: 0xff}
	r.headerTxt = canvas.NewText("RULER (drag to move)  B=start/end block", headerTextColor)
	r.headerTxt.TextSize = theme.TextSize() - 2
	r.headerTxt.TextStyle = fyne.TextStyle{Monospace: true}

	// Ruler line and text
	rulerLineColor := color.NRGBA{R: 0x55, G: 0xcc, B: 0x55, A: 0xff}
	r.rulerLine = canvas.NewRectangle(rulerLineColor)
	rulerTextColor := color.NRGBA{R: 0xbb, G: 0xbb, B: 0xbb, A: 0xff}

	r.scaleTop = make([]*canvas.Text, floatingRulerScaleCols)
	for i := 0; i < floatingRulerScaleCols; i++ {
		d := byte('0' + ((i + 1) % 10))
		txt := canvas.NewText(string([]byte{d}), rulerTextColor)
		txt.TextSize = theme.TextSize()
		txt.TextStyle = fyne.TextStyle{Monospace: true}
		r.scaleTop[i] = txt
	}

	decades := decadeMarkers(floatingRulerScaleCols)
	r.scaleBot = make([]*canvas.Text, len(decades))
	for i, d := range decades {
		txt := canvas.NewText(fmt.Sprintf("%d", d), rulerTextColor)
		txt.TextSize = theme.TextSize()
		txt.TextStyle = fyne.TextStyle{Monospace: true}
		r.scaleBot[i] = txt
	}

	r.markText = canvas.NewText("", rulerTextColor)
	r.markText.TextSize = theme.TextSize()
	r.markText.TextStyle = fyne.TextStyle{Monospace: true}

	// Distance text (the main info displayed)
	distColor := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	r.distText = canvas.NewText("", distColor)
	r.distText.TextSize = theme.TextSize()
	r.distText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}

	r.blockText = canvas.NewText("", rulerTextColor)
	r.blockText.TextSize = theme.TextSize()
	r.blockText.TextStyle = fyne.TextStyle{Monospace: true}

	r.objects = []fyne.CanvasObject{
		r.bgRect,
		r.header,
		r.rulerLine,
		r.markText,
		r.headerTxt,
		r.distText,
		r.blockText,
	}
	for _, txt := range r.scaleTop {
		r.objects = append(r.objects, txt)
	}
	for _, txt := range r.scaleBot {
		r.objects = append(r.objects, txt)
	}

	r.updateText()
}

func (r *floatingRulerRenderer) charWidth() float32 {
	// Always measure with non-bold monospace — Source Code Pro Regular and Bold
	// share identical advance widths (both are monospace), so measurements are identical.
	sz := fyne.MeasureText("M", r.markText.TextSize, fyne.TextStyle{Monospace: true})
	return sz.Width
}

func (r *floatingRulerRenderer) updateText() {
	origin := r.w.originCharPos
	cursor := r.w.cursorCharPos

	// Distance from origin to cursor
	distance := cursor - origin

	// Display distance and origin info
	originStr := fmt.Sprintf("Origin: char %d  |  Cursor: char %d", origin, cursor)
	distStr := fmt.Sprintf(">>> Distance: %+d chars <<<", distance)
	blockStr := "Block: press B at start and B at end"
	if r.w.hasBlockStart && !r.w.hasBlockEnd {
		blockStr = fmt.Sprintf("Block: start=%d (waiting for B at end)", r.w.blockStartPos)
	} else if r.w.hasBlockStart && r.w.hasBlockEnd {
		start, end := r.w.blockStartPos, r.w.blockEndPos
		if start > end {
			start, end = end, start
		}
		blockStr = fmt.Sprintf("Block: %d..%d  Total: %d chars", start, end, (end-start)+1)
	}

	r.markText.Text = originStr
	r.distText.Text = distStr
	r.blockText.Text = blockStr

	r.markText.Refresh()
	r.distText.Refresh()
	r.blockText.Refresh()
}

func decadeMarkers(cols int) []int {
	if cols < 10 {
		return nil
	}
	markers := make([]int, 0, cols/10)
	for i := 10; i <= cols; i += 10 {
		markers = append(markers, i)
	}
	return markers
}

func buildRulerScaleRows(cols int) (string, string) {
	if cols <= 0 {
		return "", ""
	}
	// Top row: continuous numbering 1234567890123...
	top := make([]byte, cols)
	for i := 1; i <= cols; i++ {
		top[i-1] = byte('0' + (i % 10))
	}

	// Bottom row: decade markers (10,20,30,...) aligned to the respective columns.
	bot := make([]byte, cols)
	for i := range bot {
		bot[i] = ' '
	}
	for i := 10; i <= cols; i += 10 {
		label := fmt.Sprintf("%d", i)
		end := i - 1
		start := end - len(label) + 1
		if start < 0 {
			start = 0
		}
		for j := 0; j < len(label) && start+j < cols; j++ {
			bot[start+j] = label[j]
		}
	}
	return string(top), string(bot)
}

func (r *floatingRulerRenderer) Layout(size fyne.Size) {
	_ = size
	headerHeight := float32(28)
	borderRadius := float32(4)
	lineHeight := float32(20)
	cw := r.charWidth()
	if cw <= 0 {
		cw = 1
	}
	panelW := float32(450)
	if cw > 0 {
		scaleW := (float32(floatingRulerScaleCols) * cw) + 12
		if scaleW > panelW {
			panelW = scaleW
		}
	}
	panelH := float32(175)

	// Background with rounded corners effect (simulated with rect)
	r.bgRect.Move(fyne.NewPos(r.w.posX, r.w.posY))
	r.bgRect.Resize(fyne.NewSize(panelW, panelH))
	r.bgRect.CornerRadius = borderRadius

	// Top scale digits: one glyph per character column to avoid proportional drift.
	baseX := r.w.posX + 6
	for i, txt := range r.scaleTop {
		x := baseX + (float32(i) * cw)
		txt.Move(fyne.NewPos(x, r.w.posY+4))
		txt.Resize(fyne.NewSize(cw, lineHeight))
	}

	// Bottom decade markers aligned exactly to columns 10,20,30...
	markers := decadeMarkers(floatingRulerScaleCols)
	for i, marker := range markers {
		if i >= len(r.scaleBot) {
			break
		}
		label := fmt.Sprintf("%d", marker)
		startCol := marker - len(label)
		if startCol < 0 {
			startCol = 0
		}
		x := baseX + (float32(startCol) * cw)
		r.scaleBot[i].Move(fyne.NewPos(x, r.w.posY+4+lineHeight))
		r.scaleBot[i].Resize(fyne.NewSize(float32(len(label))*cw, lineHeight))
	}

	// Green ruler line below scale markers.
	r.rulerLine.Move(fyne.NewPos(r.w.posX, r.w.posY+4+(lineHeight*2)))
	r.rulerLine.Resize(fyne.NewSize(panelW, 2))

	// Position info and measurement text.
	contentTop := r.w.posY + 4 + (lineHeight * 2) + 6
	r.markText.Move(fyne.NewPos(r.w.posX+6, contentTop))
	r.markText.Resize(fyne.NewSize(panelW-12, lineHeight))

	r.distText.Move(fyne.NewPos(r.w.posX+6, contentTop+lineHeight))
	r.distText.Resize(fyne.NewSize(panelW-12, lineHeight+2))

	r.blockText.Move(fyne.NewPos(r.w.posX+6, contentTop+(lineHeight*2)))
	r.blockText.Resize(fyne.NewSize(panelW-12, lineHeight))

	// Footer title bar at the bottom.
	footerY := r.w.posY + panelH - headerHeight
	r.header.Move(fyne.NewPos(r.w.posX, footerY))
	r.header.Resize(fyne.NewSize(panelW, headerHeight))
	r.header.CornerRadius = borderRadius
	r.headerTxt.Move(fyne.NewPos(r.w.posX+4, footerY+4))
	r.headerTxt.Resize(fyne.NewSize(panelW-8, headerHeight-6))
}

func (r *floatingRulerRenderer) MinSize() fyne.Size {
	panelW := float32(450)
	if cw := r.charWidth(); cw > 0 {
		scaleW := (float32(floatingRulerScaleCols) * cw) + 12
		if scaleW > panelW {
			panelW = scaleW
		}
	}
	return fyne.NewSize(panelW, 175)
}

func (r *floatingRulerRenderer) Refresh() {
	textSize := theme.TextSize()
	style := fyne.TextStyle{Monospace: true, Bold: r.w.bold, Italic: r.w.italic}
	for _, txt := range r.scaleTop {
		txt.TextSize = textSize
		txt.TextStyle = style
	}
	for _, txt := range r.scaleBot {
		txt.TextSize = textSize
		txt.TextStyle = style
	}
	r.markText.TextStyle = style
	r.blockText.TextStyle = style
	r.headerTxt.TextStyle = style
	r.distText.TextStyle = fyne.TextStyle{Monospace: true, Bold: true, Italic: r.w.italic}
	r.updateText()
	// Position depends on w.posX/w.posY; force layout so dragging moves immediately.
	r.Layout(r.w.Size())
	canvas.Refresh(r.w)
}

func (r *floatingRulerRenderer) Destroy() {}

func (r *floatingRulerRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}
