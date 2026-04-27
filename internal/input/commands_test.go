package input

import "testing"

func TestResolveSingleCommand(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdCursorLeft {
		t.Fatalf("expected %s, got %s", CmdCursorLeft, cmd)
	}
}

func TestResolveCtrlNNewTab(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("N")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdNewTab {
		t.Fatalf("expected %s, got %s", CmdNewTab, cmd)
	}
}

func TestResolveCtrlWClose(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("W")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdClose {
		t.Fatalf("expected %s, got %s", CmdClose, cmd)
	}
}

func TestResolveChordCtrlKCtrlS(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("K")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixK || !pending {
		t.Fatalf("Ctrl+K prefix should remain pending")
	}

	cmd, pending, err = r.Resolve("S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdSave {
		t.Fatalf("expected %s, got %s", CmdSave, cmd)
	}
}

func TestResolveChordCtrlOCtrlK(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("O")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixO || !pending {
		t.Fatalf("Ctrl+O prefix should remain pending")
	}

	cmd, pending, err = r.Resolve("K")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdOpenSwitch {
		t.Fatalf("expected %s, got %s", CmdOpenSwitch, cmd)
	}
}

func TestResolveChordCtrlKCtrlQCtrlX(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("K")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixK || !pending {
		t.Fatalf("Ctrl+K prefix should remain pending")
	}

	cmd, pending, err = r.Resolve("Q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixKQ || !pending {
		t.Fatalf("Ctrl+K,Q prefix should remain pending")
	}

	cmd, pending, err = r.Resolve("X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdExit {
		t.Fatalf("expected %s, got %s", CmdExit, cmd)
	}
}

func TestResolveChordCtrlPCtrlQuestion(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixP || !pending {
		t.Fatalf("Ctrl+P prefix should remain pending")
	}

	cmd, pending, err = r.Resolve("?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdChangePrinter {
		t.Fatalf("expected %s, got %s", CmdChangePrinter, cmd)
	}
}

func TestResolveUnsupportedAfterPrefix(t *testing.T) {
	r := NewResolver()
	_, _, _ = r.Resolve("K")
	// "$" is not a valid second key after Ctrl+K
	_, _, err := r.Resolve("$")
	if err == nil {
		t.Fatal("expected an error for unsupported command")
	}
	if r.HasPrefix() {
		t.Fatal("prefix should be cleared after error")
	}
}

func TestResolveCtrlQPrefix(t *testing.T) {
	r := NewResolver()
	cmd, pending, err := r.Resolve("Q")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixQ || !pending {
		t.Fatalf("Ctrl+Q prefix should remain pending, got cmd=%s pending=%v", cmd, pending)
	}
	cmd, pending, err = r.Resolve("F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pending {
		t.Fatalf("did not expect pending=true")
	}
	if cmd != CmdFind {
		t.Fatalf("expected %s, got %s", CmdFind, cmd)
	}
}

func TestResolveCtrlKBlock(t *testing.T) {
	tests := []struct {
		second string
		want   Command
	}{
		{"B", CmdMarkBlockBegin},
		{"K", CmdMarkBlockEnd},
		{"V", CmdMoveBlock},
		{"C", CmdCopyBlock},
		{"A", CmdCopyBlockOtherWin},
		{"Y", CmdDeleteBlock},
		{"U", CmdMarkPreviousBlock},
		{"N", CmdColumnBlockMode},
		{"I", CmdColumnReplaceMode},
	}
	for _, tt := range tests {
		r := NewResolver()
		r.Resolve("K") //nolint
		cmd, _, err := r.Resolve(tt.second)
		if err != nil {
			t.Fatalf("K,%s: unexpected error: %v", tt.second, err)
		}
		if cmd != tt.want {
			t.Fatalf("K,%s: expected %s, got %s", tt.second, tt.want, cmd)
		}
	}
}

func TestResolveCtrlQNav(t *testing.T) {
	tests := []struct {
		second string
		want   Command
	}{
		{"A", CmdFindReplace},
		{"G", CmdGoToChar},
		{"I", CmdGoToPage},
		{"P", CmdGoPrevPosition},
		{"V", CmdGoLastFindReplace},
		{"B", CmdGoBlockBegin},
		{"K", CmdGoBlockEnd},
		{"R", CmdGoDocBegin},
		{"C", CmdGoDocEnd},
		{"W", CmdScrollContUp},
		{"Z", CmdScrollContDown},
		{"Y", CmdDeleteLineRight},
	}
	for _, tt := range tests {
		r := NewResolver()
		r.Resolve("Q") //nolint
		cmd, _, err := r.Resolve(tt.second)
		if err != nil {
			t.Fatalf("Q,%s: unexpected error: %v", tt.second, err)
		}
		if cmd != tt.want {
			t.Fatalf("Q,%s: expected %s, got %s", tt.second, tt.want, cmd)
		}
	}
}

func TestResolveMarkers(t *testing.T) {
	// Set marker
	r := NewResolver()
	r.Resolve("K") //nolint
	cmd, _, err := r.Resolve("5")
	if err != nil {
		t.Fatalf("set marker 5: unexpected error: %v", err)
	}
	if d, ok := IsSetMarker(cmd); !ok || d != "5" {
		t.Fatalf("expected set_marker_5, got %s", cmd)
	}

	// Go to marker
	r.Resolve("Q") //nolint
	cmd, _, err = r.Resolve("3")
	if err != nil {
		t.Fatalf("go marker 3: unexpected error: %v", err)
	}
	if d, ok := IsGoToMarker(cmd); !ok || d != "3" {
		t.Fatalf("expected go_to_marker_3, got %s", cmd)
	}
}

func TestResolveCtrlONPrefix(t *testing.T) {
	r := NewResolver()
	r.Resolve("O") //nolint
	cmd, pending, err := r.Resolve("N")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdPrefixON || !pending {
		t.Fatalf("O,N prefix should remain pending")
	}
	cmd, _, err = r.Resolve("D")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != CmdEditNote {
		t.Fatalf("expected %s, got %s", CmdEditNote, cmd)
	}
}
