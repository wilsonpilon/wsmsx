package msxtoken

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var detokRoundtripStrictBytes = flag.Bool("msxtoken.detok.roundtrip.strict", false, "enable strict byte-for-byte roundtrip validation")
var detokRoundtripReport = flag.Bool("msxtoken.detok.roundtrip.report", false, "print per-fixture roundtrip summary (strict/structural)")

func TestDecodeProgramTextNonTokenized(t *testing.T) {
	in := []byte("10 PRINT \"A\"\r\n")
	text, tokenized, err := DecodeProgramText(in)
	if err != nil {
		t.Fatalf("DecodeProgramText returned error: %v", err)
	}
	if tokenized {
		t.Fatal("expected non-tokenized input")
	}
	if text != string(in) {
		t.Fatalf("decoded text changed unexpectedly: %q", text)
	}
}

func TestDecodeProgramTextTokenizedRoundtrip(t *testing.T) {
	src := "10 PRINT \"A\"\r\n20 END\r\n"
	bin, err := TokenizeProgram(src)
	if err != nil {
		t.Fatalf("TokenizeProgram returned error: %v", err)
	}
	text, tokenized, err := DecodeProgramText(bin)
	if err != nil {
		t.Fatalf("DecodeProgramText returned error: %v", err)
	}
	if !tokenized {
		t.Fatal("expected tokenized input")
	}
	if !strings.Contains(text, "10 PRINT") {
		t.Fatalf("decoded text missing line 10 PRINT: %q", text)
	}
	if !strings.Contains(text, "20 END") {
		t.Fatalf("decoded text missing line 20 END: %q", text)
	}
}

func TestDecodeProgramTextAggressiveRoundtripFixtures(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(thisFile)
	fixtures, err := filepath.Glob(filepath.Join(pkgDir, "testdata", "fixtures", "*.asc"))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("failed to list fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			mode := "structural"
			if *detokRoundtripStrictBytes {
				mode = "strict-bytes"
			}
			if *detokRoundtripReport {
				t.Logf("roundtrip fixture=%s mode=%s", filepath.Base(fixture), mode)
			}

			src, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			firstBin, err := TokenizeProgram(string(src))
			if err != nil {
				t.Fatalf("first tokenize failed: %v", err)
			}

			decoded, tokenized, err := DecodeProgramText(firstBin)
			if err != nil {
				t.Fatalf("decode failed: %v", err)
			}
			if !tokenized {
				t.Fatal("expected tokenized=true after decoding generated binary")
			}
			if !strings.Contains(decoded, "\r\n") {
				t.Fatalf("decoded text should contain CRLF line endings: %q", decoded)
			}

			secondBin, err := TokenizeProgram(decoded)
			if err != nil {
				t.Fatalf("retokenize failed: %v", err)
			}
			if *detokRoundtripReport {
				t.Logf("bytes fixture=%s first=%d second=%d equal=%v", filepath.Base(fixture), len(firstBin), len(secondBin), bytes.Equal(firstBin, secondBin))
			}

			if !bytes.Equal(firstBin, secondBin) {
				if *detokRoundtripStrictBytes {
					t.Fatalf("strict-bytes mismatch for %s: first=%d bytes second=%d bytes", filepath.Base(fixture), len(firstBin), len(secondBin))
				}
				firstLines := decodeTokenizedLines(firstBin)
				secondLines := decodeTokenizedLines(secondBin)
				if len(firstLines) != len(secondLines) {
					t.Fatalf("structural mismatch (line count) for %s: first=%d second=%d", filepath.Base(fixture), len(firstLines), len(secondLines))
				}
				for i := range firstLines {
					if firstLines[i].line != secondLines[i].line {
						t.Fatalf("structural mismatch (line number) for %s at idx=%d: first=%d second=%d", filepath.Base(fixture), i, firstLines[i].line, secondLines[i].line)
					}
				}
			}
		})
	}
}
