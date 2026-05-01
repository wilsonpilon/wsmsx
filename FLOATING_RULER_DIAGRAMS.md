# Floating Ruler - Visual Diagrams

## Current panel layout

```text
┌──────────────────────────────────────────────────────────────┐
│ 1234567890123456789012345678901234567890... up to 132       │
│         10        20        30        40 ... 130            │
│ ──────────────────────────────────────────────────────────── │  ← green guide line
│ Origem: char 21  |  Cursor: char 38                         │
│ >>> Distancia: +17 chars <<<                                │
│ Bloco: pressione B no inicio e B no fim                     │
│ RULER (drag to move)  B=inicio/fim do bloco                 │
└──────────────────────────────────────────────────────────────┘
```

## Interaction flow

```text
Ctrl+Q,R
   ↓
RULE enabled
   ↓
Current cursor position becomes anchor
   ↓
Move cursor
   ↓
Distance updates in real time
   ↓
ESC or Ctrl+Q,R again
   ↓
RULE disabled
```

## Block flow with `B`

```text
RULE active
   ↓
Press B
   ↓
Store block start
   ↓
Move cursor
   ↓
Press B again
   ↓
Store block end
   ↓
Show inclusive character total
```

## Multi-line counting model

```text
Line 1: "Hello"
Line 2: "World"

Buffer model:
"Hello\nWorld"

Absolute positions include the newline.
```

## Editor composition

```text
Editor tab
   ├─ main editor content
   ├─ line-number gutter
   ├─ status bar
   └─ floating ruler overlay (when RULE is on)
```

## Typical usage example

```text
10 PRINT "== teste ========="

1. Put cursor on first =
2. Press Ctrl+Q,R
3. Move to last =
4. Read live distance

Optional:
5. Press B at first point
6. Press B at second point
7. Read inclusive block total
```

