package ui

import "testing"

func TestNormalizeBlockRange(t *testing.T) {
	start, end, ok := normalizeBlockRange(8, 3, 20)
	if !ok {
		t.Fatal("expected valid range")
	}
	if start != 3 || end != 8 {
		t.Fatalf("expected (3,8), got (%d,%d)", start, end)
	}
}

func TestNormalizeBlockRangeEmpty(t *testing.T) {
	_, _, ok := normalizeBlockRange(4, 4, 20)
	if ok {
		t.Fatal("expected empty range to be invalid")
	}
}

func TestDeleteTextRange(t *testing.T) {
	got := deleteTextRange("abcdef", 2, 5)
	if got != "abf" {
		t.Fatalf("expected abf, got %q", got)
	}
}

func TestMoveTextRangeForward(t *testing.T) {
	got, cursor := moveTextRange("0123456789", 2, 5, 9)
	if got != "0156782349" {
		t.Fatalf("unexpected text after move: %q", got)
	}
	if cursor != 9 {
		t.Fatalf("expected cursor 9, got %d", cursor)
	}
}

func TestMoveTextRangeInsideBlock(t *testing.T) {
	got, cursor := moveTextRange("abcdefgh", 2, 6, 4)
	if got != "abcdefgh" {
		t.Fatalf("expected unchanged text when destination is inside block, got %q", got)
	}
	if cursor != 6 {
		t.Fatalf("expected cursor 6, got %d", cursor)
	}
}

func TestOffsetToRowCol(t *testing.T) {
	row, col := offsetToRowCol("aa\nbbb\ncc", 6)
	if row != 1 || col != 3 {
		t.Fatalf("expected row=1 col=3, got row=%d col=%d", row, col)
	}
}

func TestBlockIndicatorForMarks(t *testing.T) {
	if got := blockIndicatorForMarks(false, false); got != "" {
		t.Fatalf("expected empty indicator, got %q", got)
	}
	if got := blockIndicatorForMarks(true, false); got != "[WS7-BLOCK:B] " {
		t.Fatalf("expected [WS7-BLOCK:B], got %q", got)
	}
	if got := blockIndicatorForMarks(false, true); got != "[WS7-BLOCK:K] " {
		t.Fatalf("expected [WS7-BLOCK:K], got %q", got)
	}
	if got := blockIndicatorForMarks(true, true); got != "[WS7-BLOCK:B,K] " {
		t.Fatalf("expected [WS7-BLOCK:B,K], got %q", got)
	}
}

func TestInternalClipboardIndicator(t *testing.T) {
	if got := internalClipboardIndicator(""); got != "" {
		t.Fatalf("expected empty indicator, got %q", got)
	}
	if got := internalClipboardIndicator("abc"); got != "[WS7-CLIP:3]" {
		t.Fatalf("expected [WS7-CLIP:3], got %q", got)
	}
}
