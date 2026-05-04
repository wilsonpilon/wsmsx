package ui

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"ws7/internal/syntax"
)

func almostEqualFloat32(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

func firstOverlayText(t *testing.T, r *syntaxOverlayRenderer) *canvas.Text {
	t.Helper()
	for _, obj := range r.objects {
		if txt, ok := obj.(*canvas.Text); ok {
			return txt
		}
	}
	t.Fatal("expected at least one canvas.Text in syntax overlay objects")
	return nil
}

func TestSyntaxOverlayVerticalInsetMatchesEntryBaseline(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })

	w := newSyntaxOverlayWidget(syntax.DialectMSXBasicOfficial, defaultSyntaxThemeID)
	w.SetText("10 PRINT \"HELLO\"")
	w.SetViewport(0, 0, 0)

	r := &syntaxOverlayRenderer{w: w}
	r.rebuild(fyne.NewSize(800, 200))

	txt := firstOverlayText(t, r)
	wantY := theme.Size(theme.SizeNameInnerPadding) - theme.Size(theme.SizeNameInputBorder)
	if !almostEqualFloat32(txt.Position().Y, wantY) {
		t.Fatalf("first token y = %.3f, want %.3f", txt.Position().Y, wantY)
	}
}

func TestSyntaxOverlayVerticalInsetRespectsTopOffset(t *testing.T) {
	a := test.NewApp()
	t.Cleanup(func() { a.Quit() })

	w := newSyntaxOverlayWidget(syntax.DialectMSXBasicOfficial, defaultSyntaxThemeID)
	w.SetText("10 PRINT \"HELLO\"")
	w.SetViewport(0, 3.5, 0)

	r := &syntaxOverlayRenderer{w: w}
	r.rebuild(fyne.NewSize(800, 200))

	txt := firstOverlayText(t, r)
	wantY := theme.Size(theme.SizeNameInnerPadding) - theme.Size(theme.SizeNameInputBorder) - 3.5
	if !almostEqualFloat32(txt.Position().Y, wantY) {
		t.Fatalf("first token y with topOffset = %.3f, want %.3f", txt.Position().Y, wantY)
	}
}
