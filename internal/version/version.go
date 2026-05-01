package version

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Version is the current application version.
// Bump this before each release following semver: MAJOR.MINOR.PATCH
const Version = "0.1.9"

// AppName is the canonical application name used in titles and dialogs.
const AppName = "WS7"

// BuildID is injected at compile time (Unix seconds in hexadecimal).
// Example: go build -ldflags "-X ws7/internal/version.BuildID=662fb9c1"
var BuildID = ""

func init() {
	if strings.TrimSpace(BuildID) != "" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if s.Key != "vcs.time" || strings.TrimSpace(s.Value) == "" {
				continue
			}
			if ts, err := time.Parse(time.RFC3339, s.Value); err == nil {
				BuildID = strconv.FormatInt(ts.Unix(), 16)
				return
			}
		}
	}
	if exe, err := os.Executable(); err == nil {
		if stat, statErr := os.Stat(filepath.Clean(exe)); statErr == nil {
			BuildID = strconv.FormatInt(stat.ModTime().Unix(), 16)
		}
	}
}

// Full returns the full display string, e.g. "WS7 v0.1.0+662fb9c1".
func Full() string {
	full := AppName + " v" + Version
	if strings.TrimSpace(BuildID) != "" {
		full += "+" + strings.TrimSpace(BuildID)
	}
	return full
}

// Build returns only the build identifier for support/rastreio screens.
// When no build was injected, returns "n/a".
func Build() string {
	if id := strings.TrimSpace(BuildID); id != "" {
		return id
	}
	return "n/a"
}
