package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// syntaxColorSwatch renders a clickable color sample used by the syntax theme editor.
type syntaxColorSwatch struct {
	widget.BaseWidget
	fill     color.NRGBA
	enabled  bool
	onTapped func()
}

func newSyntaxColorSwatch(fill color.NRGBA, onTapped func()) *syntaxColorSwatch {
	w := &syntaxColorSwatch{fill: fill, enabled: true, onTapped: onTapped}
	w.ExtendBaseWidget(w)
	return w
}

func (w *syntaxColorSwatch) SetColor(fill color.NRGBA) {
	if w.fill == fill {
		return
	}
	w.fill = fill
	w.Refresh()
}

func (w *syntaxColorSwatch) SetEnabled(enabled bool) {
	if w.enabled == enabled {
		return
	}
	w.enabled = enabled
	w.Refresh()
}

func (w *syntaxColorSwatch) Tapped(_ *fyne.PointEvent) {
	if !w.enabled || w.onTapped == nil {
		return
	}
	w.onTapped()
}

func (w *syntaxColorSwatch) TappedSecondary(_ *fyne.PointEvent) {}

func (w *syntaxColorSwatch) CreateRenderer() fyne.WidgetRenderer {
	r := &syntaxColorSwatchRenderer{w: w}
	r.swatch = canvas.NewRectangle(w.fill)
	r.swatch.StrokeColor = color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF}
	r.swatch.StrokeWidth = 1
	r.swatch.CornerRadius = theme.Size(theme.SizeNameSelectionRadius)
	r.disabledOverlay = canvas.NewRectangle(color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0x88})
	r.objects = []fyne.CanvasObject{r.swatch, r.disabledOverlay}
	r.Refresh()
	return r
}

type syntaxColorSwatchRenderer struct {
	w               *syntaxColorSwatch
	swatch          *canvas.Rectangle
	disabledOverlay *canvas.Rectangle
	objects         []fyne.CanvasObject
}

func (r *syntaxColorSwatchRenderer) Layout(size fyne.Size) {
	r.swatch.Resize(size)
	r.disabledOverlay.Resize(size)
}

func (r *syntaxColorSwatchRenderer) MinSize() fyne.Size {
	return fyne.NewSize(40, theme.TextSize()+8)
}

func (r *syntaxColorSwatchRenderer) Refresh() {
	r.swatch.FillColor = r.w.fill
	r.swatch.Refresh()
	if r.w.enabled {
		r.disabledOverlay.Hide()
	} else {
		r.disabledOverlay.Show()
		r.disabledOverlay.Refresh()
	}
}

func (r *syntaxColorSwatchRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *syntaxColorSwatchRenderer) Destroy() {}
