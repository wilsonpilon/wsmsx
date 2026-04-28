package ui

import (
	"image/color"
	"testing"
)

func TestCustomPaletteJSONRoundTrip(t *testing.T) {
	in := syntaxPalette{
		Keyword:  color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xFF},
		Function: color.NRGBA{R: 0x22, G: 0x33, B: 0x44, A: 0xFF},
		String:   color.NRGBA{R: 0x33, G: 0x44, B: 0x55, A: 0xFF},
		Number:   color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xFF},
		Comment:  color.NRGBA{R: 0x55, G: 0x66, B: 0x77, A: 0xFF},
		Literal:  color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF},
	}

	data, err := marshalCustomPaletteJSON(in)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	out, err := parseCustomPaletteJSON(data)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if out != in {
		t.Fatalf("unexpected palette round-trip: got=%v want=%v", out, in)
	}
}

func TestParseCustomPaletteJSONInvalid(t *testing.T) {
	bad := []byte(`{"keyword":"#GG0000","function":"#112233","string":"#223344","number":"#334455","comment":"#445566"}`)
	if _, err := parseCustomPaletteJSON(bad); err == nil {
		t.Fatalf("expected parse failure for invalid hex")
	}
}

