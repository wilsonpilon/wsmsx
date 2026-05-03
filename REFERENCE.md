# REFERENCE - WS7 Commands and Keybinds

This document lists the current WS7 keyboard command catalog, including items that are already wired but still marked as not implemented (`[NI]`).

Status legend:
- `OK` = available in current build
- `NI` = present in UI/command resolver, but still not implemented

## Editor and global keybinds

| Status | Command | Shortcut |
|---|---|---|
| OK | Cursor Left | `Ctrl+S` |
| OK | Cursor Right | `Ctrl+D` |
| OK | Cursor Up | `Ctrl+E` |
| OK | Cursor Down | `Ctrl+X` |
| OK | Page Up | `Ctrl+R` |
| OK | Page Down | `Ctrl+C` |
| OK | New File | `Ctrl+N` |
| OK | Close | `Ctrl+W` |
| OK | Delete Line | `Ctrl+Y` |
| OK | Delete Word | `Ctrl+T` |
| OK | Undo | `Ctrl+U` |
| OK | Repeat Find | `Ctrl+L` |
| OK | Insert/Overtype Mode | `Ctrl+V` |
| OK | Save | `Ctrl+K,S` |
| OK | Save As | `Ctrl+K,T` |
| OK | Save and Close | `Ctrl+K,D` |
| OK | Open/Switch | `Ctrl+O,K` |
| NI | Print | `Ctrl+K,P` |
| NI | Change Printer | `Ctrl+P,?` |
| OK | Copy File | `Ctrl+K,O` |
| OK | Delete File | `Ctrl+K,J` |
| OK | Rename File | `Ctrl+K,E` |
| OK | Change Drive/Directory | `Ctrl+K,L` |
| OK | Run PS Command | `Ctrl+K,F` |
| OK | Status | `Ctrl+O,?` |
| OK | Exit | `Ctrl+K,Q,X` |
| OK | Mark Block Begin | `Ctrl+K,B` |
| OK | Mark Block End | `Ctrl+K,K` |
| OK | Move Block | `Ctrl+K,V` |
| NI | Move Block from Other Window | `Ctrl+K,G` |
| OK | Copy Block | `Ctrl+K,C` |
| NI | Copy Block from Other Window | `Ctrl+K,A` |
| OK | Paste from Clipboard | `Ctrl+K,[` |
| OK | Copy to Clipboard | `Ctrl+K,]` |
| OK | Copy to Another File | `Ctrl+K,W` |
| OK | Include File | `Ctrl+K,R` |
| OK | Convert Uppercase | `Ctrl+K,"` |
| OK | Convert Lowercase | `Ctrl+K,'` |
| OK | Convert Capitalize | `Ctrl+K,.` |
| OK | Delete Block | `Ctrl+K,Y` |
| NI | Mark Previous Block | `Ctrl+K,U` |
| NI | Column Block Mode | `Ctrl+K,N` |
| NI | Column Replace Mode | `Ctrl+K,I` |
| OK | Find | `Ctrl+Q,F` |
| OK | Find and Replace | `Ctrl+Q,A` |
| OK | Go to Character | `Ctrl+Q,G` |
| OK | Go to Page | `Ctrl+Q,I` |
| NI | Go to Font Tag | `Ctrl+Q,=` |
| NI | Go to Style Tag | `Ctrl+Q,<` |
| NI | Go to Note | `Ctrl+Q,N,G` |
| NI | Go to Previous Position | `Ctrl+Q,P` |
| NI | Go to Last Find/Replace | `Ctrl+Q,V` |
| NI | Go to Beginning of Block | `Ctrl+Q,B` |
| NI | Go to End of Block | `Ctrl+Q,K` |
| OK | Go to Document Beginning | `Ctrl+O,L` |
| OK | Go to Document End | `Ctrl+Q,C` |
| NI | Scroll Continuously Up | `Ctrl+Q,W` |
| NI | Scroll Continuously Down | `Ctrl+Q,Z` |
| OK | Delete Left of Cursor | `Ctrl+Q,DEL` |
| OK | Delete Right of Cursor | `Ctrl+Q,Y` |
| OK | BASIC DELETE | `Ctrl+Q,D` |
| OK | BASIC RENUM | `Ctrl+Q,E` |
| NI | Edit Note | `Ctrl+O,N,D` |
| NI | Convert Note | `Ctrl+O,N,V` |
| NI | Auto Align | `Ctrl+O,A` |
| OK | RULE | `Ctrl+Q,R` |
| OK | Calculator | `Ctrl+Q,M` |
| OK | Style Bold | `Ctrl+P,B` |
| OK | Style Font | `Ctrl+P,=` |
| OK | Insert Extended Character | `Ctrl+M,G` |
| OK | Close Dialog | `Ctrl+O,Enter` |
| OK | Set Marker (0-9) | `Ctrl+K,[0-9]` |
| OK | Go to Marker (0-9) | `Ctrl+Q,[0-9]` |
| NI | Scroll Up (Legacy) | `(none)` |
| OK | Insert Line (Legacy) | `(none)` |
| OK | Word Count | `(none)` |

## Opening menu mnemonic shortcuts

These are menu mnemonics shown by the opening screen labels (not all are full editor Ctrl chords):

| Status | Menu Command | Shortcut |
|---|---|---|
| OK | File > New | `S` |
| OK | File > Open Document | `D` |
| OK | File > Open Nondocument | `N` |
| NI | File > Print | `P` |
| NI | File > Print from keyboard | `K` |
| OK | File > Copy | `O` |
| OK | File > Delete | `Y` |
| OK | File > Rename | `E` |
| OK | File > Change Drive | `L` |
| OK | File > Run CMD Command | `R` |
| OK | File > Status | `?` |
| OK | File > Exit MSXStar | `X` |
| NI | Utilities > Macros > Play | `MP` |
| NI | Utilities > Macros > Record | `MR` |
| NI | Utilities > Macros > Edit/Create | `MD` |
| NI | Utilities > Macros > Single Step | `MS` |
| NI | Utilities > Macros > Copy | `MO` |
| NI | Utilities > Macros > Delete | `MY` |
| NI | Utilities > Macros > Rename | `ME` |
| NI | Additional > Character Editor | `AC` |
| NI | Additional > Hexa Editor | `AH` |
| NI | Additional > Sprite Editor | `AS` |
| NI | Additional > Graphos | `AG` |
| NI | Additional > Noise Editor | `AN` |
| OK | Help > README | `HR` |
| OK | Help > MANUAL | `HM` |
| OK | Help > OUTLINE | `HO` |

## Notes for keybind customization

- The keybind catalog is persisted in SQLite in the `keybinds` table.
- `Utilities > Keybinds...` now shows a filterable table (`Context`, `Implemented`, `Configurable`).
- Shortcuts can now be edited for configurable commands with conflict detection before save.
- The current filtered view can be exported using `Export current keybinds (JSON)`.
- Future plan: import custom profiles and add chord-capture UI for direct key recording.

