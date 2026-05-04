# Fixação do Problema de Posicionamento do Cursor na Gutter de Números de Linha

## Problema Relatado
O cursor estava aparecendo uma linha antes do devido. Por exemplo:
- Quando o texto estava sendo digitado na linha 13, coluna 1
- O cursor na gutter e na régua mostrava linha 12, coluna 2
- O cursor avançava de forma incorreta após pressionar Enter

## Causa Raiz
Havia um erro de off-by-one (erro de posicionamento por uma unidade) na conversão entre:
- `cursorRow` de Fyne (0-based: 0 = primeira linha)
- `lineNum` na gutter (1-based: 1 = primeira linha)

A comparação original estava usando diferentes bases de numeração:
- `lineNum = topLine + i + 1` (1-based)
- `isCursor = (topLine + i) == cursorLine` (0-based)

## Solução Implementada

### 1. editor.go (linha 2732)
```go
// Antes:
e.lineNums.UpdateWithOffset(lineCount, topLine, e.topOffset, e.cursorRow)

// Depois:
e.lineNums.UpdateWithOffset(lineCount, topLine, e.topOffset, e.cursorRow+1)
```
Agora o `cursorLine` é passado como 1-based (correspondendo a `lineNum`).

### 2. linenumbers.go (linha 157)
```go
// Antes:
curVisIdx := r.w.cursorLine - r.w.topLine

// Depois:
curVisIdx := r.w.cursorLine - r.w.topLine - 1
```
Ajusta o cálculo porque `cursorLine` é agora 1-based enquanto `topLine` é 0-based.

### 3. linenumbers.go (linha 213)
```go
// Antes:
isCursor := (r.w.topLine + i) == r.w.cursorLine

// Depois:
isCursor := lineNum == r.w.cursorLine
```
Simplifica a comparação usando ambos 1-based (ambos representam números de linha para exibição).

## Testes Adicionados
Criado arquivo `linenumbers_test.go` com testes para validar:
- Cursor na primeira linha
- Cursor na segunda linha
- Cursor em linha 13 com topLine em 11
- Cursor em linha 12 com topLine em 11

Todos os testes passam, confirmando que o cursor agora aparece na posição correta.

