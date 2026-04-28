package version

import "testing"

func TestFullWithoutBuildID(t *testing.T) {
	orig := BuildID
	BuildID = ""
	t.Cleanup(func() { BuildID = orig })

	got := Full()
	want := AppName + " v" + Version
	if got != want {
		t.Fatalf("unexpected full version without build id: got=%q want=%q", got, want)
	}
}

func TestFullWithBuildID(t *testing.T) {
	orig := BuildID
	BuildID = "662fb9c1"
	t.Cleanup(func() { BuildID = orig })

	got := Full()
	want := AppName + " v" + Version + "+662fb9c1"
	if got != want {
		t.Fatalf("unexpected full version with build id: got=%q want=%q", got, want)
	}
}

func TestBuildWithoutBuildID(t *testing.T) {
	orig := BuildID
	BuildID = ""
	t.Cleanup(func() { BuildID = orig })

	if got := Build(); got != "n/a" {
		t.Fatalf("unexpected build label without build id: got=%q want=%q", got, "n/a")
	}
}

func TestBuildWithBuildID(t *testing.T) {
	orig := BuildID
	BuildID = "69f0e254"
	t.Cleanup(func() { BuildID = orig })

	if got := Build(); got != "69f0e254" {
		t.Fatalf("unexpected build label with build id: got=%q want=%q", got, "69f0e254")
	}
}

