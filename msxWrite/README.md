# MSX-Write

**MSX-Write** (anteriormente msxRead) é uma ferramenta completa para desenvolvimento e visualização de arquivos para a plataforma MSX, com foco especial no MSX-BASIC.

![Screenshot do MSX-Write](read-00.png)

## Editor MSX-BASIC Principal
O coração do projeto é um editor robusto inspirado no estilo **QuickBasic**, que agora serve como a janela principal da aplicação.

![Screenshot do Editor MSX-Write](read-01.png)

### Funcionalidades de Destaque:
- **Suporte a Múltiplos Dialetos:** Escolha entre **MSX-BASIC clássico**, **MSX Basic Dignified** ou **MSXBas2Rom**. O editor ajusta regras como a obrigatoriedade de números de linha conforme o dialeto selecionado.
- **Mapa do Programa:** Ferramenta avançada para analisar variáveis (nome, tipo, frequência de uso, estimativa de memória) e fluxo de execução (`GOTO`/`GOSUB`), identificando sub-rotinas automaticamente.
- **Destaque de Sintaxe (Syntax Highlighting):** Realce em tempo real de comandos, funções, strings, comentários e números de linha, totalmente personalizável.
- **Auto-Formatação (Beautify):** Organiza o código automaticamente ao digitar, garantindo espaçamento ideal e legibilidade.
- **Renumeração Inteligente (RENUM):** Atualiza automaticamente todas as referências de salto (`GOTO`, `GOSUB`, `THEN`, `ELSE`, etc.) usando um motor baseado em SQLite.
- **Configuração por Abas:** Interface de configurações organizada em abas (Principal, Dialetos, Emulador, Extras), permitindo configurar caminhos de emuladores como **openMSX** e **fMSX**.
- **Compatibilidade:** Suporte a arquivos tokenizados (.bas) e formato ASCII (.asc/.txt) via `LOAD "FILE",A`.

## msxRead (Visualizador de Arquivos)
O visualizador original continua integrado como uma ferramenta de suporte, acessível diretamente do editor.

![Screenshot do Visualizador msxRead](read-02.png)

- **Arquivos Graphos III:** Visualização de arquivos `.SHP` (Shapes), `.ALF` (Alfabeto), `.LAY` (Layout) e `.SCR` (Screen 2).
- **Leitor de Disco:** Interface para navegar em arquivos de diretórios que simulam discos MSX.
- **Hex Dump:** Visualização binária para arquivos desconhecidos.

## Tecnologias Utilizadas
- **Python 3.10+**
- **CustomTkinter:** Interface moderna e responsiva.
- **SQLite3:** Gerenciamento de configurações e análise de código.
- **Pillow:** Processamento de imagens.

## Instalação
Requer Python 3.10+ (testado no Windows).

```sh
python -m venv .venv
.\.venv\Scripts\activate
pip install -r requirements.txt
```

## Execução
```sh
python main.py
```

## Créditos e Inspirações
- **MSXBas2Rom:** [amaurycarvalho/msxbas2rom](https://github.com/amaurycarvalho/msxbas2rom)
- **Extensão MSX Text Encoding:** [nataliapc.msx-text-encoding](https://marketplace.visualstudio.com/items?itemName=nataliapc.msx-text-encoding)
- **Basic Dignified Suite:** [farique1/basic-dignified](https://github.com/farique1/basic-dignified)
- **MSX Converter:** [fgroen/msxconverter](https://github.com/fgroen/msxconverter)

## Ferramenta de IA usada
- **Junie (JetBrains AI Agent)**
