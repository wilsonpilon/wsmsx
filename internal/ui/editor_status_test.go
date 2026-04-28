package ui

import (
	"strings"
	"testing"

	"ws7/internal/version"
)

func TestVersionBuildTraceText(t *testing.T) {
	orig := version.BuildID
	version.BuildID = "69f0e254"
	t.Cleanup(func() { version.BuildID = orig })

	got := versionBuildTraceText()
	if !strings.Contains(got, "Version: ") {
		t.Fatalf("expected version prefix in trace text, got %q", got)
	}
	if !strings.Contains(got, "Build: 69f0e254") {
		t.Fatalf("expected build id in trace text, got %q", got)
	}
}

