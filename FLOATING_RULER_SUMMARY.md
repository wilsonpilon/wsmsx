# FLOATING RULER - EXECUTIVE SUMMARY

## Current status

The `RULE` feature is now a **floating character ruler** integrated into the editor.

It is no longer documented as a fixed top ruler or a right-click origin tool.

## What it does today

- opens as a draggable overlay in the editor
- uses `Ctrl+Q,R` to toggle on/off
- uses `ESC` to exit
- shows a **132-column visual scale**
- keeps a live anchor based on the cursor position at activation time
- tracks the cursor in real time
- supports inclusive block measurement with `B` / `B`
- works across multiple lines through absolute character positions

## Primary use cases

- measure strings between quotes
- validate fixed-width fields
- check alignment and indentation
- count a span across several lines

## Operator workflow

### Live measurement

1. Move the cursor to the starting point.
2. Press `Ctrl+Q,R`.
3. Move the cursor.
4. Read the live distance.

### Block measurement

1. Press `B` at the first point.
2. Move to the final point.
3. Press `B` again.
4. Read the block total shown by the ruler.

## Visual structure

The current ruler layout is optimized for fast reading:

1. continuous top digit scale
2. decade markers underneath
3. green guide line
4. origin/cursor position summary
5. live distance
6. block summary
7. footer title bar

## Key implementation notes

- The overlay lives inside the editor tab content.
- Positioning is screen-relative and draggable.
- Character math is based on absolute text positions.
- The ruler scale is rendered with fixed character-cell alignment.
- The `B` key is consumed only while RULE mode is active for block measurement.

## Shortcut summary

| Shortcut | Meaning |
|----------|---------|
| `Ctrl+Q,R` | Toggle RULE |
| `ESC` | Exit RULE |
| `B` | Mark block start/end while RULE is active |

## Validation snapshot

- input resolver updated for `Ctrl+Q,R`
- editor menu and status updated
- UI tests cover RULE toggle and ruler behavior
- full project test suite passes

## Related docs

- `FLOATING_RULER.md` - technical reference
- `FLOATING_RULER_GUIDE.md` - user guide
- `FLOATING_RULER_EXAMPLES.md` - usage examples
- `IMPLEMENTATION_SUMMARY.md` - implementation notes

