# WS7 Floating Ruler Guide

## What RULE does now

`RULE` is a floating ruler overlay designed to measure **characters**, not centimeters or editor columns.

Current behavior:

- floating, draggable panel
- live measurement from the cursor position where RULE was enabled
- fixed **132-column scale**
- block counting with `B` / `B`
- `ESC` to exit
- `Ctrl+Q,R` to toggle

## Quick start

### Activation

```text
Ctrl+Q,R   -> Toggle RULE on/off
ESC        -> Exit RULE mode
B          -> Mark block start / end while RULE is active
```

You can also open it from:

```text
Utilities -> RULE (Regua)
```

## How to use it

### 1) Live measurement from the current cursor

1. Put the cursor where measurement should begin.
2. Press `Ctrl+Q,R`.
3. Move the cursor.
4. Read the live distance shown by the ruler.

This is useful for measuring strings, indentation, padding, and column widths.

### 2) Block measurement with `B`

1. Activate RULE.
2. Move to the first point.
3. Press `B`.
4. Move to the final point.
5. Press `B` again.

The ruler then shows the total number of characters in the selected span.

This works across multiple lines because the measurement uses absolute character positions.

## Layout of the floating ruler

The current panel is organized like this:

1. continuous digit scale across 132 columns
2. decade markers (`10`, `20`, `30`, ...)
3. green guide line
4. origin/cursor summary
5. live distance
6. block summary
7. title/footer bar

This layout is intentionally optimized for immediate visual reading.

## Example

```text
10 PRINT "== teste ========="
```

To measure the string visually:

1. place the cursor at the first `=`
2. press `Ctrl+Q,R`
3. move to the last `=`
4. read the live distance

To measure an inclusive span:

1. press `B` at the first point
2. move to the last point
3. press `B` again
4. read the block total shown in the ruler

## Keyboard reference

| Key | Action | Scope |
|-----|--------|-------|
| `Ctrl+Q,R` | Toggle RULE on/off | Editor |
| `ESC` | Exit RULE | RULE mode |
| `B` | Mark block start/end | RULE mode |
| Mouse Drag | Move floating ruler | RULE mode |

## Notes

- RULE is tracked per tab.
- The panel position is screen-relative.
- The measurement base is reset when RULE is re-enabled.
- The scale is aligned to fixed character cells.

## Recommended checks

RULE is especially useful for:

- quoted string lengths
- fixed-width records
- alignment in BASIC data lines
- multi-line spans with `B` / `B`

---

Try it with `Ctrl+Q,R`, drag the ruler where you want it, and use `B` when you need a full span count.

