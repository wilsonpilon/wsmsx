import * as vscode from 'vscode';
import { getAllCharsets, getCharset, MsxCharset } from './charsets';

/**
 * Control code labels for bytes 0x00-0x1F.
 */
const CONTROL_LABELS: Record<number, string> = {
    0x00: 'NUL', 0x01: 'ESC¹', 0x02: 'STX', 0x03: 'ETX',
    0x04: 'EOT', 0x05: 'ENQ', 0x06: 'ACK', 0x07: 'BEL',
    0x08: 'BS',  0x09: 'TAB', 0x0A: 'LF',  0x0B: 'VT',
    0x0C: 'FF',  0x0D: 'CR',  0x0E: 'SO',  0x0F: 'SI',
    0x10: 'DLE', 0x11: 'DC1', 0x12: 'DC2', 0x13: 'DC3',
    0x14: 'DC4', 0x15: 'NAK', 0x16: 'SYN', 0x17: 'ETB',
    0x18: 'CAN', 0x19: 'EM',  0x1A: 'EOF', 0x1B: 'ESC',
    0x1C: 'FS',  0x1D: 'GS',  0x1E: 'RS',  0x1F: 'US',
};

/**
 * Build grid cell data from a charset for the webview.
 */
function buildGridData(charset: MsxCharset): object[] {
    const cells: object[] = [];
    for (let byte = 0; byte < 256; byte++) {
        const char = charset.decodeTable[byte];
        const isGraphic = byte < 0x20;
        const isUnmapped = char === '\uFFFD';
        const isControl = isGraphic; // 0x00-0x1F are control/graphic dual-use
        const controlLabel = CONTROL_LABELS[byte] || '';
        const displayChar = isUnmapped ? '' : char;

        cells.push({
            byte,
            char: displayChar,
            hex: byte.toString(16).toUpperCase().padStart(2, '0'),
            isGraphic,
            isUnmapped,
            isControl,
            controlLabel,
        });
    }
    return cells;
}

/**
 * MSX Character Map webview panel (singleton).
 */
export class MsxCharacterMap {
    private panel: vscode.WebviewPanel | undefined;
    private lastTextEditor: vscode.TextEditor | undefined;
    private currentCharsetId: string | undefined;
    private disposables: vscode.Disposable[] = [];

    constructor(private readonly context: vscode.ExtensionContext) {
        // Track the last active text editor (non-webview) for insertion
        this.lastTextEditor = vscode.window.activeTextEditor;
        const editorListener = vscode.window.onDidChangeActiveTextEditor(editor => {
            if (editor) {
                this.lastTextEditor = editor;
            }
        });
        this.disposables.push(editorListener);
        context.subscriptions.push(editorListener);
    }

    /**
     * Show the character map panel. Creates it if it doesn't exist, or reveals it.
     */
    show(initialCharsetId?: string): void {
        if (this.panel) {
            this.panel.reveal(vscode.ViewColumn.Beside);
            if (initialCharsetId && initialCharsetId !== this.currentCharsetId) {
                this.updateCharset(initialCharsetId);
            }
            return;
        }

        const charsetId = initialCharsetId || this.getDefaultCharsetId();
        const charset = getCharset(charsetId);
        if (!charset) {
            vscode.window.showErrorMessage(`Unknown charset: ${charsetId}`);
            return;
        }

        this.currentCharsetId = charsetId;

        this.panel = vscode.window.createWebviewPanel(
            'msxCharacterMap',
            'MSX Character Map',
            vscode.ViewColumn.Beside,
            {
                enableScripts: true,
                retainContextWhenHidden: true,
            }
        );

        this.panel.webview.html = this.getWebviewHtml(charset);

        // Handle messages from the webview
        this.panel.webview.onDidReceiveMessage(
            message => this.handleMessage(message),
            undefined,
            this.disposables
        );

        // Clean up on dispose
        this.panel.onDidDispose(() => {
            this.panel = undefined;
            this.currentCharsetId = undefined;
        }, null, this.disposables);
    }

    /**
     * Handle messages received from the webview.
     */
    private async handleMessage(message: { command: string; text?: string; charsetId?: string }): Promise<void> {
        switch (message.command) {
            case 'insert': {
                const text = message.text || '';
                if (!text) { return; }

                const editor = this.lastTextEditor;
                if (!editor) {
                    vscode.window.showWarningMessage('No active text editor to insert into');
                    return;
                }

                // Check if the document is still open
                try {
                    await editor.edit(editBuilder => {
                        if (editor.selection.isEmpty) {
                            editBuilder.insert(editor.selection.active, text);
                        } else {
                            editBuilder.replace(editor.selection, text);
                        }
                    });
                } catch {
                    vscode.window.showWarningMessage('Could not insert text — the target editor may have been closed');
                }
                break;
            }

            case 'changeCharset': {
                const charsetId = message.charsetId;
                if (charsetId) {
                    this.updateCharset(charsetId);
                }
                break;
            }

            case 'copy': {
                const copyText = message.text || '';
                if (copyText) {
                    await vscode.env.clipboard.writeText(copyText);
                }
                break;
            }
        }
    }

    /**
     * Update the webview with a new charset's data.
     */
    private updateCharset(charsetId: string): void {
        const charset = getCharset(charsetId);
        if (!charset || !this.panel) { return; }

        this.currentCharsetId = charsetId;
        const gridData = buildGridData(charset);
        this.panel.webview.postMessage({
            command: 'updateGrid',
            gridData,
            charsetId,
            charsetName: charset.name,
        });
    }

    /**
     * Get the default charset ID from settings.
     */
    private getDefaultCharsetId(): string {
        return vscode.workspace.getConfiguration('msxEncoding')
            .get<string>('defaultCharset', 'msx-international');
    }

    /**
     * Generate the full webview HTML.
     */
    private getWebviewHtml(charset: MsxCharset): string {
        const charsets = getAllCharsets();
        const charsetOptions = charsets.map(cs =>
            `<option value="${cs.id}"${cs.id === charset.id ? ' selected' : ''}>${cs.name}</option>`
        ).join('\n');

        const gridData = buildGridData(charset);
        const gridDataJson = JSON.stringify(gridData);

        return /*html*/`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline';">
    <title>MSX Character Map</title>
    <style>
        :root {
            --cell-size: 36px;
            --header-size: 28px;
        }
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: var(--vscode-font-family, 'Segoe UI', Tahoma, sans-serif);
            color: var(--vscode-editor-foreground);
            background-color: var(--vscode-editor-background);
            padding: 12px;
            user-select: none;
        }

        /* Header / charset selector */
        .header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 12px;
            flex-wrap: wrap;
        }
        .header label {
            font-size: 13px;
            font-weight: 600;
        }
        .header select {
            font-family: var(--vscode-font-family, 'Segoe UI', Tahoma, sans-serif);
            font-size: 13px;
            color: var(--vscode-dropdown-foreground);
            background-color: var(--vscode-dropdown-background);
            border: 1px solid var(--vscode-dropdown-border);
            padding: 4px 8px;
            border-radius: 2px;
            outline: none;
            cursor: pointer;
        }
        .header select:focus {
            border-color: var(--vscode-focusBorder);
        }

        /* Grid container */
        .grid-container {
            overflow-x: auto;
            margin-bottom: 14px;
        }
        .grid {
            display: grid;
            grid-template-columns: var(--header-size) repeat(16, var(--cell-size));
            grid-template-rows: var(--header-size) repeat(16, var(--cell-size));
            gap: 1px;
            width: fit-content;
        }

        /* Header cells (row/column labels) */
        .grid-header {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 11px;
            font-weight: 700;
            font-family: var(--vscode-editor-font-family, 'Consolas', 'Courier New', monospace);
            color: var(--vscode-descriptionForeground);
            background-color: var(--vscode-editorGroupHeader-tabsBackground);
        }
        .grid-corner {
            background-color: transparent;
        }

        /* Character cells */
        .cell {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 16px;
            font-family: var(--vscode-editor-font-family, 'Consolas', 'Courier New', monospace);
            background-color: var(--vscode-editor-background);
            border: 1px solid var(--vscode-editorWidget-border, rgba(128,128,128,0.2));
            cursor: pointer;
            position: relative;
            transition: background-color 0.1s;
        }
        .cell:hover {
            background-color: var(--vscode-list-hoverBackground);
            border-color: var(--vscode-focusBorder);
            z-index: 1;
        }
        .cell:active {
            background-color: var(--vscode-list-activeSelectionBackground);
            color: var(--vscode-list-activeSelectionForeground);
        }

        /* Graphic chars (0x00-0x1F) — blue tint */
        .cell.graphic {
            background-color: color-mix(in srgb, var(--vscode-charts-blue, #1e90ff) 10%, var(--vscode-editor-background));
            border-color: color-mix(in srgb, var(--vscode-charts-blue, #1e90ff) 30%, var(--vscode-editorWidget-border, rgba(128,128,128,0.2)));
        }
        .cell.graphic:hover {
            background-color: color-mix(in srgb, var(--vscode-charts-blue, #1e90ff) 25%, var(--vscode-list-hoverBackground));
        }

        /* Unmapped chars */
        .cell.unmapped {
            color: var(--vscode-disabledForeground, rgba(128,128,128,0.4));
            cursor: default;
            font-size: 10px;
        }
        .cell.unmapped:hover {
            background-color: var(--vscode-editor-background);
            border-color: var(--vscode-editorWidget-border, rgba(128,128,128,0.2));
        }

        /* Delete (0x7F) */
        .cell.delete-char {
            font-size: 12px;
            color: var(--vscode-descriptionForeground);
        }

        /* Tooltip */
        .tooltip {
            display: none;
            position: absolute;
            bottom: calc(100% + 6px);
            left: 50%;
            transform: translateX(-50%);
            background-color: var(--vscode-editorHoverWidget-background, #2d2d30);
            color: var(--vscode-editorHoverWidget-foreground, #cccccc);
            border: 1px solid var(--vscode-editorHoverWidget-border, #454545);
            border-radius: 3px;
            padding: 4px 8px;
            font-size: 11px;
            white-space: nowrap;
            z-index: 100;
            pointer-events: none;
            box-shadow: 0 2px 8px rgba(0,0,0,0.3);
        }
        .cell:hover .tooltip {
            display: block;
        }
        .tooltip .tt-char {
            font-size: 18px;
            margin-right: 6px;
            vertical-align: middle;
        }
        .tooltip .tt-hex {
            font-family: var(--vscode-editor-font-family, monospace);
            color: var(--vscode-charts-yellow, #dcdcaa);
        }
        .tooltip .tt-unicode {
            font-family: var(--vscode-editor-font-family, monospace);
            color: var(--vscode-charts-green, #6a9955);
            margin-left: 6px;
        }
        .tooltip .tt-label {
            color: var(--vscode-charts-blue, #569cd6);
            margin-left: 6px;
        }

        /* Accumulator area */
        .accumulator {
            display: flex;
            align-items: stretch;
            gap: 6px;
            margin-top: 4px;
        }
        .acc-field {
            flex: 1;
            font-family: var(--vscode-editor-font-family, 'Consolas', 'Courier New', monospace);
            font-size: 15px;
            color: var(--vscode-input-foreground);
            background-color: var(--vscode-input-background);
            border: 1px solid var(--vscode-input-border, rgba(128,128,128,0.4));
            padding: 6px 10px;
            border-radius: 2px;
            outline: none;
            min-height: 36px;
        }
        .acc-field:focus {
            border-color: var(--vscode-focusBorder);
        }
        .acc-buttons {
            display: flex;
            gap: 4px;
        }
        .btn {
            font-family: var(--vscode-font-family, 'Segoe UI', Tahoma, sans-serif);
            font-size: 12px;
            color: var(--vscode-button-foreground);
            background-color: var(--vscode-button-background);
            border: none;
            padding: 6px 14px;
            border-radius: 2px;
            cursor: pointer;
            white-space: nowrap;
        }
        .btn:hover {
            background-color: var(--vscode-button-hoverBackground);
        }
        .btn.secondary {
            color: var(--vscode-button-secondaryForeground);
            background-color: var(--vscode-button-secondaryBackground);
        }
        .btn.secondary:hover {
            background-color: var(--vscode-button-secondaryHoverBackground);
        }

        /* Legend */
        .legend {
            margin-top: 10px;
            font-size: 11px;
            color: var(--vscode-descriptionForeground);
            display: flex;
            gap: 16px;
            flex-wrap: wrap;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 4px;
        }
        .legend-swatch {
            width: 14px;
            height: 14px;
            border-radius: 2px;
            border: 1px solid var(--vscode-editorWidget-border, rgba(128,128,128,0.3));
        }
        .swatch-graphic {
            background-color: color-mix(in srgb, var(--vscode-charts-blue, #1e90ff) 15%, var(--vscode-editor-background));
        }
        .swatch-normal {
            background-color: var(--vscode-editor-background);
        }
        .swatch-unmapped {
            background-color: var(--vscode-editor-background);
            opacity: 0.4;
        }
    </style>
</head>
<body>
    <div class="header">
        <label for="charset-select">Charset:</label>
        <select id="charset-select">
            ${charsetOptions}
        </select>
    </div>

    <div class="grid-container">
        <div class="grid" id="char-grid"></div>
    </div>

    <div class="accumulator">
        <input type="text" class="acc-field" id="acc-field" placeholder="Click characters to accumulate..." readonly />
        <div class="acc-buttons">
            <button class="btn" id="btn-insert" title="Insert accumulated text into the active editor">Insert</button>
            <button class="btn secondary" id="btn-copy" title="Copy to clipboard">Copy</button>
            <button class="btn secondary" id="btn-clear" title="Clear accumulated text">Clear</button>
            <button class="btn secondary" id="btn-backspace" title="Delete last character">⌫</button>
        </div>
    </div>

    <div class="legend">
        <div class="legend-item">
            <div class="legend-swatch swatch-graphic"></div>
            <span>Graphic chars (0x00-0x1F, escaped with 0x01)</span>
        </div>
        <div class="legend-item">
            <div class="legend-swatch swatch-normal"></div>
            <span>Standard chars</span>
        </div>
        <div class="legend-item">
            <div class="legend-swatch swatch-unmapped"></div>
            <span>Unmapped</span>
        </div>
    </div>

    <script>
        (function() {
            const vscode = acquireVsCodeApi();
            const HEX = '0123456789ABCDEF';

            let gridData = ${gridDataJson};
            const accField = document.getElementById('acc-field');

            function buildGrid() {
                const grid = document.getElementById('char-grid');
                grid.innerHTML = '';

                // Corner cell
                const corner = document.createElement('div');
                corner.className = 'grid-header grid-corner';
                grid.appendChild(corner);

                // Column headers (high nibble: _0, _1, ..., _F)
                for (let col = 0; col < 16; col++) {
                    const hdr = document.createElement('div');
                    hdr.className = 'grid-header';
                    hdr.textContent = HEX[col] + '_';
                    grid.appendChild(hdr);
                }

                // Rows
                for (let row = 0; row < 16; row++) {
                    // Row header (low nibble: _0, _1, ..., _F)
                    const rowHdr = document.createElement('div');
                    rowHdr.className = 'grid-header';
                    rowHdr.textContent = '_' + HEX[row];
                    grid.appendChild(rowHdr);

                    for (let col = 0; col < 16; col++) {
                        const idx = col * 16 + row;
                        const d = gridData[idx];
                        const cell = document.createElement('div');

                        let classes = 'cell';
                        if (d.isUnmapped) {
                            classes += ' unmapped';
                        } else if (d.isGraphic) {
                            classes += ' graphic';
                        }

                        // Special: 0x7F (DEL)
                        if (d.byte === 0x7F) {
                            classes += ' delete-char';
                        }

                        cell.className = classes;

                        // Display content
                        if (d.isUnmapped) {
                            cell.textContent = '·';
                        } else if (d.byte === 0x7F) {
                            cell.textContent = d.char || 'DEL';
                        } else {
                            cell.textContent = d.char;
                        }

                        // Tooltip
                        const tooltip = document.createElement('div');
                        tooltip.className = 'tooltip';
                        let tooltipHtml = '<span class="tt-hex">0x' + d.hex + '</span>';
                        if (!d.isUnmapped && d.char) {
                            const cp = d.char.codePointAt(0);
                            const cpHex = cp !== undefined ? cp.toString(16).toUpperCase().padStart(4, '0') : '';
                            tooltipHtml = '<span class="tt-char">' + escapeHtml(d.char) + '</span>' + tooltipHtml;
                            tooltipHtml += '<span class="tt-unicode">U+' + cpHex + '</span>';
                        }
                        if (d.isGraphic && d.controlLabel) {
                            tooltipHtml += '<span class="tt-label">' + d.controlLabel + '</span>';
                        }
                        if (d.isUnmapped) {
                            tooltipHtml += '<span class="tt-label">unmapped</span>';
                        }
                        tooltip.innerHTML = tooltipHtml;
                        cell.appendChild(tooltip);

                        // Click handler — add char to accumulator
                        if (!d.isUnmapped && d.char) {
                            cell.addEventListener('click', () => {
                                accField.value += d.char;
                            });
                        }

                        grid.appendChild(cell);
                    }
                }
            }

            function escapeHtml(str) {
                return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
            }

            // Charset selector
            document.getElementById('charset-select').addEventListener('change', function() {
                vscode.postMessage({ command: 'changeCharset', charsetId: this.value });
            });

            // Insert button
            document.getElementById('btn-insert').addEventListener('click', () => {
                const text = accField.value;
                if (text) {
                    vscode.postMessage({ command: 'insert', text: text });
                    accField.value = '';
                }
            });

            // Copy button
            document.getElementById('btn-copy').addEventListener('click', () => {
                const text = accField.value;
                if (text) {
                    navigator.clipboard.writeText(text).catch(() => {
                        // Fallback: send to extension to copy
                        vscode.postMessage({ command: 'copy', text: text });
                    });
                }
            });

            // Clear button
            document.getElementById('btn-clear').addEventListener('click', () => {
                accField.value = '';
            });

            // Backspace button
            document.getElementById('btn-backspace').addEventListener('click', () => {
                const text = accField.value;
                if (text.length > 0) {
                    // Handle surrogate pairs
                    const chars = Array.from(text);
                    chars.pop();
                    accField.value = chars.join('');
                }
            });

            // Receive messages from extension
            window.addEventListener('message', event => {
                const msg = event.data;
                if (msg.command === 'updateGrid') {
                    gridData = msg.gridData;
                    // Update charset selector
                    const sel = document.getElementById('charset-select');
                    sel.value = msg.charsetId;
                    buildGrid();
                }
            });

            // Initial build
            buildGrid();
        })();
    </script>
</body>
</html>`;
    }

    /**
     * Dispose all resources.
     */
    dispose(): void {
        this.panel?.dispose();
        for (const d of this.disposables) {
            d.dispose();
        }
        this.disposables = [];
    }
}
