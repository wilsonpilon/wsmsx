package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigureInitialDirectoryPrefersExistingDirectory(t *testing.T) {
	tmp := t.TempDir()
	got := configureInitialDirectory(tmp, "")
	if got != tmp {
		t.Fatalf("unexpected directory: got=%q want=%q", got, tmp)
	}
}

func TestConfigureInitialDirectoryUsesParentOfLegacyExecutablePath(t *testing.T) {
	tmp := t.TempDir()
	legacyPath := filepath.Join(tmp, "openmsx.exe")
	got := configureInitialDirectory(legacyPath, "")
	if got != tmp {
		t.Fatalf("unexpected parent directory: got=%q want=%q", got, tmp)
	}
}

func TestConfigureInitialDirectoryFallsBackToLastDir(t *testing.T) {
	fallback := t.TempDir()
	got := configureInitialDirectory("Z:\\does-not-exist\\tool.exe", fallback)
	if got != fallback {
		t.Fatalf("unexpected fallback directory: got=%q want=%q", got, fallback)
	}
}

func TestDetectConfiguredToolPathFindsOpenMSXExecutable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "openmsx.exe")
	if err := os.WriteFile(path, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	got := detectConfiguredToolPath(settingOpenMSXExeKey, dir)
	if got != path {
		t.Fatalf("unexpected detected path: got=%q want=%q", got, path)
	}
}

func TestDetectConfiguredToolPathFindsBasicDignifiedScript(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badig.py")
	if err := os.WriteFile(path, []byte("print('stub')"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	got := detectConfiguredToolPath(settingBasicDignifiedExeKey, dir)
	if got != path {
		t.Fatalf("unexpected detected path: got=%q want=%q", got, path)
	}
}

func TestDetectConfiguredToolPathFindsMSXEncodingDistEntry(t *testing.T) {
	dir := t.TempDir()
	dist := filepath.Join(dir, "dist")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		t.Fatalf("mkdir dist: %v", err)
	}
	path := filepath.Join(dist, "extension.js")
	if err := os.WriteFile(path, []byte("// stub"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	got := detectConfiguredToolPath(settingMSXEncodingExeKey, dir)
	if got != path {
		t.Fatalf("unexpected detected path: got=%q want=%q", got, path)
	}
}

func TestDetectConfiguredToolPathFallsBackToSelectedDirectory(t *testing.T) {
	dir := t.TempDir()
	got := detectConfiguredToolPath(settingMSXBas2RomExeKey, dir)
	if got != dir {
		t.Fatalf("unexpected fallback path: got=%q want=%q", got, dir)
	}
}

func TestResolveConfiguredToolPathUsesExactFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "openmsx.exe")
	if err := os.WriteFile(file, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	got, err := resolveConfiguredToolPath(settingOpenMSXExeKey, file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != file {
		t.Fatalf("unexpected path: got=%q want=%q", got, file)
	}
}

func TestResolveConfiguredToolPathDetectsFromDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "msxbas2rom.exe")
	if err := os.WriteFile(file, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	got, err := resolveConfiguredToolPath(settingMSXBas2RomExeKey, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != file {
		t.Fatalf("unexpected path: got=%q want=%q", got, file)
	}
}

func TestResolveConfiguredToolPathFailsOnUndetectedDirectory(t *testing.T) {
	dir := t.TempDir()
	if _, err := resolveConfiguredToolPath(settingMSXBas2RomExeKey, dir); err == nil {
		t.Fatal("expected error for directory without detected executable")
	}
}

func TestResolveConfiguredToolPathUsesParentFallbackForMissingLegacyFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "openmsx.exe")
	if err := os.WriteFile(file, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	legacyMissing := filepath.Join(dir, "old-openmsx.exe")
	got, err := resolveConfiguredToolPath(settingOpenMSXExeKey, legacyMissing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != file {
		t.Fatalf("unexpected fallback path: got=%q want=%q", got, file)
	}
}

func TestBuildConfiguredToolCommandForPython(t *testing.T) {
	name, args, workDir := buildConfiguredToolCommand(settingBasicDignifiedExeKey, `C:\tools\badig.py`, []string{"input.asc"})
	if name != "python" {
		t.Fatalf("unexpected command: %q", name)
	}
	if len(args) != 3 || args[0] != "-u" || args[1] != `C:\tools\badig.py` || args[2] != "input.asc" {
		t.Fatalf("unexpected args: %#v", args)
	}
	if workDir != `C:\tools` {
		t.Fatalf("unexpected workdir: %q", workDir)
	}
}

func TestBuildConfiguredToolCommandForNodeJS(t *testing.T) {
	name, args, workDir := buildConfiguredToolCommand(settingMSXEncodingExeKey, `C:\tools\dist\extension.js`, nil)
	if name != "node" {
		t.Fatalf("unexpected command: %q", name)
	}
	if len(args) != 1 || args[0] != `C:\tools\dist\extension.js` {
		t.Fatalf("unexpected args: %#v", args)
	}
	if workDir != `C:\tools\dist` {
		t.Fatalf("unexpected workdir: %q", workDir)
	}
}

func TestBuildConfiguredToolCommandForMSXEncodingPackageJSON(t *testing.T) {
	name, args, workDir := buildConfiguredToolCommand(settingMSXEncodingExeKey, `C:\tools\msx-encoding\package.json`, nil)
	if name != "npm" {
		t.Fatalf("unexpected command: %q", name)
	}
	if len(args) != 4 || args[0] != "--prefix" || args[1] != `C:\tools\msx-encoding` || args[2] != "run" || args[3] != "compile" {
		t.Fatalf("unexpected args: %#v", args)
	}
	if workDir != `C:\tools\msx-encoding` {
		t.Fatalf("unexpected workdir: %q", workDir)
	}
}

func TestBuildConfiguredToolProbeSpecsForOpenMSX(t *testing.T) {
	specs := buildConfiguredToolProbeSpecs(settingOpenMSXExeKey, `C:\tools\openmsx.exe`)
	if len(specs) < 1 {
		t.Fatal("expected at least one probe spec")
	}
	if specs[0].name != `C:\tools\openmsx.exe` {
		t.Fatalf("unexpected probe command: %q", specs[0].name)
	}
	if len(specs[0].args) < 1 || specs[0].args[0] != "--version" {
		t.Fatalf("unexpected probe args: %#v", specs[0].args)
	}
}

func TestBuildConfiguredToolProbeSpecsForPythonScript(t *testing.T) {
	specs := buildConfiguredToolProbeSpecs(settingBasicDignifiedExeKey, `C:\tools\badig.py`)
	if len(specs) < 1 {
		t.Fatal("expected at least one probe spec")
	}
	if specs[0].name != "python" {
		t.Fatalf("unexpected probe command: %q", specs[0].name)
	}
	if len(specs[0].args) < 3 || specs[0].args[0] != "-u" || specs[0].args[1] != `C:\tools\badig.py` {
		t.Fatalf("unexpected probe args: %#v", specs[0].args)
	}
}

func TestBuildConfiguredToolProbeSpecsForMSXEncodingPackageJSON(t *testing.T) {
	specs := buildConfiguredToolProbeSpecs(settingMSXEncodingExeKey, `C:\tools\msx-encoding\package.json`)
	if len(specs) != 1 {
		t.Fatalf("expected 1 probe spec, got %d", len(specs))
	}
	if specs[0].name != "npm" {
		t.Fatalf("unexpected probe command: %q", specs[0].name)
	}
	if len(specs[0].args) != 3 || specs[0].args[0] != "--prefix" || specs[0].args[1] != `C:\tools\msx-encoding` || specs[0].args[2] != "--version" {
		t.Fatalf("unexpected probe args: %#v", specs[0].args)
	}
}
