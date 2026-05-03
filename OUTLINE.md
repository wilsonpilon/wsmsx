# OUTLINE - Continuidade do Projeto WS7 (handoff para outro PC)

Este arquivo e o guia rapido para retomar o desenvolvimento do WS7 em outro ambiente com apoio do Copilot.

> Fonte de verdade: o codigo do repositorio. Use este OUTLINE como contexto inicial, mas valide comportamento real em `internal/ui/editor.go`, `internal/ui/theme.go`, `internal/ui/syntax_overlay.go` e `internal/input/commands.go`.

## 1) Estado atual (snapshot)

- Versao atual: `0.2.5`
- Branch de trabalho: `main` (validar no Git antes de continuar)
- Workspace atual de referencia: `E:\wsmsx`
- Stack:
  - Go `1.22`
  - Fyne `v2.7.3`
  - SQLite (`modernc.org/sqlite v1.30.1`)

## 2) Objetivo do produto

- Editor/IDE leve para MSX-BASIC com fluxo inspirado no WordStar.
- Prioridade em produtividade por teclado (chords `Ctrl+K`, `Ctrl+Q`, etc.).
- UX moderna sem perder fidelidade de operacao classica.

## 3) Arquivos-chave para retomada

- Entry point: `cmd/ws7/main.go`
- Core da UI/editor: `internal/ui/editor.go`
- Browser inicial: `internal/ui/filebrowser.go`
- Tema e temas de sintaxe: `internal/ui/theme.go`
- Overlay de syntax highlight: `internal/ui/syntax_overlay.go`
- Entry/ruler widgets: `internal/ui/ruler.go`, `internal/ui/linenumbers.go`
- Botao com tooltip por icone: `internal/ui/icon_tooltip_button.go`
- Resolver de comandos: `internal/input/commands.go`
- Persistencia local: `internal/store/sqlite/store.go`
- Versao da app: `internal/version/version.go`
- Docs de produto: `README.md`, `MANUAL.md`, `CHANGELOG.md`

## 4) Entregas recentes importantes (ate 0.2.5)

1. Syntax highlighting MSX-BASIC expandido por categorias:
   - instruction, jump, function, operator, number, string, comment, identifier.
2. Renderizacao inline de syntax highlight via overlay no editor.
3. Sistema de temas de sintaxe com presets:
   - `MSX Dark`, `MSX Light`, `MSX Green Screen`, `Cobalt`, `Amber`.
4. Tema custom de sintaxe:
   - criar, editar, resetar por preset, deletar, importar/exportar JSON.
5. `Configure` com preview ao vivo:
   - mudancas de tema (editor/sintaxe) aplicam preview enquanto dialogo esta aberto;
   - `Cancel`/fechar sem salvar faz rollback do preview.
6. UI compacta de cores de sintaxe:
   - acao de abrir picker por icone (palette) + tooltip curto;
   - acao de copiar hex por icone + tooltip curto.
7. Correcao de duplicacao visual de texto no editor:
   - camada base do `Entry` permanece transparente sob o overlay.

## 5) Regras de decisao para continuar

- Nao presumir comportamento por historico de chat; validar no codigo.
- Se alterar UX de comandos, atualizar tambem `MANUAL.md` e `CHANGELOG.md`.
- Se alterar fluxo de tema/sintaxe, revisar impacto em:
  - `cmdConfigure` em `internal/ui/editor.go`
  - serializacao/parse em `internal/ui/theme.go`
  - render overlay em `internal/ui/syntax_overlay.go`
- Sempre preservar separacao entre:
  - clipboard interno WS7 (blocos)
  - clipboard do sistema operacional.

## 6) Checklist de migracao para novo PC

1. Clonar o repositorio no path desejado.
2. Confirmar versoes de Go e dependencias.
3. Rodar testes.
4. Executar app e validar fluxo principal.

Comandos base:

```powershell
cd E:\wsmsx
go mod tidy
go test ./...
go run ./cmd/ws7
```

Build opcional:

```powershell
./build.ps1
./build.ps1 -Configuration Release
```

## 7) Smoke test manual recomendado

- Abrir arquivo e editar texto em aba.
- Validar chords basicos (`Ctrl+S/D/E/X`, `Ctrl+R/C`, `Ctrl+N`, `Ctrl+W`).
- Validar blocos WS7 (`Ctrl+K,B/K/C/V/Y`) e indicadores `[WS7-*]`.
- Abrir `Utilities > Configure...` e validar:
  - troca de `Editor Theme` com preview ao vivo;
  - troca de `Syntax Theme` com preview ao vivo;
  - em tema custom, alterar cor por icone de palette;
  - copiar hex por icone de copy;
  - `Cancel` desfaz preview; `Save` persiste.
- Validar syntax overlay:
  - sem texto duplicado/deslocado;
  - categorias de token com cores distintas.
- Validar launch de ferramentas configuradas (openMSX/msxbas2rom/BASIC Dignified/MSX Encoding).

## 8) Prompt pronto para colar no Copilot (novo PC)

Use este prompt no inicio da sessao:

```text
Contexto: Projeto WS7 (Go + Fyne), editor estilo WordStar para MSX-BASIC.
Workspace atual: E:\wsmsx.
Versao alvo atual: 0.2.5.

Antes de qualquer mudanca:
1) Leia OUTLINE.md, README.md, MANUAL.md e CHANGELOG.md.
2) Leia internal/ui/editor.go, internal/ui/theme.go, internal/ui/syntax_overlay.go e internal/input/commands.go.
3) Rode go test ./... e reporte o resultado.

Regras:
- Nao presumir comportamento; confirmar no codigo.
- Priorizar bugs/regressao e impacto de UX.
- Ao mudar funcionalidade, atualizar docs relevantes e changelog.
- Manter separacao entre clipboard interno WS7 e clipboard do sistema.

Objetivo desta sessao:
[descreva aqui o proximo passo concreto]
```

## 9) Proximos passos priorizados

1. Melhorar acessibilidade/usabilidade do editor de temas (atalhos, feedback visual de foco).
2. Revisar e reduzir itens `[NI]` mais usados no fluxo WordStar.
3. Expandir integracao de ferramentas externas com parametros por projeto.
4. Adicionar mais testes de regressao para `Configure` + syntax preview.
5. Continuar disciplina de release:
   - registrar primeiro em `## [Unreleased]` no `CHANGELOG.md`;
   - depois cortar secao versionada.
