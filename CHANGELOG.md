# CHANGELOG - WS7 Editor

All notable changes to this project are documented here.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added

- Added `Utilities > Word Count` command to display text statistics (words and byte count).
  - Future enhancement: tokenized byte count for MSX BASIC format.

### Changed

- Updated documentation (`README.md` and `MANUAL.md`) with `build.ps1` examples for `-NoConsole` and `-Console`, including the note that these overrides cannot be used together.

### Fixed

- Fixed top ruler alignment so it starts at text column 1 (after the line-number gutter) instead of the far-left editor edge.

## [0.1.9] - 2026-05-01

### Added

- Added `Style -> Font... (Ctrl+P,=)` dialog to choose bundled fixed-width fonts, size, weight, and italic style.
- Added persistence for editor font family, weight, size, and italic preferences.

### Changed

- Editor theme application now rebuilds with configurable font family/weight/size.
- `Style -> Bold (Ctrl+P,B)` now preserves the current italic preference while toggling bold.
- `Configure` now offers a folder browser for each external tool location and auto-detects the most likely executable/script path when found.
- Tool launch routines now consume configured paths at runtime, accepting direct file paths or directory paths with auto-detection fallback.
- `Configure` now includes a per-tool `Test` action that performs lightweight real execution probes (for example `--help`/`--version`) before saving.
- Bumped app version to `0.1.9` in `internal/version/version.go`.

## [0.1.7] - 2026-04-28

### Added

- Added `Utilities -> Configure...` in both Opening Menu and Editor Menu.
- Added editor configuration dialog with light/dark theme selection.
- Added configurable executable paths for `openMSX`, `msxbas2rom`, `BASIC Dignified`, and `MSX Encoding`.
- Added automated tests for split-view toggle rebuild behavior in `internal/ui/editor_split_view_test.go`.

### Changed

- App exit flow now checks unsaved changes across all open tabs before closing.
- Split syntax preview toggle now uses `View > Show Split Syntax Preview` / `View > Hide Split Syntax Preview`.
- Build automation now supports `build.ps1 -Run` and `build.ps1 -OpenOutputFolder`.
- Bumped app version to `0.1.7` in `internal/version/version.go`.

## [0.1.5] - 2026-04-28

### Changed

- Restored optional split editor view in `View` menu (`Show/Hide Split Syntax Preview`), with normal inline syntax highlighting preserved.
- Split view now shows plain editing on one side and live syntax-highlight preview on the other.
- Added persistence for split view preference across sessions.
- Bumped app version to `0.1.5` in `internal/version/version.go`.

### Added

- New documents now start as `untitled.asc` by default, matching the MSX-BASIC ASCII editing mode.
- `Open`, `Save As`, and file copy dialogs now use an explicit MSX source filter for `.asc` and `.amx`, while new-file suggestions default to `.asc`.
- `New` now opens a file-type selector with `MSX BASIC ASCII (*.asc)` and `MSX BASIC Tokenized/AMX (*.amx)`, with structure prepared for future `Assembly (*.asm)` and `C (*.c)`.
- Added syntax highlighting infrastructure in `internal/syntax` with an initial official MSX-BASIC lexer (keywords based on `msxWrite/msx_basic_decoder.py`) and placeholder dialect options for `MSXBAS2ROM` and `BASIC Dignified`.
- Added live visual syntax rendering in the editor through a side syntax preview panel with token colors (keywords/comments/strings/numbers/operators) and `View` toggle support.
- Opening Menu now includes a rightmost `Help` menu with Markdown viewers:
  - `README` (`HR`) opens `README.md`
  - `MANUAL` (`HM`) opens `MANUAL.md`
  - `OUTLINE` (`HO`) opens `OUTLINE.md`
- Opening Menu `Utilities` now includes `Macros` with options:
  - `Play... MP`, `Record... MR`, `Edit/Create... MD`, `Single Step... MS`, `Copy... MO`, `Delete... MY`, `Rename... ME`
- Opening Menu `Additional` now includes:
  - `Character Editor... AC`, `Hexa Editor... AH`, `Sprite Editor... AS`, `Graphos... AG`, `Noise Editor... AN`

## [0.1.0] - 2026-04-28

### Added

- Initial versioned release. Version constant introduced at `internal/version/version.go`.
- Version displayed in window title bar and in the Status dialog (`Ctrl+O,?`).
- `OUTLINE.md` and `CHANGELOG.md` for project continuity and migration support.

#### Editor Core
- Tabbed editor (`DocTabs`) supporting multiple open documents simultaneously.
- Dirty-tab indicator (`*` suffix + warning icon) on modified files.
- Duplicate-open prevention: focusing the existing tab instead of creating a duplicate.
- Tab close confirmation when unsaved changes exist.
- New untitled tab via `Ctrl+N`.
- Close current tab via `Ctrl+W`.

#### WordStar Chord Resolver
- Multi-key `Ctrl` chord resolver with prefix state and 2-second timeout.
- Supported prefixes: `Ctrl+K`, `Ctrl+K,Q`, `Ctrl+O`, `Ctrl+O,N`, `Ctrl+P`, `Ctrl+Q`, `Ctrl+Q,N`.
- Status bar feedback for every chord step and completion.

#### Navigation
- `Ctrl+S/D/E/X` — cursor left/right/up/down.
- `Ctrl+R/C` — page up/down.
- `Ctrl+Q,R` / `Ctrl+Q,C` — document start/end.

#### File Operations
- `Ctrl+K,S` — Save.
- `Ctrl+K,T` — Save As.
- `Ctrl+K,D` — Save and Close.
- `Ctrl+O,K` — Open/Switch (OS file dialog).
- `Ctrl+K,O` — Copy file.
- `Ctrl+K,J` — Delete file.
- `Ctrl+K,E` — Rename file.
- `Ctrl+K,L` — Change drive/directory.
- `Ctrl+K,F` — Run PowerShell command.
- `Ctrl+O,?` — Status (now includes version).
- `Ctrl+K,Q,X` — Exit application.

#### Editing
- `Ctrl+Y` — Delete current line.
- `Ctrl+T` — Delete word right.
- `Ctrl+Q,Y` — Delete text right of cursor.
- `Ctrl+Q,DEL` — Delete text left of cursor.
- `Ctrl+U` — Undo.

#### WS7 Internal Block Clipboard (independent of Windows clipboard)
- `Ctrl+K,B` — Mark block begin.
- `Ctrl+K,K` — Mark block end.
- `Ctrl+K,C` — Copy marked block to internal WS7 clipboard.
- `Ctrl+K,V` — Move marked block; paste from internal clipboard when no block is marked.
- `Ctrl+K,Y` — Delete marked block.
- Visual indicators in status bar: `[WS7-BLOCK:B]`, `[WS7-BLOCK:K]`, `[WS7-BLOCK:B,K]`, `[WS7-CLIP:N]`.

#### Find / Replace
- `Ctrl+Q,F` — Find dialog with options: Backward, Whole Word, Match Case, Regular Expression.
- `Ctrl+Q,A` — Find and Replace.
- `Ctrl+L` — Repeat last find.
- `Ctrl+Q,G` — Go to character offset.
- `Ctrl+Q,I` — Go to page.

#### File Browser (Opening Menu)
- Directory listing with `[DIR]` and `[..]` navigation.
- Status bar help: `Move   Enter Open   [DIR] Folder   [..] Parent`.
- Persists last-used directory via SQLite.

#### UI / Theme
- Source Code Pro Bold monospace font from bundled TTF resources.
- Ruler widget, line-number gutter, status bar per tab.

### Fixed
- All UI strings translated from Portuguese to English.
- Status bar, dialogs, menu labels and error messages are fully in English.

### Not Yet Implemented (marked `[NI]` in menus)
- Move/Copy Block from Other Window (`Ctrl+K,G` / `Ctrl+K,A`).
- Column Block Mode (`Ctrl+K,N`) and Column Replace Mode (`Ctrl+K,I`).
- Go to Previous Position (`Ctrl+Q,P`), Go to Last Find/Replace (`Ctrl+Q,V`).
- Go to Beginning/End of Block (`Ctrl+Q,B` / `Ctrl+Q,K`).
- Scroll Continuously Up/Down (`Ctrl+Q,W` / `Ctrl+Q,Z`).
- Auto Align (`Ctrl+O,A`), Edit/Convert Note (`Ctrl+O,N,D` / `Ctrl+O,N,V`).
- Printing (`Ctrl+K,P`) and Change Printer (`Ctrl+P,?`).
- Visual block highlight in the editor text area.

---

## Versioning convention

- **PATCH** (`0.1.x`): bug fixes, text corrections, minor adjustments.
- **MINOR** (`0.x.0`): new features or commands added.
- **MAJOR** (`x.0.0`): breaking changes or major redesign.

To bump the version, edit `internal/version/version.go`:

```go
const Version = "0.1.8"  // example patch bump
```

Then add a new entry at the top of this file under `## [0.1.8] - YYYY-MM-DD`.

