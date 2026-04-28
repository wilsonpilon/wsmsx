import * as vscode from 'vscode';
import { MsxFileSystemProvider } from './filesystem-provider';
import { registerCommands } from './commands';
import { MsxStatusBar } from './status-bar';

/**
 * MSX Text Encoding extension entry point.
 * 
 * Registers:
 * - FileSystemProvider for msxenc:// scheme (transparent encode/decode)
 * - Commands for opening/converting files with MSX encodings
 * - Status bar item showing active MSX charset
 */
export function activate(context: vscode.ExtensionContext): void {
    // Register the MSX filesystem provider
    const provider = new MsxFileSystemProvider();
    context.subscriptions.push(
        vscode.workspace.registerFileSystemProvider('msxenc', provider, {
            isCaseSensitive: true,
            isReadonly: false,
        })
    );

    // Register commands
    registerCommands(context);

    // Register status bar
    const statusBar = new MsxStatusBar();
    statusBar.register(context);

    console.log('MSX Text Encoding extension activated');
}

export function deactivate(): void {
    // Nothing to clean up — all disposables are managed via context.subscriptions
}
