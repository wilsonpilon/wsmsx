import * as vscode from 'vscode';
import { getCharset } from './charsets';

/**
 * MSX Encoding Status Bar Manager.
 * 
 * Shows the current MSX charset in the status bar when editing files
 * opened via the msxenc:// scheme. Clicking the item opens the charset picker.
 */
export class MsxStatusBar {
    private statusBarItem: vscode.StatusBarItem;

    constructor() {
        this.statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Right,
            99  // Just before the built-in encoding indicator (priority 100)
        );
        this.statusBarItem.command = 'msx-encoding.reopenWithEncoding';
        this.statusBarItem.tooltip = 'MSX Text Encoding — Click to change';
    }

    /**
     * Update the status bar based on the active editor.
     */
    update(editor: vscode.TextEditor | undefined): void {
        if (!editor || editor.document.uri.scheme !== 'msxenc') {
            this.statusBarItem.hide();
            return;
        }

        const charsetId = editor.document.uri.authority;
        const charset = getCharset(charsetId);
        if (charset) {
            this.statusBarItem.text = `$(file-binary) ${charset.name}`;
            this.statusBarItem.show();
        } else {
            this.statusBarItem.hide();
        }
    }

    /**
     * Register event listeners and add to subscriptions.
     */
    register(context: vscode.ExtensionContext): void {
        context.subscriptions.push(this.statusBarItem);

        // Update on active editor change
        context.subscriptions.push(
            vscode.window.onDidChangeActiveTextEditor(editor => {
                this.update(editor);
            })
        );

        // Initial update
        this.update(vscode.window.activeTextEditor);
    }
}
