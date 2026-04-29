package ui

import (
	"strings"
	"testing"

	"ws7/internal/basic/renum"
)

func TestFormatRenumWarnings(t *testing.T) {
	refs := []renum.UndefinedReference{
		{SourceLine: 10, Command: "GOTO", Target: 999, Category: renum.WarningCategoryFlow, Severity: renum.WarningSeverityWarning},
		{SourceLine: 20, Command: "LLIST", Target: 888, Category: renum.WarningCategoryListing, Severity: renum.WarningSeverityInfo},
	}
	got := formatRenumWarnings(refs)
	if !strings.Contains(got, "Flow warnings (severity: warning)") {
		t.Fatalf("warning text missing flow section: %q", got)
	}
	if !strings.Contains(got, "Listing notices (severity: info)") {
		t.Fatalf("warning text missing listing section: %q", got)
	}
	if !strings.Contains(got, "Source 10: GOTO 999") {
		t.Fatalf("warning text missing first reference: %q", got)
	}
	if !strings.Contains(got, "Source 20: LLIST 888") {
		t.Fatalf("warning text missing second reference: %q", got)
	}
}

func TestFormatRenumWarningsTruncatesLongList(t *testing.T) {
	refs := make([]renum.UndefinedReference, 0, 13)
	for i := 0; i < 13; i++ {
		refs = append(refs, renum.UndefinedReference{SourceLine: i + 1, Command: "GOTO", Target: i + 100})
	}
	got := formatRenumWarnings(refs)
	if !strings.Contains(got, "and 1 more item(s)") {
		t.Fatalf("expected truncation note, got %q", got)
	}
}
