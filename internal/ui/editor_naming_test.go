package ui

import "testing"

func TestNextUntitledNameUsesASCExtension(t *testing.T) {
	e := &editorUI{}

	if got := e.nextUntitledName(); got != "untitled.asc" {
		t.Fatalf("expected first untitled name to be untitled.asc, got %q", got)
	}
	if got := e.nextUntitledName(); got != "untitled-2.asc" {
		t.Fatalf("expected second untitled name to be untitled-2.asc, got %q", got)
	}
	if got := e.nextUntitledName(); got != "untitled-3.asc" {
		t.Fatalf("expected third untitled name to be untitled-3.asc, got %q", got)
	}
}

func TestDisplayDocumentName(t *testing.T) {
	if got := displayDocumentName("E:/wsmsx/examples/demo.asc", "untitled.asc"); got != "demo.asc" {
		t.Fatalf("expected saved file basename, got %q", got)
	}
	if got := displayDocumentName("", "untitled.asc"); got != "untitled.asc" {
		t.Fatalf("expected fallback untitled name, got %q", got)
	}
	if got := displayDocumentName("", ""); got != "[New]" {
		t.Fatalf("expected [New] when both file path and fallback are empty, got %q", got)
	}
}

func TestNormalizeMSXSourceFileName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty defaults to asc", in: "", want: "untitled.asc"},
		{name: "new marker defaults to asc", in: "[New]", want: "untitled.asc"},
		{name: "plain name gets asc", in: "demo", want: "demo.asc"},
		{name: "asc is preserved", in: "demo.asc", want: "demo.asc"},
		{name: "amx is preserved", in: "demo.amx", want: "demo.amx"},
		{name: "other extension is preserved", in: "demo.bas", want: "demo.bas"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeMSXSourceFileName(tc.in); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSuggestMSXSourceFileName(t *testing.T) {
	if got := suggestMSXSourceFileName("E:/wsmsx/examples/demo.amx", "untitled.asc"); got != "demo.amx" {
		t.Fatalf("expected saved AMX basename, got %q", got)
	}
	if got := suggestMSXSourceFileName("", "untitled"); got != "untitled.asc" {
		t.Fatalf("expected untitled fallback with asc extension, got %q", got)
	}
	if got := suggestMSXSourceFileName("", ""); got != "untitled.asc" {
		t.Fatalf("expected default untitled.asc suggestion, got %q", got)
	}
}

func TestNextUntitledNameForExtUsesIndependentCounters(t *testing.T) {
	e := &editorUI{}

	if got := e.nextUntitledNameForExt(".asc"); got != "untitled.asc" {
		t.Fatalf("expected first ASC untitled name, got %q", got)
	}
	if got := e.nextUntitledNameForExt(".amx"); got != "untitled.amx" {
		t.Fatalf("expected first AMX untitled name, got %q", got)
	}
	if got := e.nextUntitledNameForExt(".asc"); got != "untitled-2.asc" {
		t.Fatalf("expected second ASC untitled name, got %q", got)
	}
	if got := e.nextUntitledNameForExt("amx"); got != "untitled-2.amx" {
		t.Fatalf("expected second AMX untitled name with normalized ext, got %q", got)
	}
}

func TestDefaultNewFileTypeIsEnabledASCII(t *testing.T) {
	ft := defaultNewFileType()
	if !ft.Enabled {
		t.Fatal("expected default new file type to be enabled")
	}
	if ft.DefaultExt != ".asc" {
		t.Fatalf("expected default extension .asc, got %q", ft.DefaultExt)
	}
}

func TestEnabledNewFileTypesOnlyReturnsEnabledEntries(t *testing.T) {
	types := enabledNewFileTypes()
	if len(types) < 2 {
		t.Fatalf("expected at least two enabled file types, got %d", len(types))
	}
	for _, ft := range types {
		if !ft.Enabled {
			t.Fatalf("expected only enabled entries, found disabled type %q", ft.ID)
		}
	}
}
