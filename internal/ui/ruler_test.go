package ui

import (
	"strings"
	"testing"
)

func TestRulerUses132ColumnsAndZeroBasedDigits(t *testing.T) {
	rw := newRulerWidget()
	rr := &rulerRenderer{w: rw}
	rr.init()

	if got := len(rr.rowUnits.Text); got != 132 {
		t.Fatalf("units length = %d, want 132", got)
	}
	if got := rr.rowUnits.Text[:20]; got != "01234567890123456789" {
		t.Fatalf("units prefix = %q, want %q", got, "01234567890123456789")
	}
	if got := len(rr.markRects); got != 3 {
		t.Fatalf("mark rect count = %d, want 3", got)
	}
}

func TestRulerCursorTextUsesZeroBasedColumn(t *testing.T) {
	rw := newRulerWidget()
	rw.UpdateCursor(4, 31) // Ln 5, Col 31
	rr := &rulerRenderer{w: rw}
	rr.init()

	if !strings.Contains(rr.rowCursor.Text, "Col:31") {
		t.Fatalf("cursor text missing zero-based column: %q", rr.rowCursor.Text)
	}
	if !strings.Contains(rr.rowCursor.Text, "Ln:5") {
		t.Fatalf("cursor text missing line: %q", rr.rowCursor.Text)
	}
}
