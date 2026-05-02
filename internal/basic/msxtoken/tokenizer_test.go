package msxtoken

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestTokenizeProgramSimplePrint(t *testing.T) {
	program := "10 PRINT \"A\"\n20 END\n"
	out, err := TokenizeProgram(program)
	if err != nil {
		t.Fatalf("TokenizeProgram returned error: %v", err)
	}
	if len(out) < 10 {
		t.Fatalf("unexpected short output: %d bytes", len(out))
	}
	if out[0] != 0xff {
		t.Fatalf("first byte = %02x, want ff", out[0])
	}
	if out[len(out)-2] != 0x00 || out[len(out)-1] != 0x00 {
		t.Fatalf("expected program terminator 0000, got %s", hex.EncodeToString(out[len(out)-2:]))
	}
	if !strings.Contains(hex.EncodeToString(out), "91") {
		t.Fatalf("expected PRINT token 0x91 in output")
	}
}

func TestTokenizeProgramRejectsOutOfOrderLines(t *testing.T) {
	_, err := TokenizeProgram("20 PRINT 1\n10 PRINT 2\n")
	if err == nil {
		t.Fatal("expected out-of-order error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "out of order") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTokenizeProgramNumericModes(t *testing.T) {
	program := strings.Join([]string{
		"10 A=9",
		"20 B=255",
		"30 C=32767",
		"40 D=32768",
		"50 E=1.25",
		"60 F=1.5E+3",
		"70 G=12#",
		"80 H=12!",
		"90 I=12%",
		"100 END",
	}, "\n") + "\n"

	out, err := TokenizeProgram(program)
	if err != nil {
		t.Fatalf("TokenizeProgram returned error: %v", err)
	}
	hexOut := hex.EncodeToString(out)
	for _, marker := range []string{"1d", "1f", "0f", "1c"} {
		if !strings.Contains(hexOut, marker) {
			t.Fatalf("expected output to contain numeric marker %q, got %s", marker, hexOut)
		}
	}
}

func TestTokenizeProgramIntegerOverflow(t *testing.T) {
	_, err := TokenizeProgram("10 A=32768%\n")
	if err == nil {
		t.Fatal("expected integer overflow error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "integer overflow") {
		t.Fatalf("unexpected error: %v", err)
	}
}
