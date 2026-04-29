# Floating Ruler - Practical Examples

## Example 1: Measuring a BASIC string visually

```text
10 PRINT "== TESTE MEASURE ===="
```

Steps:

1. Put the cursor on the first `=`.
2. Press `Ctrl+Q,R`.
3. Move the cursor to the last `=`.
4. Read the live distance shown by the ruler.

Use this when you want a quick visual count without creating a block.

---

## Example 2: Inclusive span with `B` / `B`

```text
10 DATA "ABCDEF"
```

To count the full span from `A` to `F` inclusively:

1. Activate RULE with `Ctrl+Q,R`.
2. Put the cursor on `A`.
3. Press `B`.
4. Move the cursor to `F`.
5. Press `B` again.

The ruler block summary shows the total character count for the span.

---

## Example 3: Multi-line block measurement

```text
10 PRINT "HELLO"
20 PRINT "WORLD"
```

To count a span from the `H` of the first line to the `D` of the second line:

1. Activate RULE.
2. Move to `H`.
3. Press `B`.
4. Move to `D`.
5. Press `B` again.

Because RULE uses absolute text positions, the count includes the newline between the lines.

---

## Example 4: Checking indentation width

```text
10 IF X > 0 THEN
20   PRINT "OK"
```

To validate the indentation before `PRINT`:

1. Put the cursor at the start of line `20`.
2. Press `Ctrl+Q,R`.
3. Move to the `P` in `PRINT`.
4. Read the live distance.

This is useful for checking spacing in aligned BASIC code.

---

## Example 5: Fixed-width field validation

```text
NAME      AGE  CITY
JOHN      25   NYC
```

To verify that `AGE` begins at the expected visual position:

1. Put the cursor at the start of `NAME` data.
2. Press `Ctrl+Q,R`.
3. Move to the first digit in `25`.
4. Compare the live distance with the expected field offset.

---

## Keyboard reference

```text
Action                          Keyboard
──────────────────────────────────────────
Activate / Deactivate RULE      Ctrl+Q,R
Exit RULE                       ESC
Block start / block end         B
Move floating ruler             Mouse drag
```

---

## Troubleshooting

| Issue | Suggested action |
|-------|------------------|
| Ruler not appearing | Press `Ctrl+Q,R` in the editor |
| Ruler is covering text | Drag it to another screen position |
| Distance looks wrong | Re-enable RULE from the correct starting cursor position |
| Need full span count | Use `B` at the first point and `B` again at the end |
| Ruler disappeared | Press `Ctrl+Q,R` again or check if `ESC` was used |

---

## Quick reminder

- Use live distance for fast visual measurement.
- Use `B` / `B` for inclusive span counting.
- Use `ESC` when you are done.

Start measuring with `Ctrl+Q,R`.

### Cross-Line Measurements

The ruler works across multiple lines since it uses absolute positions:

```
Line 1: ...start...
Line 2: ...end...

Measure from position at line 1 all the way to line 2
= combined character distance
```

---

## Practice Exercises

1. **Easy**: Measure the length of any quoted string in your code
2. **Medium**: Verify all fields in a CSV have consistent widths
3. **Hard**: Check multi-line comment block length and indentation

---

Start measuring! Press `Ctrl+Q,R` now to activate the floating ruler! 🚀

