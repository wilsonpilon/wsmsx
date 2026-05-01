package ui

import (
	"strings"
	"testing"
)

func TestBuildOpenMSXBridgeCommandAddsControlStdio(t *testing.T) {
	name, args, workDir := buildOpenMSXBridgeCommand(`C:\tools\openmsx.exe`, "", nil)
	if name != `C:\tools\openmsx.exe` {
		t.Fatalf("name = %q, want %q", name, `C:\tools\openmsx.exe`)
	}
	if len(args) != 2 || args[0] != "-control" || args[1] != "stdio" {
		t.Fatalf("args = %#v, want [-control stdio]", args)
	}
	if workDir != `C:\tools` {
		t.Fatalf("workDir = %q, want %q", workDir, `C:\tools`)
	}
}

func TestBuildOpenMSXBridgeCommandAddsMachineAndExtensions(t *testing.T) {
	name, args, workDir := buildOpenMSXBridgeCommand(`C:\tools\openmsx.exe`, "Panasonic_FS-A1GT", []string{"scc", "moonsound"})
	if name != `C:\tools\openmsx.exe` {
		t.Fatalf("name = %q, want %q", name, `C:\tools\openmsx.exe`)
	}
	want := []string{"-machine", "Panasonic_FS-A1GT", "-ext", "scc", "-ext", "moonsound", "-control", "stdio"}
	if len(args) != len(want) {
		t.Fatalf("args len = %d, want %d (%#v)", len(args), len(want), args)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q (all=%#v)", i, args[i], want[i], args)
		}
	}
	if workDir != `C:\tools` {
		t.Fatalf("workDir = %q, want %q", workDir, `C:\tools`)
	}
}

func TestBuildOpenMSXCommandXML(t *testing.T) {
	got := buildOpenMSXCommandXML(`set title "A&B<1>"`)
	want := `<command>set title &#34;A&amp;B&lt;1&gt;&#34;</command>`
	if got != want {
		t.Fatalf("xml command = %q, want %q", got, want)
	}
}

func TestScanOpenMSXOutput(t *testing.T) {
	input := "line1\nline2\n"
	var lines []string
	scanOpenMSXOutput(strings.NewReader(input), "[out] ", func(line string) {
		lines = append(lines, line)
	})
	if len(lines) != 2 {
		t.Fatalf("line count = %d, want 2", len(lines))
	}
	if lines[0] != "[out] line1" || lines[1] != "[out] line2" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}

func TestOpenMSXBootCommands(t *testing.T) {
	got := openMSXBootCommands()
	if len(got) != 2 {
		t.Fatalf("boot command count = %d, want 2", len(got))
	}
	if got[0] != "set renderer sdlgl-pp" {
		t.Fatalf("boot command[0] = %q, want %q", got[0], "set renderer sdlgl-pp")
	}
	if got[1] != "set power on" {
		t.Fatalf("boot command[1] = %q, want %q", got[1], "set power on")
	}
}
