# WS7 Manual

Usage guide for the WS7 editor (Go + Fyne), with commands inspired by WordStar.

## 1) How to Start the Program

### Run in Development

```bash
go mod tidy
go run ./cmd/ws7
```

### Build an Executable on Windows

```powershell
./build.ps1
./build.ps1 -Configuration Release
```

## 2) Startup Screen (Opening Menu)

When WS7 starts, the startup screen lists directories and files.

- Use arrow keys to navigate the list.
- Press `Enter` to open a file or enter a directory.
- `[..]` moves up one level.
- Mouse click also selects/opens.

Top menu items on the startup screen:

- `File`
- `Utilities`
- `Additional`
- `Help`

Startup menu details (current):

- `Utilities > Macros`
  - `Play... MP`, `Record... MR`, `Edit/Create... MD`, `Single Step... MS`, `Copy... MO`, `Delete... MY`, `Rename... ME`
- `Utilities > Configure...`
  - Choose editor theme (`Dark` / `Light`).
  - Configure executable paths for `openMSX`, `msxbas2rom`, `BASIC Dignified`, and `MSX Encoding`.
- `Additional`
  - `Character Editor... AC`, `Hexa Editor... AH`, `Sprite Editor... AS`, `Graphos... AG`, `Noise Editor... AN`
- `Help`
  - `README HR`, `MANUAL HM`, `OUTLINE HO`
  - These entries open project documents rendered as Markdown.

## 3) Open a File

You can open a file in two ways:

1. From the startup screen by selecting an item and pressing `Enter`.
2. In the editor via `Ctrl+O` `Ctrl+K` (Open/Switch).

If the file is already open in another tab, WS7 focuses the existing tab (no duplicate tab is created).

## 4) Editor Operation

When you enter the editor, the top menu changes to:

- `File`
- `Edit`
- `Insert`
- `Style`
- `Utilities`

Current editor features:

- Tab-based editing (multiple documents).
- `New` opens a type selector for new source files.
- Type options currently available:
  - `MSX BASIC ASCII (*.asc)`
  - `MSX BASIC Tokenized/AMX (*.amx)`
- New tab names follow the chosen type, e.g. `untitled.asc` or `untitled.amx`.
- `View > Syntax` exposes BASIC dialect options for highlighting.
  - Current active highlighter: `MSX-BASIC Official`.
  - `MSXBAS2ROM` and `BASIC Dignified` are listed as future options.
- Editor supports optional split syntax preview beside the text editor.
  - Toggle it in `View > Show Split Syntax Preview` / `View > Hide Split Syntax Preview`.
  - Inline syntax highlighting remains active in normal (non-split) mode.
- **Style menu** (`Style`):
  - `Bold (Ctrl+P,B)`: Toggle bold rendering on current tab; gutter and ruler adapt to font metrics.
  - `Font... (Ctrl+P,=)`: Open font configuration dialog.
    - Choose font family (Source Code Pro, MSX Screen).
    - Select size (8–48 pt), weight (ExtraLight–Black), and italic style.
    - Options are persisted across sessions.
- `Insert menu`:
  - `Include File (Ctrl+K,R)`: Insert external file at cursor position.
  - `Extended Character (Ctrl+M,G)`: Insert non-ASCII characters.
  - `Convert Case` submenu:
    - `Uppercase (Ctrl+K,")`: Convert to UPPERCASE.
    - `Lowercase (Ctrl+K,')`: Convert to lowercase.
    - `Capitalize (Ctrl+K,.)`: Capitalize First Letter Of Each Word.
- `Utilities > Configure...` is also available inside the editor with the same settings.
  - Folder browser for each external tool.
  - Auto-detection of tool executable/script when folder is selected.
  - `Test` button per tool to pre-validate path (lightweight probe execution).
- `Utilities > RULE (Regua)` opens the floating ruler overlay.
  - The fixed top ruler starts at text column 1 (after the line-number gutter).
  - It uses a fixed 132-column visual scale.
  - It is draggable on screen.
  - It tracks live character distance from the cursor position where RULE was enabled.
  - While RULE is active, `B` marks block start / block end for inclusive span counting.
- `Utilities > Calculator` opens an expression calculator dialog.
  - Shortcut: `Ctrl+Q` `Ctrl+M`.
  - Input supports decimal by default, plus `&H` (hex) and `&B` (binary) prefixes.
  - Output shows Decimal / Hex / Binary.
- `Utilities > Word Count` displays text statistics:
  - Number of words detected in the active editor.
  - Total character count (bytes).
  - (Future: tokenized byte count for MSX BASIC).
- `Utilities > Open openMSX`: Launch MSX emulator (detached).
- `Utilities > Run msxbas2rom`: Convert MSX BASIC files.
- `Utilities > Run BASIC Dignified`: Transpile BASIC Dignified syntax.
- `Utilities > Run MSX Encoding`: Handle MSX text encoding.
- Tab close confirmation when unsaved changes exist.
- Global exit confirmation now checks all open tabs for unsaved changes before closing the app.
- Dirty tab indicator with `*` and warning icon.

## 5) Main Shortcuts

### Basic Navigation and Editing

| Key      | Action |
|----------|--------|
| `Ctrl+S` | Move cursor left |
| `Ctrl+D` | Move cursor right |
| `Ctrl+E` | Move cursor up |
| `Ctrl+X` | Move cursor down |
| `Ctrl+R` | Page up |
| `Ctrl+C` | Page down |
| `Ctrl+Y` | Delete line |
| `Ctrl+T` | Delete word right |
| `Ctrl+U` | Undo |
| `Ctrl+N` | New tab (opens type selector; default `*.asc`) |
| `Ctrl+W` | Close current tab |
| `Ctrl+O` `Ctrl+L` | Document beginning |

### File Commands (WordStar)

| Key | Action |
|-----|--------|
| `Ctrl+K` `Ctrl+S` | Save |
| `Ctrl+K` `Ctrl+T` | Save As |
| `Ctrl+K` `Ctrl+D` | Save and Close |
| `Ctrl+O` `Ctrl+K` | Open/Switch |
| `Ctrl+O` `Ctrl+?` | Status |
| `Ctrl+P` `Ctrl+?` | Change Printer |
| `Ctrl+K` `Ctrl+Q` `Ctrl+X` | Exit |

### Style Commands

| Key | Action |
|-----|--------|
| `Ctrl+P` `Ctrl+B` | Toggle bold text style |
| `Ctrl+P` `Ctrl+=` | Open font configuration (family/size/weight/italic) |

### Text Operations

| Key | Action |
|-----|--------|
| `Ctrl+K` `Ctrl+R` | Include file at cursor |
| `Ctrl+K` `Ctrl+"` | Convert selection/line to UPPERCASE |
| `Ctrl+K` `Ctrl+'` | Convert selection/line to lowercase |
| `Ctrl+K` `Ctrl+.` | Capitalize selection/line |

### RULE Mode

| Key | Action |
|-----|--------|
| `Ctrl+Q` `Ctrl+R` | Toggle floating ruler on/off |
| `ESC` | Exit RULE mode |
| `B` | Mark block start / block end while RULE is active |

### Calculator Mode

| Key | Action |
|-----|--------|
| `Ctrl+Q` `Ctrl+M` | Open calculator dialog |

### Tool Launch

| Menu Action | Purpose |
|-------------|---------|
| `Utilities > Open openMSX` | Launch MSX emulator |
| `Utilities > Run msxbas2rom` | Convert MSX BASIC (with current file as arg) |
| `Utilities > Run BASIC Dignified` | Transpile BASIC Dignified (with current file as arg) |
| `Utilities > Run MSX Encoding` | Run MSX encoding tool (with current file as arg) |

## 6) Font Configuration

Font settings are available via `Style > Font... (Ctrl+P,=)`.

Dialog options:

- **Font Family**: Source Code Pro (default), MSX Screen 0, MSX Screen 1.
- **Size**: 8–48 pt (default: 14 pt).
- **Weight**: ExtraLight, Light, Regular, Medium, SemiBold, Bold, ExtraBold, Black.
  - Only applicable to Source Code Pro; MSX Screen fonts use Regular.
- **Style**: Italic checkbox (enabled only for Source Code Pro).
- **Width**: Currently shows "Normal" (Narrow widths are not available in bundled fonts).

Settings are automatically persisted across sessions.

The line-number gutter, floating ruler, and column ruler adapt to the selected font family/size/weight.

## 7) Configure Dialog

Access via `Utilities > Configure...` (in both Opening Menu and Editor).

Configuration items:

- **Editor Theme**: Dark, Light, One Dark, Monokai, Solarized Dark, Github Dark.
- **Tool paths** (one entry per external tool):
  - openMSX
  - msxbas2rom
  - BASIC Dignified
  - MSX Encoding

Each tool path field includes:

- **Browse button**: Opens a folder selector. Auto-detects the most likely executable/script inside the selected folder.
- **Test button**: Pre-validates the configured path by executing a lightweight probe command:
  - `--help` or `--version` for standalone executables.
  - `python -u script --help` for Python scripts.
  - `node --check script.js` for Node.js files.
  - `npm --prefix <dir> --version` for `package.json` (MSX Encoding).
- **Manual entry**: You can type or paste a direct file path; it is accepted as-is.

The Test button displays:

- Resolved file path
- Probe command used
- Work directory
- Execution result (success / failure)
- Output or error message

If a tool command fails, no tool is launched, and an error dialog appears with the resolved path displayed.

## 8) RULE Mode (Floating Ruler)

`RULE` is a floating measurement tool for counting characters visually.

Current behavior:

- opens as an overlay inside the editor
- can be dragged away from the text you are inspecting
- shows a 132-column scale
- fixed top ruler is aligned to the editable text area (starts at text column 1, after the line-number gutter)
- updates distance live as the cursor moves
- supports inclusive block measurement with `B` / `B`
- works across multiple lines

### Typical use

1. Put the cursor where measurement should begin.
2. Press `Ctrl+Q` `Ctrl+R`.
3. Move the cursor and read the live distance.
4. If you want a full inclusive span, press `B` at the first point and `B` again at the last point.
5. Press `ESC` to leave RULE mode.

### Practical examples

- measure the visible length of a string between quotes
- validate fixed-width fields
- confirm indentation width
- count a multi-line span

## 9) Calculator Utility

The calculator dialog is available under `Utilities > Calculator`.

Dialog fields:

- `Enter Mathematical Expression to be Calculated:`
- `Result of Last Calculation`
- `Ok` (calculate)
- `Cancel` (close)

Supported operations:

- Arithmetic: `+`, `-`, `*`, `/`, `^`
- Functions: `sqr(...)`, `int(...)`, `hex(...)`, `bin(...)`, `dec(...)`
- Bitwise: `XOR`, `AND`, `OR`, `NOT`
- Shift/rotate: `<<`, `>>`, `shl(a,n)`, `shr(a,n)`, `rol(a,n)`, `ror(a,n)`

Examples:

- `2+3*4`
- `sqr(81)+2^3`
- `&H10 + &B11`
- `NOT 0 AND 15`
- `(1 << 4) + rol(1,3)`

### Save / Copy dialog behavior

- `Save As` and file copy dialogs now suggest the `.asc` extension explicitly for new MSX-BASIC ASCII documents.
- These dialogs accept both `.asc` and `.amx`, keeping the workflow ready for future `.amx` file creation.

## 10) Tool Launch

External tools can be launched from `Utilities` menu:

- **Open openMSX**: Launches the MSX emulator (detached process; app continues running).
- **Run msxbas2rom**: Converts the current file via `msxbas2rom`.
  - Shows output in a dialog when complete.
  - Path configured in `Utilities > Configure...`.
- **Run BASIC Dignified**: Transpiles the current file via BASIC Dignified.
  - Shows output in a dialog when complete.
  - Path configured in `Utilities > Configure...`.
- **Run MSX Encoding**: Runs encoding tool on the current file.
  - Shows output in a dialog when complete.
  - Path configured in `Utilities > Configure...`.

Tool paths accept either:

- **Direct file path**: Executable or script file location (e.g., `C:\tools\openmsx.exe`).
- **Directory path**: Folder where the tool is installed; auto-detection finds the most likely executable.
- **Fallback**: If the configured path is missing, the folder parent is scanned for candidates.

## 11) Recommended Workflow

1. Start WS7.
2. Open a file from the startup screen.
3. Use `Ctrl+N` for new working tabs.
4. Use `Ctrl+K` `Ctrl+S` to save frequently.
5. Use `Ctrl+W` to close the current tab safely.
6. Configure external tools in `Utilities > Configure...` and test them before use.

## 12) Notes

- The project is evolving with a focus on WordStar 7 interaction fidelity.
- Not all legacy commands are complete yet; pending items appear as "next block" in the app.
- Font, style, and theme preferences persist across sessions via local SQLite storage.
- Tool paths are validated and cached, with lightweight probes executed to ensure availability before launch.

