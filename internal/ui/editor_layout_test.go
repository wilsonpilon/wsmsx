package ui

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
)

func almostEqual32(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.001
}

func TestRulerStartAtTextLayoutOffsetsByGutterWidth(t *testing.T) {
	gutter := newLineNumbersWidget()
	ruler := newRulerWidget()
	layout := &rulerStartAtTextLayout{gutter: gutter}

	size := fyne.NewSize(500, ruler.MinSize().Height)
	layout.Layout([]fyne.CanvasObject{ruler}, size)

	wantX := gutter.MinSize().Width
	if got := ruler.Position().X; !almostEqual32(got, wantX) {
		t.Fatalf("ruler x = %.3f, want %.3f", got, wantX)
	}
	if got := ruler.Size().Width; !almostEqual32(got, size.Width-wantX) {
		t.Fatalf("ruler width = %.3f, want %.3f", got, size.Width-wantX)
	}
}

func TestRulerStartAtTextLayoutMinSizeIncludesGutterWidth(t *testing.T) {
	gutter := newLineNumbersWidget()
	ruler := newRulerWidget()
	layout := &rulerStartAtTextLayout{gutter: gutter}

	got := layout.MinSize([]fyne.CanvasObject{ruler})
	wantWidth := gutter.MinSize().Width + ruler.MinSize().Width
	wantHeight := ruler.MinSize().Height

	if !almostEqual32(got.Width, wantWidth) {
		t.Fatalf("min width = %.3f, want %.3f", got.Width, wantWidth)
	}
	if !almostEqual32(got.Height, wantHeight) {
		t.Fatalf("min height = %.3f, want %.3f", got.Height, wantHeight)
	}
}
