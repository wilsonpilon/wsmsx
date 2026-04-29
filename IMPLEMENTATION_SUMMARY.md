# Implementation Summary: RULE / Floating Ruler

## Scope

This document summarizes the current implementation state of the `RULE` feature after the recent UI and shortcut iterations.

## Current behavior

`RULE` is now implemented as a **floating ruler overlay** inside the editor.

It provides:

- floating draggable panel
- `Ctrl+Q,R` activation/toggle
- `ESC` exit
- fixed **132-column scale**
- live distance tracking from the activation cursor position
- inclusive block measurement with `B` / `B`
- multi-line support via absolute character positions

## Main files involved

### `internal/ui/floating_ruler.go`

Contains the widget and renderer for the floating ruler panel.

Current responsibilities:

- panel screen position (`posX`, `posY`)
- origin/cursor tracking
- block selection state (`blockStartPos`, `blockEndPos`)
- fixed-cell 132-column scale rendering
- decade marker rendering
- live distance text
- block summary text
- drag handling

Important behavior:

- the ruler is rendered as an overlay panel
- the visual scale is aligned by character-cell width
- `MarkBlockPoint()` handles the `B` / `B` workflow

### `internal/ui/editor.go`

Provides integration with the editor tab.

Current integration points:

- one `floatingRuler` per editor tab
- overlay composition through `container.NewStack(...)`
- absolute character conversion through `absoluteCharPos(...)`
- RULE toggle via `cmdRule()` / `setRuleMode(...)`
- live cursor synchronization to the floating ruler
- `B` handling while RULE is active
- `ESC` interception to leave RULE mode

### `internal/input/commands.go`

Current relevant mappings:

- `Ctrl+Q,R` -> `CmdRule`
- `Ctrl+O,L` -> `CmdGoDocBegin`

This avoids conflicts with the classic WordStar movement keys `Ctrl+E/S/D/X`.

## Display structure

The current ruler panel is laid out like this:

1. top continuous digit scale
2. decade markers row
3. green guide line
4. origin/cursor summary
5. live distance summary
6. block total summary
7. footer title bar

## Measurement model

### Live mode

- on activation, the current cursor position becomes the initial anchor
- moving the cursor updates the distance immediately

### Block mode

- first `B`: stores block start
- second `B`: stores block end and shows inclusive character count
- next `B`: starts a new block measurement cycle

### Multi-line support

All counts are based on absolute positions in the text buffer:

```text
absolute position = characters in previous lines + newline bytes + current column
```

## Tests updated/maintained

Relevant automated coverage includes:

- resolver chord tests for `Ctrl+Q,R`
- RULE toggle tests in the editor UI
- floating ruler drag tests
- scale generation/alignment tests
- block measurement tests

Typical commands used during validation:

```bash
go test ./internal/input ./internal/ui
go test ./...
```

## Documentation alignment

Documentation was cleaned up to match current behavior:

- no more references to the old `Ctrl+O,R` shortcut
- no more references to right-click origin as current behavior
- no more descriptions of RULE as a fixed top ruler
- current docs now describe `Ctrl+Q,R`, `ESC`, and `B` / `B`

## Practical result

The feature now matches the intended workflow much better:

- open a floating ruler where needed
- drag it out of the way
- read live character distances immediately
- use `B` to measure a full inclusive span when needed

## Suggested future improvements

- visual highlight in the editor for the measured block span
- optional panel close button
- optional save/restore of ruler position per tab
- richer block visualization in the ruler itself

## Status

The current RULE implementation is integrated, documented, and covered by automated tests.

