package calc

import "testing"

func TestEvaluateArithmetic(t *testing.T) {
	res, err := Evaluate("2+3*4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Decimal != "14" {
		t.Fatalf("decimal=%q, want 14", res.Decimal)
	}
}

func TestEvaluatePowAndSqrt(t *testing.T) {
	res, err := Evaluate("sqr(81)+2^3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Decimal != "17" {
		t.Fatalf("decimal=%q, want 17", res.Decimal)
	}
}

func TestEvaluatePrefixedBases(t *testing.T) {
	res, err := Evaluate("&H10 + &B11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Decimal != "19" {
		t.Fatalf("decimal=%q, want 19", res.Decimal)
	}
	if res.Hex != "&H13" {
		t.Fatalf("hex=%q, want &H13", res.Hex)
	}
	if res.Binary != "&B10011" {
		t.Fatalf("bin=%q, want &B10011", res.Binary)
	}
}

func TestEvaluateBitwiseOperators(t *testing.T) {
	res, err := Evaluate("NOT 0 AND 15 XOR 3 OR 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// (((NOT 0) AND 15) XOR 3) OR 1 = 13
	if res.Decimal != "13" {
		t.Fatalf("decimal=%q, want 13", res.Decimal)
	}
}

func TestEvaluateShiftAndRotate(t *testing.T) {
	res, err := Evaluate("(1 << 4) + ror(8,1) + rol(1,3)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Decimal != "28" {
		t.Fatalf("decimal=%q, want 28", res.Decimal)
	}
}

func TestEvaluateIntAndConverters(t *testing.T) {
	res, err := Evaluate("hex(10.9)+bin(1.7)+dec(2.2)+int(5.8)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Decimal != "18" {
		t.Fatalf("decimal=%q, want 18", res.Decimal)
	}
}

func TestEvaluateWithLastDotReference(t *testing.T) {
	first, err := Evaluate("1+1")
	if err != nil {
		t.Fatalf("unexpected error on first eval: %v", err)
	}
	if first.Decimal != "2" {
		t.Fatalf("first decimal=%q, want 2", first.Decimal)
	}

	second, err := EvaluateWithLast("4 * .", first.Value, true)
	if err != nil {
		t.Fatalf("unexpected error on second eval: %v", err)
	}
	if second.Decimal != "8" {
		t.Fatalf("second decimal=%q, want 8", second.Decimal)
	}
}

func TestEvaluateWithLastDotWithoutPrevious(t *testing.T) {
	if _, err := EvaluateWithLast("4 * .", 0, false); err == nil {
		t.Fatalf("expected error when '.' is used without previous result")
	}
}

func TestEvaluateErrors(t *testing.T) {
	cases := []string{
		"",
		"1/0",
		"sqr(-1)",
		"&H",
		"2+",
		"foo(1)",
	}
	for _, expr := range cases {
		if _, err := Evaluate(expr); err == nil {
			t.Fatalf("expected error for %q", expr)
		}
	}
}
