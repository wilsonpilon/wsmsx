package version

// Version is the current application version.
// Bump this before each release following semver: MAJOR.MINOR.PATCH
const Version = "0.1.0"

// AppName is the canonical application name used in titles and dialogs.
const AppName = "WS7"

// Full returns the full display string, e.g. "WS7 v0.1.0".
func Full() string { return AppName + " v" + Version }

