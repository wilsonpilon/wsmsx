# MSX Text Encoding

VS Code extension that adds support for MSX legacy text encodings. Open, edit, and save files using the original MSX character sets.

## Supported Character Sets

| Charset | ID | Description |
|---------|-----|-------------|
| **International** | `msx-international` | Default MSX charset (CP437-based with MSX block elements) |
| **Brazilian (BR)** | `msx-brazilian-br` | Gradiente Expert DDPlus / Sharp Hotbit 1.2 |
| **Brazilian (BG)** | `msx-brazilian-bg` | Gradiente Expert XP-800 |
| **Brazilian (BH)** | `msx-brazilian-bh` | Sharp Hotbit 1.1 |
| **Arabic (AR)** | `msx-arabic-ar` | Bawareth Perfect MSX1 / Yamaha AX500 |
| **Arabic (AE)** | `msx-arabic-ae` | Al Alamiah AX-170 |
| **Japanese** | `msx-japanese` | JIS X 0201 katakana + hiragana + kanji |
| **Russian** | `msx-russian` | Cyrillic in JCUKEN layout order |

## Features

### Open files with MSX encoding
Use **MSX Encoding: Open File with MSX Encoding** from the Command Palette or the explorer context menu. The file is opened via a virtual filesystem (`msxenc://`) that transparently decodes MSX bytes to Unicode and encodes back on save.

### Convert files
- **MSX Encoding: Convert MSX → UTF-8** — Reads an MSX-encoded file and converts it to UTF-8 in place.
- **MSX Encoding: Convert UTF-8 → MSX Encoding** — Converts a UTF-8 file to MSX encoding in place.

### Status bar
When editing a file opened with MSX encoding, the status bar shows the active charset. Click it to switch to a different charset.

### Character Map
A 16×16 grid showing all 256 characters of the selected MSX charset. Use **MSX Encoding: Show MSX Character Map** or `Ctrl+Shift+M` (`Cmd+Shift+M` on macOS).

- Choose any of the 8 supported MSX charsets.
- Click cells to accumulate characters in a text field.
- **Insert** button sends the text to your cursor in the active editor.
- Hover over cells to see hex values and Unicode codepoints.
- Graphic characters (0x00-0x1F) are highlighted with a blue tint.

## Escape mechanism (0x01)

MSX uses byte `0x01` as an escape prefix for graphic characters in the `0x00-0x1F` VRAM range. The byte following the escape has a `+0x40` offset added:

- **Decoding**: `0x01` + byte → graphic character at position `byte - 0x40` (e.g., `0x01 0x54` → position `0x14` = ├)
- **Encoding**: graphic characters that map to `0x00-0x1F` → `0x01` + `(byte + 0x40)`
- Without the escape prefix, bytes `0x00-0x1F` are treated as standard control codes (TAB, LF, CR, etc.)

## Arabic charset (simplified)

The Arabic charset uses a simplified mapping: each byte decodes to the abstract Unicode letter (e.g., ب BEH) regardless of the contextual form (initial, medial, final, isolated). The Arabic Shaping Algorithm is not implemented — Unicode rendering handles visual presentation.

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `msxEncoding.defaultCharset` | `msx-international` | Default charset when opening files |

## Development

```bash
npm install        # Install dependencies
npm run compile    # Build the extension
npm run watch      # Build in watch mode
```

Press **F5** to launch the Extension Development Host for testing.

## Packaging

To generate a `.vsix` file for distribution:

```bash
npm run package
```

This creates `msx-text-encoding-<version>.vsix`. Install it in VS Code with:

```bash
code --install-extension msx-text-encoding-0.1.0.vsix
```

Or via the Command Palette: **Extensions: Install from VSIX...**

## References

Character mappings are based on the [openMSX](https://github.com/openMSX/openMSX) emulator's authoritative `MSXVID*.TXT` mapping files from Unicode 13.0+ (Symbols for Legacy Computing block).
