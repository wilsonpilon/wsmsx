package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/syntax"
)

const syntaxMaxVisibleLines = 180

// syntaxOverlayWidget renders highlighted tokens over the hidden text entry.
type syntaxOverlayWidget struct {
	widget.BaseWidget
	dialectID string
	text      string
	tokens    [][]syntax.Token
	topLine   int
	bold      bool
	italic    bool
	palette   syntaxPalette
}

func newSyntaxOverlayWidget(dialectID, syntaxThemeID string) *syntaxOverlayWidget {
	w := &syntaxOverlayWidget{
		dialectID: dialectID,
		palette:   syntaxPaletteByID(syntaxThemeID),
		tokens:    [][]syntax.Token{{{Kind: syntax.TokenPlain, Value: ""}}},
	}
	w.ExtendBaseWidget(w)
	return w
}

func (w *syntaxOverlayWidget) SetText(text string) {
	if w.text == text {
		return
	}
	w.text = text
	w.tokens = syntax.HighlightDocument(w.dialectID, text)
	if len(w.tokens) == 0 {
		w.tokens = [][]syntax.Token{{{Kind: syntax.TokenPlain, Value: ""}}}
	}
	w.Refresh()
}

func (w *syntaxOverlayWidget) SetTopLine(topLine int) {
	if topLine < 0 {
		topLine = 0
	}
	if w.topLine == topLine {
		return
	}
	w.topLine = topLine
	w.Refresh()
}

func (w *syntaxOverlayWidget) SetTextStyle(bold, italic bool) {
	if w.bold == bold && w.italic == italic {
		return
	}
	w.bold = bold
	w.italic = italic
	w.Refresh()
}

func (w *syntaxOverlayWidget) SetSyntaxThemeID(themeID string) {
	next := syntaxPaletteByID(themeID)
	if w.palette == next {
		return
	}
	w.palette = next
	w.Refresh()
}

func (w *syntaxOverlayWidget) SetDialect(dialectID string) {
	if w.dialectID == dialectID {
		return
	}
	w.dialectID = dialectID
	w.tokens = syntax.HighlightDocument(w.dialectID, w.text)
	if len(w.tokens) == 0 {
		w.tokens = [][]syntax.Token{{{Kind: syntax.TokenPlain, Value: ""}}}
	}
	w.Refresh()
}

func (w *syntaxOverlayWidget) tokenColor(kind syntax.TokenKind) color.NRGBA {
	switch kind {
	case syntax.TokenKeyword:
		return w.palette.Instruction
	case syntax.TokenJump:
		return w.palette.Jump
	case syntax.TokenFunction:
		return w.palette.Function
	case syntax.TokenOperator:
		return w.palette.Operator
	case syntax.TokenNumber:
		return w.palette.Number
	case syntax.TokenString:
		return w.palette.String
	case syntax.TokenComment:
		return w.palette.Comment
	case syntax.TokenIdent:
		return w.palette.Identifier
	default:
		return w.palette.Plain
	}
}

func (w *syntaxOverlayWidget) CreateRenderer() fyne.WidgetRenderer {
	r := &syntaxOverlayRenderer{w: w}
	r.rebuild(w.Size())
	return r
}

type syntaxOverlayRenderer struct {
	w       *syntaxOverlayWidget
	objects []fyne.CanvasObject
}

func (r *syntaxOverlayRenderer) textStyle() fyne.TextStyle {
	return fyne.TextStyle{Monospace: true, Bold: r.w.bold, Italic: r.w.italic}
}

func (r *syntaxOverlayRenderer) lineHeight() float32 {
	sz := fyne.MeasureText("M", theme.TextSize(), r.textStyle())
	if sz.Height < 1 {
		return 18
	}
	return sz.Height + 2
}

func (r *syntaxOverlayRenderer) rebuild(size fyne.Size) {
	style := r.textStyle()
	textSize := theme.TextSize()
	lineHeight := r.lineHeight()
	if lineHeight <= 0 {
		lineHeight = 18
	}
	visible := int(size.Height/lineHeight) + 2
	if visible < 1 {
		visible = 1
	}
	if visible > syntaxMaxVisibleLines {
		visible = syntaxMaxVisibleLines
	}

	r.objects = r.objects[:0]
	if len(r.w.tokens) == 0 {
		return
	}

	start := r.w.topLine
	if start < 0 {
		start = 0
	}
	if start >= len(r.w.tokens) {
		start = len(r.w.tokens) - 1
	}
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > len(r.w.tokens) {
		end = len(r.w.tokens)
	}

	y := float32(0)
	for lineIdx := start; lineIdx < end; lineIdx++ {
		x := float32(0)
		lineTokens := r.w.tokens[lineIdx]
		for _, tok := range lineTokens {
			if tok.Value == "" {
				continue
			}
			t := canvas.NewText(tok.Value, r.w.tokenColor(tok.Kind))
			t.TextStyle = style
			t.TextSize = textSize
			t.Move(fyne.NewPos(x, y))
			measured := fyne.MeasureText(tok.Value, textSize, style).Width
			if measured < 1 {
				measured = 1
			}
			t.Resize(fyne.NewSize(measured+1, lineHeight))
			r.objects = append(r.objects, t)
			x += measured
		}
		y += lineHeight
	}
}

func (r *syntaxOverlayRenderer) Layout(size fyne.Size) {
	r.rebuild(size)
}

func (r *syntaxOverlayRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, r.lineHeight())
}

func (r *syntaxOverlayRenderer) Refresh() {
	size := r.w.Size()
	if size.Width <= 0 || size.Height <= 0 {
		size = r.MinSize()
	}
	r.rebuild(size)
	canvas.Refresh(r.w)
}

func (r *syntaxOverlayRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *syntaxOverlayRenderer) Destroy() {}
