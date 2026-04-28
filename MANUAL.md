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
- `View`
- `Insert`
- `Utilities`

Current editor features:

- Tab-based editing (multiple documents).
- New tab named `untitled`.
- Tab close confirmation when unsaved changes exist.
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
| `Ctrl+N` | New tab (`untitled`) |
| `Ctrl+W` | Close current tab |

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

## 6) Recommended Workflow

1. Start WS7.
2. Open a file from the startup screen.
3. Use `Ctrl+N` for new working tabs.
4. Use `Ctrl+K` `Ctrl+S` to save frequently.
5. Use `Ctrl+W` to close the current tab safely.

## 7) Notes

- The project is evolving with a focus on WordStar 7 interaction fidelity.
- Not all legacy commands are complete yet; pending items appear as "next block" in the app.

