# OUTLINE - Continuidade do Projeto WS7 (migracao de PC)

Este arquivo resume o plano e o estado do projeto para retomar rapidamente em outra maquina.

> **Fonte de verdade:** o codigo no repositorio. O chat ajuda no contexto, mas decisoes finais devem ser validadas em arquivos como `internal/ui/editor.go` e `internal/input/commands.go`.

## 1) Objetivo

- Manter o WS7 (Go + Fyne) com experiencia inspirada no WordStar.
- Padronizar toda a UI para ingles.
- Evoluir comandos por chords `Ctrl` (especialmente `Ctrl+K` e `Ctrl+Q`).
- Implementar edicao por blocos com **clipboard interno do WS7** (separado do clipboard do Windows).

## 2) Contexto do projeto (estrutura e stack)

- Workspace: `C:\Users\wilso\go\src\ws7`
- Entry point: `cmd/ws7/main.go`
- UI principal/editor: `internal/ui/editor.go`
- Tela inicial/browser: `internal/ui/filebrowser.go`
- Resolver de comandos/chords: `internal/input/commands.go`
- Persistencia local: `internal/store/sqlite/store.go`
- Paths locais: `internal/config/paths.go`
- Versionamento da app: `internal/version/version.go`
- Manual: `MANUAL.md`
- Release notes: `CHANGELOG.md`
- Build Windows: `build.ps1`

Stack confirmada:

- Go 1.22 (`go.mod`)
- `fyne.io/fyne/v2 v2.5.0`
- `modernc.org/sqlite v1.30.1`

## 3) Decisoes de arquitetura/UX tomadas

- **Editor com abas** (`DocTabs`) e indicador de arquivo sujo (`*` + icone).
- **Resolver WordStar** com estado de prefixo e timeout de chord.
- **Blocos WS7 independentes do Windows clipboard**:
  - Marcacao por offsets internos (`B`/`K`).
  - Clipboard interno (`internalBlockClipboard`) para operacoes de bloco.
- **Indicadores visuais exclusivos do WS7** na status bar:
  - `[WS7-BLOCK:B]`, `[WS7-BLOCK:K]`, `[WS7-BLOCK:B,K]`
  - `[WS7-CLIP:N]`
- **Separar o que e implementado vs nao implementado** no menu (`[NI]`).
- **Version Build centralizado** em `internal/version/version.go` (titulo da janela e Status mostram versao).
- **Opening Menu expandido**:
  - `Utilities > Macros` (`MP/MR/MD/MS/MO/MY/ME`)
  - `Additional` (`AC/AH/AS/AG/AN`)
  - `Help` (`HR/HM/HO`) abrindo `README.md`, `MANUAL.md`, `OUTLINE.md` com render Markdown.

## 4) Linha do tempo funcional (resumo ate o ponto atual)

1. Base do editor com abas, abrir/salvar/fechar e status.
2. Traducao ampla para ingles em UI e documentacao (`README.md`, `MANUAL.md`, `internal/ui/*`).
3. Ajustes na tela inicial (`internal/ui/filebrowser.go`) para textos em ingles.
4. Implementacao dos comandos de bloco (prefixo `Ctrl+K`) no editor:
   - Marcar inicio/fim de bloco
   - Copiar/mover/deletar bloco
5. Inclusao de indicador visual de marcacao de bloco + indicador do clipboard interno.
6. Testes unitarios adicionados/atualizados em `internal/ui/editor_block_test.go`.
7. Versionamento inicial `0.1.0`, com `CHANGELOG.md` e secao `Unreleased` para fluxo continuo.
8. Menus da tela inicial ampliados (Utilities/Macros, Additional e Help com docs em Markdown).
9. Correcao do gutter de linhas em `internal/ui/linenumbers.go` (numeracao volta a atualizar corretamente).

## 5) Mapeamento de teclas WordStar (estado atual)

### Implementado (relevante para continuidade)

- Navegacao basica: `Ctrl+S/D/E/X`, `Ctrl+R/C`
- Arquivo: `Ctrl+K,S/T/D`, `Ctrl+O,K`, `Ctrl+O,?`, `Ctrl+P,?`, `Ctrl+K,Q,X`
- Abas: `Ctrl+N`, `Ctrl+W`
- Edicao: `Ctrl+Y`, `Ctrl+T`, `Ctrl+Q,Y`, `Ctrl+Q,[DEL]`
- Blocos WS7:
  - `Ctrl+K,B` = Mark Block Begin
  - `Ctrl+K,K` = Mark Block End
  - `Ctrl+K,C` = Copy Block (para clipboard interno WS7)
  - `Ctrl+K,V` = Move Block (ou paste interno quando nao ha bloco marcado)
  - `Ctrl+K,Y` = Delete Block

### Parcial / nao implementado

- Itens marcados com `[NI]` no menu (ex.: operacoes com "Other Window").
- Alguns comandos de navegacao/legado ainda exibem "will be implemented".
- Opcoes novas de Opening Menu (placeholders):
  - `Utilities > Macros`: `MP`, `MR`, `MD`, `MS`, `MO`, `MY`, `ME`
  - `Additional`: `AC`, `AH`, `AS`, `AG`, `AN`
- `Help` na tela inicial: `HR`, `HM`, `HO`.

## 6) Estado atual de blocos WS7 vs clipboard do Windows

- **WS7 bloco interno (novo):**
  - Usa marcacao `B/K` e `internalBlockClipboard`.
  - Nao depende do clipboard do SO para `Ctrl+K,C/V/Y`.
  - Possui indicadores dedicados na status bar (`[WS7-BLOCK:*]`, `[WS7-CLIP:*]`).

- **Windows clipboard (separado):**
  - Continua por atalhos dedicados no menu/comandos (`Ctrl+K,[` / `Ctrl+K,]` conforme mapeado no projeto).
  - Nao deve ser confundido com a logica de bloco WS7.

## 7) Checklist de validacao no novo PC

1. Clonar/copiar o projeto para o novo ambiente.
2. Validar dependencias e build local.
3. Rodar testes.
4. Executar app e validar manualmente os chords principais.
5. Validar menus da tela inicial e viewer de docs em Markdown.
6. Validar numeracao de linhas durante scroll/edicao.

Comandos:

```powershell
cd C:\Users\<SEU_USUARIO>\go\src\ws7
go mod tidy
go test ./...
go run ./cmd/ws7
```

Build opcional Windows:

```powershell
./build.ps1
./build.ps1 -Configuration Release
```

Validacao funcional recomendada no app:

- Abrir arquivo, marcar `Ctrl+K,B` e `Ctrl+K,K`.
- Copiar bloco `Ctrl+K,C` e confirmar `[WS7-CLIP:N]`.
- Mover bloco `Ctrl+K,V`.
- Deletar bloco `Ctrl+K,Y`.
- Conferir se o status mostra `[WS7-BLOCK:*]` e nao mistura com clipboard do Windows.
- Na tela inicial, validar `Help` (`HR/HM/HO`) abrindo Markdown de `README.md`, `MANUAL.md`, `OUTLINE.md`.
- Na tela inicial, validar `Utilities > Macros` e `Additional` com itens corretos.
- No editor, validar se o gutter mostra todas as linhas e acompanha rolagem/cursor.

## 8) Prompt de retomada para colar no Copilot no novo PC

Use este prompt inicial:

```text
Contexto: Projeto WS7 em Go + Fyne (WordStar-like editor) em `C:\Users\<user>\go\src\ws7`.

Antes de qualquer alteracao:
1) Leia `OUTLINE.md`, `README.md`, `MANUAL.md`.
2) Leia `internal/input/commands.go` e `internal/ui/editor.go` para confirmar mapeamento real.
3) Rode `go test ./...` e reporte o resultado.

Objetivo atual:
- Continuar evolucao dos comandos WordStar e UX do editor.
- Preservar separacao entre clipboard interno WS7 (blocos) e clipboard do Windows.
- Manter toda a UI em ingles.

Ao propor mudancas:
- Liste arquivos/simbolos afetados.
- Nao presumir comportamento; validar no codigo.
- Atualizar/adicao de testes quando houver logica nova.
```

## 9) Proximos passos prioritarios

1. Destacar visualmente bloco selecionado no texto (alem da status bar).
2. Implementar "Go to Beginning/End of Block" (`Ctrl+Q,B` / `Ctrl+Q,K`).
3. Revisar comandos `[NI]` e priorizar os mais usados.
4. Padronizar testes para fluxos de teclado (chords) e regressao de status.
5. Manter disciplina de release: atualizar `CHANGELOG.md` (`Unreleased`) e depois gerar secao versionada.


