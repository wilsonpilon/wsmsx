# RULE Mode - Floating Character Ruler

## Overview

`RULE` is a floating measurement overlay for counting characters visually inside the editor.

It is **not** the old fixed ruler-at-the-top behavior. Instead, it opens a draggable panel that:

- follows the current cursor position,
- anchors measurement at the cursor position where RULE was enabled,
- shows a **132-column visual scale**,
- supports **multi-line counting** through absolute character positions,
- can measure a full span with the `B` / `B` block flow.

## Activation

- Shortcut: `Ctrl+Q,R`
- Menu: `Utilities -> RULE (Regua)`
- Exit: `ESC`
- Toggle off: `Ctrl+Q,R` again

When RULE is activated, the current cursor location becomes the initial measurement anchor.

## What the ruler shows

The floating panel contains:

1. **Top scale row**: continuous digit scale across **132 columns**
2. **Decade markers**: `10`, `20`, `30` ... `130`
3. **Green guide line** below the scale
4. **Position summary**: origin and current cursor absolute positions
5. **Live distance** in characters
6. **Block measurement summary**
7. **Footer title** for the draggable ruler panel

## Basic measurement flow

1. Place the cursor where you want to start measuring.
2. Press `Ctrl+Q,R`.
3. Drag the ruler if needed so it does not cover your text.
4. Move the cursor.
5. Read the live `Distance` value.
6. Press `ESC` when finished.

Distance is calculated as:

```text
distance = current_cursor_absolute_position - rule_anchor_position
```

## Block measurement with `B`

RULE also supports a quick block-count workflow:

1. Activate RULE.
2. Move to the first point.
3. Press `B` once -> block start is stored.
4. Move to the final point.
5. Press `B` again -> block end is stored and the ruler shows total characters.

Notes:

- The block count is **inclusive**.
- It works across multiple lines.
- Pressing `B` again after a completed block starts a new block measurement.

## Important behavior notes

- RULE is **per tab**.
- The ruler is **floating and draggable**.
- The scale is aligned to fixed character cells.
- The measurement model is based on **absolute text positions**, not screen columns.
- Mouse dragging is for moving the ruler panel; block measuring is keyboard-driven with `B`.

## Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Q,R` | Toggle RULE on/off |
| `ESC` | Exit RULE mode |
| `B` | Mark block start / block end while RULE is active |

## Recommended use cases

- Measure the length of strings between quotes
- Validate fixed-width fields
- Check code alignment and spacing
- Count a character span across multiple lines

## Current implementation summary

- Floating overlay inside the editor tab
- 132-column scale rendered with fixed cell alignment
- Live cursor tracking
- Block span measurement with `B`
- Clean exit with `ESC`

