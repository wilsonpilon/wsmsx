package msxtoken

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	goldenDebugDump = flag.Bool("msxtoken.golden.debug", false, "export side-by-side Go/Python tokenized line dump on mismatch")
	goldenDebugDir  = flag.String("msxtoken.golden.debugdir", "", "directory to write golden debug dumps")
	goldenStrict    = flag.Bool("msxtoken.golden.strict", false, "fail with inline first divergent tokenized line diff")
)

func TestGoldenParityWithPythonTokenizer(t *testing.T) {
	python, args, ok := pythonCommand()
	if !ok {
		t.Skip("python interpreter not found; skipping parity test")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	pkgDir := filepath.Dir(thisFile)
	root := filepath.Clean(filepath.Join(pkgDir, "..", "..", "..", ".."))
	pyTokenizer := filepath.Join(root, "basic-dignified", "msx", "msxbatoken", "msxbatoken.py")
	if _, err := os.Stat(pyTokenizer); err != nil {
		t.Skip("python tokenizer script not found")
	}

	fixtures, err := filepath.Glob(filepath.Join(pkgDir, "testdata", "fixtures", "*.asc"))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("failed to list fixtures: %v", err)
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			src, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			goOut, err := TokenizeProgram(string(src))
			if err != nil {
				t.Fatalf("go tokenizer failed: %v", err)
			}

			tmpDir := t.TempDir()
			pyOut := filepath.Join(tmpDir, "out.bmx")
			cmdArgs := append(append([]string{}, args...), pyTokenizer, fixture, pyOut, "-el", "0", "-vb", "0")
			cmd := exec.Command(python, cmdArgs...)
			cmd.Dir = root
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("python tokenizer failed: %v\n%s", err, string(out))
			}
			pyBytes, err := os.ReadFile(pyOut)
			if err != nil {
				t.Fatalf("read python output: %v", err)
			}
			if !bytes.Equal(goOut, pyBytes) {
				goLines := decodeTokenizedLines(goOut)
				pyLines := decodeTokenizedLines(pyBytes)
				dumpPath := ""

				// Strict mode now always includes both: inline first-line diff + full dump file.
				if *goldenStrict {
					if p, dumpErr := writeDebugDump(fixture, goOut, pyBytes); dumpErr == nil {
						dumpPath = p
					} else {
						t.Logf("debug dump failed: %v", dumpErr)
					}
					lineIdx, goLine, pyLine := firstDiffLine(goLines, pyLines)
					goHex := truncateHex(hex.EncodeToString(goLine.record), 72)
					pyHex := truncateHex(hex.EncodeToString(pyLine.record), 72)
					if dumpPath != "" {
						t.Fatalf("tokenized line mismatch at idx=%d go(line=%d addr=%04x) py(line=%d addr=%04x)\nGO: %s\nPY: %s\nfull dump: %s", lineIdx+1, goLine.line, goLine.addr, pyLine.line, pyLine.addr, goHex, pyHex, dumpPath)
					}
					t.Fatalf("tokenized line mismatch at idx=%d go(line=%d addr=%04x) py(line=%d addr=%04x)\nGO: %s\nPY: %s", lineIdx+1, goLine.line, goLine.addr, pyLine.line, pyLine.addr, goHex, pyHex)
				}

				if *goldenDebugDump {
					if p, dumpErr := writeDebugDump(fixture, goOut, pyBytes); dumpErr == nil {
						dumpPath = p
					} else {
						t.Logf("debug dump failed: %v", dumpErr)
					}
				}
				diffAt := firstDiffOffset(goOut, pyBytes)
				if dumpPath != "" {
					t.Fatalf("tokenized output mismatch at byte %d: go=%d bytes python=%d bytes (debug dump: %s)", diffAt, len(goOut), len(pyBytes), dumpPath)
				}
				t.Fatalf("tokenized output mismatch at byte %d: go=%d bytes python=%d bytes (rerun with -args -msxtoken.golden.strict or -msxtoken.golden.debug)", diffAt, len(goOut), len(pyBytes))
			}
		})
	}
}

func writeDebugDump(fixture string, goOut, pyOut []byte) (string, error) {
	outDir := strings.TrimSpace(*goldenDebugDir)
	if outDir == "" {
		outDir = filepath.Join(os.TempDir(), "ws7-msxtoken-golden")
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(fixture), filepath.Ext(fixture))
	path := filepath.Join(outDir, base+"_golden_diff.txt")

	goLines := decodeTokenizedLines(goOut)
	pyLines := decodeTokenizedLines(pyOut)
	maxLines := len(goLines)
	if len(pyLines) > maxLines {
		maxLines = len(pyLines)
	}

	var b strings.Builder
	b.WriteString("MSX Tokenizer Golden Diff\n")
	b.WriteString(fmt.Sprintf("Fixture: %s\n", fixture))
	b.WriteString(fmt.Sprintf("Go bytes: %d\nPython bytes: %d\n", len(goOut), len(pyOut)))
	b.WriteString(fmt.Sprintf("First diff byte: %d\n", firstDiffOffset(goOut, pyOut)))
	b.WriteString("\n")
	b.WriteString("#  Addr   Line  GoHex                                   PythonHex                               Match\n")
	b.WriteString("----------------------------------------------------------------------------------------------------\n")

	for i := 0; i < maxLines; i++ {
		var g, p tokenizedLine
		if i < len(goLines) {
			g = goLines[i]
		}
		if i < len(pyLines) {
			p = pyLines[i]
		}
		gHex := hex.EncodeToString(g.record)
		pHex := hex.EncodeToString(p.record)
		match := gHex == pHex
		b.WriteString(fmt.Sprintf("%02d %04x   %5d %-40s %-40s %v\n", i+1, g.addr, g.line, truncateHex(gHex, 40), truncateHex(pHex, 40), match))
		if !match {
			b.WriteString(fmt.Sprintf("    GO: %s\n", gHex))
			b.WriteString(fmt.Sprintf("    PY: %s\n", pHex))
		}
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

type tokenizedLine struct {
	addr   uint16
	line   uint16
	record []byte
}

func decodeTokenizedLines(data []byte) []tokenizedLine {
	lines := make([]tokenizedLine, 0, 32)
	if len(data) == 0 {
		return lines
	}
	i := 0
	if data[0] == 0xff {
		i = 1
	}
	currAddr := uint16(baseAddress)
	for i+1 < len(data) {
		if data[i] == 0x00 && data[i+1] == 0x00 {
			break
		}
		if i+4 > len(data) {
			break
		}
		next := uint16(data[i]) | uint16(data[i+1])<<8
		line := uint16(data[i+2]) | uint16(data[i+3])<<8
		recStart := i
		i += 4
		for i < len(data) && data[i] != 0x00 {
			i++
		}
		if i < len(data) {
			i++
		}
		rec := append([]byte(nil), data[recStart:i]...)
		lines = append(lines, tokenizedLine{addr: currAddr, line: line, record: rec})
		currAddr = next
	}
	return lines
}

func firstDiffLine(goLines, pyLines []tokenizedLine) (int, tokenizedLine, tokenizedLine) {
	maxLines := len(goLines)
	if len(pyLines) > maxLines {
		maxLines = len(pyLines)
	}
	for i := 0; i < maxLines; i++ {
		var g, p tokenizedLine
		if i < len(goLines) {
			g = goLines[i]
		}
		if i < len(pyLines) {
			p = pyLines[i]
		}
		if !bytes.Equal(g.record, p.record) {
			return i, g, p
		}
	}
	return -1, tokenizedLine{}, tokenizedLine{}
}

func truncateHex(v string, max int) string {
	if len(v) <= max {
		return v
	}
	if max <= 3 {
		return v[:max]
	}
	return v[:max-3] + "..."
}

func firstDiffOffset(a, b []byte) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	if len(a) != len(b) {
		return minLen
	}
	return -1
}

func pythonCommand() (string, []string, bool) {
	if p, err := exec.LookPath("python"); err == nil {
		return p, nil, true
	}
	if p, err := exec.LookPath("python3"); err == nil {
		return p, nil, true
	}
	if p, err := exec.LookPath("py"); err == nil {
		return p, []string{"-3"}, true
	}
	return "", nil, false
}
