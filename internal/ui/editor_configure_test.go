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

func TestNormalizeOpenMSXResourceName(t *testing.T) {
	if got := normalizeOpenMSXResourceName("  Panasonic_FS-A1GT.xml  "); got != "Panasonic_FS-A1GT" {
		t.Fatalf("unexpected normalized machine: got=%q", got)
	}
	if got := normalizeOpenMSXResourceName(`C:\openmsx\share\extensions\scc.xml`); got != "scc" {
		t.Fatalf("unexpected normalized extension: got=%q", got)
	}
}

func TestListOpenMSXXMLResourceNamesFromExecutableDir(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "openmsx.exe")
	if err := os.WriteFile(exe, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write exe: %v", err)
	}
	machinesDir := filepath.Join(root, "share", "machines")
	if err := os.MkdirAll(machinesDir, 0o755); err != nil {
		t.Fatalf("mkdir machines: %v", err)
	}
	if err := os.WriteFile(filepath.Join(machinesDir, "Panasonic_FS-A1GT.xml"), []byte("<machine/>"), 0o644); err != nil {
		t.Fatalf("write machine xml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(machinesDir, "README.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	got := listOpenMSXXMLResourceNames(exe, "machines")
	if len(got) != 1 || got[0] != "Panasonic_FS-A1GT" {
		t.Fatalf("unexpected machines list: %#v", got)
	}
}

func TestListOpenMSXXMLResourceNamesFromBinParentDir(t *testing.T) {
	root := t.TempDir()
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	exe := filepath.Join(binDir, "openmsx.exe")
	if err := os.WriteFile(exe, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write exe: %v", err)
	}
	extDir := filepath.Join(root, "share", "extensions")
	if err := os.MkdirAll(extDir, 0o755); err != nil {
		t.Fatalf("mkdir extensions: %v", err)
	}
	if err := os.WriteFile(filepath.Join(extDir, "scc.xml"), []byte("<extension/>"), 0o644); err != nil {
		t.Fatalf("write ext xml: %v", err)
	}

	got := listOpenMSXXMLResourceNames(exe, "extensions")
	if len(got) != 1 || got[0] != "scc" {
		t.Fatalf("unexpected extensions list: %#v", got)
	}
}

func TestBuildOpenMSXResourceOptionsIncludesCurrentWhenNotDetected(t *testing.T) {
	got := buildOpenMSXResourceOptions([]string{"Panasonic_FS-A1GT", "Philips_NMS_8250"}, "CustomMachine")
	if len(got) != 4 {
		t.Fatalf("unexpected options length: %#v", got)
	}
	if got[0] != "" {
		t.Fatalf("first option must be blank, got=%q", got[0])
	}
	if got[3] != "CustomMachine" {
		t.Fatalf("expected custom machine as last option, got=%#v", got)
	}
}

func TestBuildOpenMSXResourceOptionsNormalizesAndDeduplicates(t *testing.T) {
	got := buildOpenMSXResourceOptions([]string{"scc.xml", "SCC", "moonsound"}, `C:\openmsx\share\extensions\scc.xml`)
	if len(got) != 3 {
		t.Fatalf("unexpected options length: %#v", got)
	}
	if got[1] != "scc" || got[2] != "moonsound" {
		t.Fatalf("unexpected options: %#v", got)
	}
}

func TestNormalizeWS7BaseDirectoryFromExecutablePath(t *testing.T) {
	got := normalizeWS7BaseDirectory(`C:\apps\ws7\ws7.exe`)
	if got != `C:\apps\ws7` {
		t.Fatalf("unexpected ws7 base dir: got=%q", got)
	}
}

func TestBuildWS7SubdirectoryPaths(t *testing.T) {
	paths := buildWS7SubdirectoryPaths(`C:\apps\ws7`)
	if paths["TEMP"] != `C:\apps\ws7\TEMP` {
		t.Fatalf("unexpected TEMP path: %q", paths["TEMP"])
	}
	if paths["UTIL"] != `C:\apps\ws7\UTIL` {
		t.Fatalf("unexpected UTIL path: %q", paths["UTIL"])
	}
}

func TestCreateWS7Subdirectories(t *testing.T) {
	root := t.TempDir()
	if err := createWS7Subdirectories(root); err != nil {
		t.Fatalf("create ws7 subdirs: %v", err)
	}
	for _, name := range ws7SubdirectoryNames {
		path := filepath.Join(root, name)
		if st, err := os.Stat(path); err != nil || !st.IsDir() {
			t.Fatalf("expected directory %q to exist", path)
		}
	}
}

func TestValidateWS7BaseDirectory(t *testing.T) {
	if err := validateWS7BaseDirectory(""); err == nil {
		t.Fatal("expected error for empty ws7 directory")
	}

	if err := validateWS7BaseDirectory(`Z:\does-not-exist\ws7`); err == nil {
		t.Fatal("expected error for missing ws7 directory")
	}

	dir := t.TempDir()
	if err := validateWS7BaseDirectory(dir); err != nil {
		t.Fatalf("expected valid existing directory, got error: %v", err)
	}

	exe := filepath.Join(dir, "ws7.exe")
	if err := os.WriteFile(exe, []byte("stub"), 0o644); err != nil {
		t.Fatalf("write ws7.exe stub: %v", err)
	}
	if err := validateWS7BaseDirectory(exe); err != nil {
		t.Fatalf("expected valid ws7.exe path, got error: %v", err)
	}
}
