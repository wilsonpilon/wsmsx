# CHANGELOG - WS7 Editor

All notable changes to this project are documented here.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/).

---

## [Unreleased]

### Added

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
- Ruler widget, line numbers gutter, status bar per tab.

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
const Version = "0.1.1"  // example patch bump
```

Then add a new entry at the top of this file under `## [0.1.1] - YYYY-MM-DD`.

