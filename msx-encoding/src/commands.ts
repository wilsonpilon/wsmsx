import * as vscode from 'vscode';
import * as fs from 'fs';
import { getAllCharsets, getCharset, MsxCharset } from './charsets';
import { decode, encode } from './codec';
import { MsxFileSystemProvider } from './filesystem-provider';
import { MsxCharacterMap } from './character-map';

/**
 * QuickPick item for charset selection.
 */
interface CharsetQuickPickItem extends vscode.QuickPickItem {
    charsetId: string;
}

/**
 * Show a QuickPick to select an MSX charset.
 */
async function pickCharset(): Promise<MsxCharset | undefined> {
    const charsets = getAllCharsets();
    const items: CharsetQuickPickItem[] = charsets.map(cs => ({
        label: cs.name,
        description: cs.id,
        detail: cs.description,
        charsetId: cs.id,
    }));

    const selected = await vscode.window.showQuickPick(items, {
        placeHolder: 'Select MSX character set',
        title: 'MSX Encoding',
    });

    if (!selected) {
        return undefined;
    }

    return getCharset(selected.charsetId);
}

/**
 * Get the file URI from the command arguments or the active editor.
 */
function getFileUri(args: unknown[]): vscode.Uri | undefined {
    // From explorer context menu
    if (args.length > 0 && args[0] instanceof vscode.Uri) {
        return args[0];
    }
    // From active editor
    const editor = vscode.window.activeTextEditor;
    if (editor) {
        return editor.document.uri;
    }
    return undefined;
}

/**
 * Get the real filesystem path from a URI (handles both file:// and msxenc://).
 */
function getRealPath(uri: vscode.Uri): string {
    if (uri.scheme === 'msxenc') {
        return uri.path;
    }
    return uri.fsPath;
}

/**
 * Register all MSX Encoding commands.
 */
export function registerCommands(context: vscode.ExtensionContext): void {

    // ========================================================================
    // Open File with MSX Encoding
    // ========================================================================
    context.subscriptions.push(
        vscode.commands.registerCommand('msx-encoding.openWithEncoding', async (...args: unknown[]) => {
            let fileUri = getFileUri(args);

            // If no file provided, show file picker
            if (!fileUri) {
                const uris = await vscode.window.showOpenDialog({
                    canSelectMany: false,
                    title: 'Select file to open with MSX encoding',
                });
                if (!uris || uris.length === 0) { return; }
                fileUri = uris[0];
            }

            const charset = await pickCharset();
            if (!charset) { return; }

            const realPath = getRealPath(fileUri);
            const msxUri = MsxFileSystemProvider.createUri(realPath, charset.id);

            try {
                const doc = await vscode.workspace.openTextDocument(msxUri);
                await vscode.window.showTextDocument(doc);
            } catch (err) {
                vscode.window.showErrorMessage(`Failed to open file: ${err}`);
            }
        })
    );

    // ========================================================================
    // Reopen Active File with MSX Encoding
    // ========================================================================
    context.subscriptions.push(
        vscode.commands.registerCommand('msx-encoding.reopenWithEncoding', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showWarningMessage('No active editor');
                return;
            }

            const charset = await pickCharset();
            if (!charset) { return; }

            const realPath = getRealPath(editor.document.uri);
            const msxUri = MsxFileSystemProvider.createUri(realPath, charset.id);

            try {
                const doc = await vscode.workspace.openTextDocument(msxUri);
                await vscode.window.showTextDocument(doc);
            } catch (err) {
                vscode.window.showErrorMessage(`Failed to reopen file: ${err}`);
            }
        })
    );

    // ========================================================================
    // Convert MSX → UTF-8
    // ========================================================================
    context.subscriptions.push(
        vscode.commands.registerCommand('msx-encoding.convertToUtf8', async (...args: unknown[]) => {
            let fileUri = getFileUri(args);
            if (!fileUri) {
                const uris = await vscode.window.showOpenDialog({
                    canSelectMany: false,
                    title: 'Select MSX file to convert to UTF-8',
                });
                if (!uris || uris.length === 0) { return; }
                fileUri = uris[0];
            }

            const charset = await pickCharset();
            if (!charset) { return; }

            const realPath = getRealPath(fileUri);

            try {
                const rawBytes = fs.readFileSync(realPath);
                const text = decode(new Uint8Array(rawBytes), charset);

                // Write UTF-8 back to the same file
                fs.writeFileSync(realPath, text, 'utf-8');

                // Reopen the file normally
                const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(realPath));
                await vscode.window.showTextDocument(doc);

                vscode.window.showInformationMessage(
                    `Converted ${fileUri.fsPath} from ${charset.name} to UTF-8`
                );
            } catch (err) {
                vscode.window.showErrorMessage(`Conversion failed: ${err}`);
            }
        })
    );

    // ========================================================================
    // Convert UTF-8 → MSX Encoding
    // ========================================================================
    context.subscriptions.push(
        vscode.commands.registerCommand('msx-encoding.convertFromUtf8', async (...args: unknown[]) => {
            let fileUri = getFileUri(args);
            if (!fileUri) {
                const uris = await vscode.window.showOpenDialog({
                    canSelectMany: false,
                    title: 'Select UTF-8 file to convert to MSX encoding',
                });
                if (!uris || uris.length === 0) { return; }
                fileUri = uris[0];
            }

            const charset = await pickCharset();
            if (!charset) { return; }

            const realPath = getRealPath(fileUri);

            try {
                const text = fs.readFileSync(realPath, 'utf-8');
                const encoded = encode(text, charset);

                // Write MSX bytes back to the same file
                fs.writeFileSync(realPath, Buffer.from(encoded));

                vscode.window.showInformationMessage(
                    `Converted ${fileUri.fsPath} from UTF-8 to ${charset.name}`
                );
            } catch (err) {
                vscode.window.showErrorMessage(`Conversion failed: ${err}`);
            }
        })
    );

    // ========================================================================
    // Show MSX Character Map
    // ========================================================================
    const characterMap = new MsxCharacterMap(context);
    context.subscriptions.push(
        vscode.commands.registerCommand('msx-encoding.showCharacterMap', () => {
            // If a msxenc:// file is active, use its charset; otherwise use default
            const editor = vscode.window.activeTextEditor;
            let charsetId: string | undefined;
            if (editor && editor.document.uri.scheme === 'msxenc') {
                charsetId = editor.document.uri.authority;
            }
            characterMap.show(charsetId);
        })
    );
}
