package input

import "testing"

func TestNormalizeShortcut(t *testing.T) {
	got, err := NormalizeShortcut("ctrl+k, s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Ctrl+K,S" {
		t.Fatalf("normalized = %q, want Ctrl+K,S", got)
	}
}

func TestNormalizeShortcutRejectsMissingCtrlPrefix(t *testing.T) {
	if _, err := NormalizeShortcut("K,S"); err == nil {
		t.Fatal("expected error for missing Ctrl+ prefix")
	}
}

func TestShortcutToResolverChord(t *testing.T) {
	keys, err := ShortcutToResolverChord("Ctrl+Q,DEL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 || keys[0] != "Q" || keys[1] != "\b" {
		t.Fatalf("resolver keys = %#v, want [Q \\b]", keys)
	}
}

func TestResolverApplyKeybindsOverridesDefault(t *testing.T) {
	r := NewResolver()
	defs := DefaultKeybindDefinitions()
	for i := range defs {
		if defs[i].ID == string(CmdSave) {
			defs[i].Shortcut = "Ctrl+K,Z"
		}
	}
	if err := r.ApplyKeybinds(defs); err != nil {
		t.Fatalf("apply keybinds: %v", err)
	}
	if chord := r.ShortcutForCommand(CmdSave); chord != "Ctrl+K,Z" {
		t.Fatalf("save shortcut = %q, want Ctrl+K,Z", chord)
	}

	cmd, pending, err := r.Resolve("K")
	if err != nil || !pending || cmd != CmdPrefixK {
		t.Fatalf("prefix K mismatch cmd=%q pending=%v err=%v", cmd, pending, err)
	}
	cmd, pending, err = r.Resolve("Z")
	if err != nil || pending {
		t.Fatalf("resolve K,Z failed: cmd=%q pending=%v err=%v", cmd, pending, err)
	}
	if cmd != CmdSave {
		t.Fatalf("command = %q, want %q", cmd, CmdSave)
	}
}
