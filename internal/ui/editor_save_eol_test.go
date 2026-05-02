package ui

import (
	"bytes"
	"testing"
)

func TestNormalizeDOSLineEndings(t *testing.T) {
	in := "10 PRINT \"A\"\n20 PRINT \"B\"\r30 END\r\n"
	got := normalizeDOSLineEndings(in)
	want := "10 PRINT \"A\"\r\n20 PRINT \"B\"\r\n30 END\r\n"
	if got != want {
		t.Fatalf("normalizeDOSLineEndings() = %q, want %q", got, want)
	}
}

func TestNormalizeDOSLineEndingsPreservesSingleLine(t *testing.T) {
	in := "10 END"
	got := normalizeDOSLineEndings(in)
	if got != in {
		t.Fatalf("single line changed unexpectedly: got %q want %q", got, in)
	}
}

func TestWriteNormalizedASCIIForCopyExport(t *testing.T) {
	var out bytes.Buffer
	in := "10 PRINT \"A\"\n20 END\n"
	wrote, err := writeNormalizedASCII(&out, in)
	if err != nil {
		t.Fatalf("writeNormalizedASCII returned error: %v", err)
	}
	want := "10 PRINT \"A\"\r\n20 END\r\n"
	if out.String() != want {
		t.Fatalf("written content = %q, want %q", out.String(), want)
	}
	if wrote != len(want) {
		t.Fatalf("written bytes = %d, want %d", wrote, len(want))
	}
}
