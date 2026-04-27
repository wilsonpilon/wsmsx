# WS7 Editor

Text editor in Go + Fyne, inspired by the WordStar 7.0 workflow, focused on MSX-BASIC development.

## Motivation and Inspiration

- Recreate a classic editing experience centered on keyboard-driven productivity.
- Preserve the WordStar 7 style `Ctrl` prefix command logic.
- Provide a modern environment for building software for the MSX ecosystem.

## Project Goals

- Deliver a lightweight editor/IDE for writing and organizing MSX-BASIC code.
- Maintain high interaction fidelity with WordStar before adding extra features.
- Persist settings and usage context (recent files and directories) for a continuous workflow.

## Technologies and Tools

- **Go**: main application language.
- **Fyne**: desktop GUI framework.
- **SQLite**: local settings and history storage.
- **PowerShell** (`build.ps1`): Windows build automation.
- **Go test / go build**: continuous validation of changes.

## Recent Changes

- Tabbed editor (`DocTabs`) for multiple documents.
- Strong visual dirty-tab indicator (`*` + warning icon).
- Duplicate open prevention for the same file (focuses existing tab).
- Tab close confirmation when unsaved changes exist.
- Updated shortcut map:
  - `Ctrl+N` opens a new tab (`untitled`).
  - `Ctrl+W` closes the current tab.
  - `Ctrl+D` moves the cursor right again.

## Main Structure

```text
cmd/ws7/main.go                  application entry point
internal/ui/editor.go            global state, screens, menus, and tabs
internal/ui/filebrowser.go       file navigation (Opening Menu)
internal/ui/theme.go             Source Code Pro theme
internal/input/commands.go       Ctrl/WordStar command resolver
internal/store/sqlite/store.go   SQLite (settings, projects, recent_files)
internal/config/paths.go         local data paths
res/                             TTF fonts and wordstar7.pdf manual
build.ps1                        Windows build
```

## Usage Documentation

- Full operational guide: `MANUAL.md`

## Quick Run

```bash
go mod tidy
go run ./cmd/ws7
```

## Build (Windows)

```powershell
./build.ps1
./build.ps1 -Configuration Release
./build.ps1 -Output dist/ws7.exe -SkipTests
```

## Tests

```bash
go test ./...
```
