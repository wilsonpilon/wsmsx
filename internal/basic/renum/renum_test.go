package renum

import (
	"fmt"
	"testing"
)

func TestRenumberRewritesLabelsAndReferences(t *testing.T) {
	in := "10 GOTO 30\n20 GOSUB 40\n30 IF A=1 THEN 20\n40 REM END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 5, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 GOTO 110\n105 GOSUB 115\n110 IF A=1 THEN 105\n115 REM END\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
	if res.RenumberedLines != 4 {
		t.Fatalf("RenumberedLines = %d, want 4", res.RenumberedLines)
	}
}

func TestRenumberFromLineKeepsLowerLabelsAndUpdatesCrossReferences(t *testing.T) {
	in := "10 GOTO 100\n50 IF A THEN GOTO 100\n100 PRINT \"X\"\n110 GOSUB 100\n"
	res, err := Renumber(in, Options{StartLine: 1000, Increment: 10, FromLine: 100})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "10 GOTO 1000\n50 IF A THEN GOTO 1000\n1000 PRINT \"X\"\n1010 GOSUB 1000\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberDoesNotTouchStringsOrComments(t *testing.T) {
	in := "10 PRINT \"GOTO 200\": REM GOTO 200\n20 IF A=1 THEN 10 ' GOSUB 10\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 PRINT \"GOTO 200\": REM GOTO 200\n110 IF A=1 THEN 100 ' GOSUB 10\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesOnGotoAndOnGosubLists(t *testing.T) {
	in := "10 ON X GOTO 30,50,70\n20 ON Y GOSUB 70,30\n30 PRINT \"A\"\n50 PRINT \"B\"\n70 RETURN\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 ON X GOTO 120,130,140\n110 ON Y GOSUB 140,120\n120 PRINT \"A\"\n130 PRINT \"B\"\n140 RETURN\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesElseLineTargets(t *testing.T) {
	in := "10 IF A=1 THEN 50 ELSE 80\n50 PRINT \"OK\"\n80 GOTO 50\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 IF A=1 THEN 110 ELSE 120\n110 PRINT \"OK\"\n120 GOTO 110\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesElseGotoAndElseGosubTargets(t *testing.T) {
	in := "10 IF A=1 THEN 50 ELSE GOTO 80\n20 IF B=1 THEN 80 ELSE GOSUB 50\n50 RETURN\n80 END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 IF A=1 THEN 120 ELSE GOTO 130\n110 IF B=1 THEN 130 ELSE GOSUB 120\n120 RETURN\n130 END\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesRestoreResumeRunAndListTargets(t *testing.T) {
	in := "10 RESTORE 50\n20 RESUME 60\n30 RUN 70\n40 LIST 80\n50 DATA 1\n60 PRINT \"ERR\"\n70 END\n80 PRINT \"LIST\"\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 RESTORE 140\n110 RESUME 150\n120 RUN 160\n130 LIST 170\n140 DATA 1\n150 PRINT \"ERR\"\n160 END\n170 PRINT \"LIST\"\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesReturnLineTargets(t *testing.T) {
	in := "10 GOSUB 50\n20 END\n50 PRINT \"A\":RETURN 20\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 GOSUB 120\n110 END\n120 PRINT \"A\":RETURN 110\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberLeavesResumeZeroNextAndRewritesListRange(t *testing.T) {
	in := "10 ON ERROR GOTO 50\n20 RESUME 0\n30 RESUME NEXT\n40 LIST 10-50\n50 END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 ON ERROR GOTO 140\n110 RESUME 0\n120 RESUME NEXT\n130 LIST 100-140\n140 END\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberLeavesRestoreWithoutArgumentUntouched(t *testing.T) {
	in := "10 RESTORE\n20 DATA 1,2,3\n30 GOTO 20\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 RESTORE\n110 DATA 1,2,3\n120 GOTO 110\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberLeavesErlErrAndErrorUntouchedButRewritesOnErrorGoto(t *testing.T) {
	in := "10 ON ERROR GOTO 100\n20 PRINT ERL,ERR\n30 ERROR 52\n100 RESUME 0\n"
	res, err := Renumber(in, Options{StartLine: 1000, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "1000 ON ERROR GOTO 1030\n1010 PRINT ERL,ERR\n1020 ERROR 52\n1030 RESUME 0\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesComplexThenElseMultiCommandLine(t *testing.T) {
	in := "10 IF A=1 THEN GOTO 100 ELSE PRINT \"NO\":GOSUB 200\n100 PRINT \"YES\":RESTORE 300:ELSE 999\n200 RETURN\n300 DATA 1\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 IF A=1 THEN GOTO 110 ELSE PRINT \"NO\":GOSUB 120\n110 PRINT \"YES\":RESTORE 130:ELSE 999\n120 RETURN\n130 DATA 1\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesListCommaAndOpenEndedFormats(t *testing.T) {
	in := "10 LIST 100,200\n20 LIST ,200\n30 LIST 100,\n100 PRINT \"A\"\n200 PRINT \"B\"\n"
	res, err := Renumber(in, Options{StartLine: 1000, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "1000 LIST 1030,1040\n1010 LIST ,1040\n1020 LIST 1030,\n1030 PRINT \"A\"\n1040 PRINT \"B\"\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
}

func TestRenumberRewritesLLISTAndCollectsListingInfo(t *testing.T) {
	in := "10 LLIST 50,999\n20 LLIST ,50\n50 PRINT \"A\"\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 LLIST 120,999\n110 LLIST ,120\n120 PRINT \"A\"\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
	if len(res.UndefinedRefs) != 1 {
		t.Fatalf("UndefinedRefs length = %d, want 1", len(res.UndefinedRefs))
	}
	ref := res.UndefinedRefs[0]
	if ref.SourceLine != 10 || ref.Command != "LLIST" || ref.Target != 999 {
		t.Fatalf("unexpected listing warning: %+v", ref)
	}
	if ref.Category != WarningCategoryListing || ref.Severity != WarningSeverityInfo {
		t.Fatalf("unexpected listing warning classification: %+v", ref)
	}
}

func TestRenumberDoesNotTreatKeyListAsLineListing(t *testing.T) {
	in := "10 KEY LIST\n20 LIST 10\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 KEY LIST\n110 LIST 100\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
	if len(res.UndefinedRefs) != 0 {
		t.Fatalf("UndefinedRefs length = %d, want 0", len(res.UndefinedRefs))
	}
}

func TestRenumberRewritesDeleteAndEditOptionalTargets(t *testing.T) {
	in := "10 EDIT 100\n20 DELETE 100-200\n30 DELETE ,200\n40 DELETE 100,\n100 PRINT \"A\"\n200 PRINT \"B\"\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	want := "100 EDIT 140\n110 DELETE 140-150\n120 DELETE ,150\n130 DELETE 140,\n140 PRINT \"A\"\n150 PRINT \"B\"\n"
	if res.Text != want {
		t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
	if len(res.UndefinedRefs) != 0 {
		t.Fatalf("UndefinedRefs length = %d, want 0", len(res.UndefinedRefs))
	}
}

func TestRenumberCollectsUndefinedRefsForDeleteAndEdit(t *testing.T) {
	in := "10 DELETE 999\n20 EDIT 888\n30 END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	if len(res.UndefinedRefs) != 2 {
		t.Fatalf("UndefinedRefs length = %d, want 2", len(res.UndefinedRefs))
	}
	want := map[string]bool{
		"10:DELETE:999": false,
		"20:EDIT:888":   false,
	}
	for _, ref := range res.UndefinedRefs {
		if ref.Category != WarningCategoryFlow || ref.Severity != WarningSeverityWarning {
			t.Fatalf("unexpected flow warning classification: %+v", ref)
		}
		key := fmt.Sprintf("%d:%s:%d", ref.SourceLine, ref.Command, ref.Target)
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, found := range want {
		if !found {
			t.Fatalf("missing undefined reference %s in %+v", key, res.UndefinedRefs)
		}
	}
}

func TestRenumberCollectsUndefinedReferencesWithoutFailing(t *testing.T) {
	in := "10 GOTO 999\n20 IF A THEN 888 ELSE GOSUB 777\n30 ON ERROR GOTO 0\n40 END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	if len(res.UndefinedRefs) != 3 {
		t.Fatalf("UndefinedRefs length = %d, want 3", len(res.UndefinedRefs))
	}
	want := map[string]bool{
		"10:GOTO:999":  false,
		"20:THEN:888":  false,
		"20:GOSUB:777": false,
	}
	for _, ref := range res.UndefinedRefs {
		if ref.Category != WarningCategoryFlow || ref.Severity != WarningSeverityWarning {
			t.Fatalf("unexpected flow warning classification: %+v", ref)
		}
		key := fmt.Sprintf("%d:%s:%d", ref.SourceLine, ref.Command, ref.Target)
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, found := range want {
		if !found {
			t.Fatalf("missing undefined reference %s in %+v", key, res.UndefinedRefs)
		}
	}
}

func TestSummarizeWarningsSeparatesFlowAndListingSeverity(t *testing.T) {
	stats := SummarizeWarnings([]UndefinedReference{
		{SourceLine: 10, Command: "GOTO", Target: 999},
		{SourceLine: 20, Command: "LIST", Target: 888},
		{SourceLine: 30, Command: "LLIST", Target: 777},
	})
	if stats.Total != 3 {
		t.Fatalf("Total = %d, want 3", stats.Total)
	}
	if stats.Flow != 1 || stats.Listing != 2 {
		t.Fatalf("unexpected category summary: %+v", stats)
	}
	if stats.Warning != 1 || stats.Info != 2 {
		t.Fatalf("unexpected severity summary: %+v", stats)
	}
}

func TestRenumberStrictMSXParityFailsOnUndefinedFlowReference(t *testing.T) {
	in := "10 GOTO 999\n20 END\n"
	_, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0, StrictMSXParity: true})
	if err == nil {
		t.Fatalf("Renumber() expected strict parity error, got nil")
	}
	if got := err.Error(); got != "Undefined line 999 in 10" {
		t.Fatalf("strict error mismatch: %q", got)
	}
}

func TestRenumberStrictMSXParityAllowsListingOnlyUndefined(t *testing.T) {
	in := "10 LLIST 999\n20 END\n"
	res, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0, StrictMSXParity: true})
	if err != nil {
		t.Fatalf("Renumber() error = %v", err)
	}
	if len(res.UndefinedRefs) != 1 {
		t.Fatalf("UndefinedRefs length = %d, want 1", len(res.UndefinedRefs))
	}
	if res.UndefinedRefs[0].Category != WarningCategoryListing {
		t.Fatalf("unexpected category for listing-only warning: %+v", res.UndefinedRefs[0])
	}
}

func TestRenumberStrictMSXParityFailsOnUndefinedDeleteEditReference(t *testing.T) {
	in := "10 DELETE 999\n20 EDIT 888\n"
	_, err := Renumber(in, Options{StartLine: 100, Increment: 10, FromLine: 0, StrictMSXParity: true})
	if err == nil {
		t.Fatalf("Renumber() expected strict parity error, got nil")
	}
	if got := err.Error(); got != "Undefined line 999 in 10 (and 1 more)" {
		t.Fatalf("strict error mismatch: %q", got)
	}
}

func TestAnalyzeReferencesCollectsFlowAndListingTargets(t *testing.T) {
	in := "10 ON X GOTO 100,200\n20 IF A THEN 100 ELSE GOSUB 300\n30 LLIST 100,400\n100 END\n200 END\n300 RETURN\n"
	refs := AnalyzeReferences(in)
	if len(refs) != 6 {
		t.Fatalf("AnalyzeReferences() length = %d, want 6", len(refs))
	}
	want := map[string]bool{
		"10:GOTO:100":  false,
		"10:GOTO:200":  false,
		"20:THEN:100":  false,
		"20:GOSUB:300": false,
		"30:LLIST:100": false,
		"30:LLIST:400": false,
	}
	for _, ref := range refs {
		key := fmt.Sprintf("%d:%s:%d", ref.SourceLine, ref.Command, ref.Target)
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, found := range want {
		if !found {
			t.Fatalf("missing reference %s in %+v", key, refs)
		}
	}
	for _, ref := range refs {
		if ref.Command == "LLIST" && (ref.Category != WarningCategoryListing || ref.Severity != WarningSeverityInfo) {
			t.Fatalf("unexpected LLIST classification: %+v", ref)
		}
	}
}

func TestAnalyzeReferencesIgnoresStringsCommentsAndZeroTargets(t *testing.T) {
	in := "10 PRINT \"GOTO 200\": REM GOSUB 300\n20 ON ERROR GOTO 0\n30 RESUME 0\n40 ' LIST 100\n"
	refs := AnalyzeReferences(in)
	if len(refs) != 0 {
		t.Fatalf("AnalyzeReferences() length = %d, want 0", len(refs))
	}
}

func TestDeleteRangeDeletesUnreferencedLines(t *testing.T) {
	in := "10 PRINT \"A\"\n20 PRINT \"B\"\n30 PRINT \"C\"\n40 END\n"
	res, err := DeleteRange(in, 20, 30)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}
	if len(res.BlockingRefs) != 0 {
		t.Fatalf("DeleteRange() blocking refs = %+v, want none", res.BlockingRefs)
	}
	want := "10 PRINT \"A\"\n40 END\n"
	if res.Text != want {
		t.Fatalf("DeleteRange() text mismatch\nwant:\n%q\ngot:\n%q", want, res.Text)
	}
	if res.DeletedLines != 2 {
		t.Fatalf("DeletedLines = %d, want 2", res.DeletedLines)
	}
}

func TestDeleteRangeBlocksWhenKeptFlowReferencesDeletedLine(t *testing.T) {
	in := "10 GOTO 30\n20 PRINT \"KEEP\"\n30 PRINT \"DELETE\"\n"
	res, err := DeleteRange(in, 30, 30)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}
	if len(res.BlockingRefs) != 1 {
		t.Fatalf("blocking refs length = %d, want 1", len(res.BlockingRefs))
	}
	ref := res.BlockingRefs[0]
	if ref.SourceLine != 10 || ref.Command != "GOTO" || ref.Target != 30 {
		t.Fatalf("unexpected blocking ref: %+v", ref)
	}
	if ref.Category != WarningCategoryFlow || ref.Severity != WarningSeverityWarning {
		t.Fatalf("unexpected blocking classification: %+v", ref)
	}
	if res.Text != in {
		t.Fatalf("DeleteRange() should leave text unchanged when blocked")
	}
}

func TestDeleteRangeBlocksListingReferencesFromKeptLines(t *testing.T) {
	in := "10 LIST 30\n20 END\n30 PRINT \"DELETE\"\n"
	res, err := DeleteRange(in, 30, 30)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}
	if len(res.BlockingRefs) != 1 {
		t.Fatalf("blocking refs length = %d, want 1", len(res.BlockingRefs))
	}
	ref := res.BlockingRefs[0]
	if ref.Command != "LIST" || ref.Category != WarningCategoryListing || ref.Severity != WarningSeverityInfo {
		t.Fatalf("unexpected listing blocking ref: %+v", ref)
	}
}

func TestDeleteRangeAllowsDeletingAllReferencedLinesTogether(t *testing.T) {
	in := "10 GOTO 30\n20 PRINT \"KEEP\"\n30 PRINT \"DELETE\"\n"
	res, err := DeleteRange(in, 10, 30)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}
	if len(res.BlockingRefs) != 0 {
		t.Fatalf("DeleteRange() blocking refs = %+v, want none", res.BlockingRefs)
	}
	if res.Text != "" {
		t.Fatalf("DeleteRange() text = %q, want empty string", res.Text)
	}
	if res.DeletedLines != 3 {
		t.Fatalf("DeletedLines = %d, want 3", res.DeletedLines)
	}
}

func TestDeleteRangeReturnsOriginalTextWhenNothingMatches(t *testing.T) {
	in := "10 PRINT \"A\"\n"
	res, err := DeleteRange(in, 999, 1000)
	if err != nil {
		t.Fatalf("DeleteRange() error = %v", err)
	}
	if res.Text != in {
		t.Fatalf("DeleteRange() text = %q, want %q", res.Text, in)
	}
	if res.DeletedLines != 0 {
		t.Fatalf("DeletedLines = %d, want 0", res.DeletedLines)
	}
	if len(res.BlockingRefs) != 0 {
		t.Fatalf("blocking refs length = %d, want 0", len(res.BlockingRefs))
	}
}

func TestRenumberMSXBasicEdgeCaseComparisons(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "on error goto zero remains zero",
			in:   "10 ON ERROR GOTO 0\n20 END\n",
			want: "100 ON ERROR GOTO 0\n110 END\n",
		},
		{
			name: "resume zero remains zero",
			in:   "10 RESUME 0\n20 END\n",
			want: "100 RESUME 0\n110 END\n",
		},
		{
			name: "list comma pair and range are rewritten",
			in:   "10 LIST 100,200\n20 LIST 100-200\n100 PRINT \"A\"\n200 PRINT \"B\"\n",
			want: "100 LIST 120,130\n110 LIST 120-130\n120 PRINT \"A\"\n130 PRINT \"B\"\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Renumber(tc.in, Options{StartLine: 100, Increment: 10, FromLine: 0})
			if err != nil {
				t.Fatalf("Renumber() error = %v", err)
			}
			if res.Text != tc.want {
				t.Fatalf("Renumber() text mismatch\nwant:\n%q\ngot:\n%q", tc.want, res.Text)
			}
		})
	}
}
