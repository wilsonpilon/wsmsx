import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import { getCharset } from './charsets';
import { decode, encode } from './codec';

/**
 * FileSystemProvider for MSX-encoded files.
 * 
 * URI format: msxenc://charset-id/absolute/path/to/file
 * - authority = charset ID (e.g., 'msx-international')
 * - path = absolute filesystem path (e.g., '/home/user/game.bas')
 * 
 * Transparently decodes MSX bytes → UTF-8 on read,
 * and encodes UTF-8 → MSX bytes on write.
 */
export class MsxFileSystemProvider implements vscode.FileSystemProvider {

    private _emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();
    readonly onDidChangeFile: vscode.Event<vscode.FileChangeEvent[]> = this._emitter.event;

    // --- Required interface methods ---

    watch(_uri: vscode.Uri, _options: { recursive: boolean; excludes: string[] }): vscode.Disposable {
        // Minimal watcher — no auto-refresh on external changes
        return new vscode.Disposable(() => { });
    }

    stat(uri: vscode.Uri): vscode.FileStat {
        const realPath = uri.path;
        try {
            const stat = fs.statSync(realPath);
            return {
                type: stat.isDirectory() ? vscode.FileType.Directory : vscode.FileType.File,
                ctime: stat.ctimeMs,
                mtime: stat.mtimeMs,
                size: stat.size,
            };
        } catch {
            throw vscode.FileSystemError.FileNotFound(uri);
        }
    }

    readDirectory(uri: vscode.Uri): [string, vscode.FileType][] {
        const realPath = uri.path;
        try {
            const entries = fs.readdirSync(realPath, { withFileTypes: true });
            return entries.map(entry => [
                entry.name,
                entry.isDirectory() ? vscode.FileType.Directory : vscode.FileType.File,
            ]);
        } catch {
            throw vscode.FileSystemError.FileNotFound(uri);
        }
    }

    createDirectory(uri: vscode.Uri): void {
        fs.mkdirSync(uri.path, { recursive: true });
    }

    readFile(uri: vscode.Uri): Uint8Array {
        const charsetId = uri.authority;
        const realPath = uri.path;

        const charset = getCharset(charsetId);
        if (!charset) {
            throw new Error(`Unknown MSX charset: ${charsetId}`);
        }

        let rawBytes: Buffer;
        try {
            rawBytes = fs.readFileSync(realPath);
        } catch {
            throw vscode.FileSystemError.FileNotFound(uri);
        }

        // Decode MSX bytes → Unicode text → UTF-8 bytes for VS Code
        const text = decode(new Uint8Array(rawBytes), charset);
        return Buffer.from(text, 'utf-8');
    }

    writeFile(uri: vscode.Uri, content: Uint8Array, _options: { create: boolean; overwrite: boolean }): void {
        const charsetId = uri.authority;
        const realPath = uri.path;

        const charset = getCharset(charsetId);
        if (!charset) {
            throw new Error(`Unknown MSX charset: ${charsetId}`);
        }

        // Decode UTF-8 bytes from VS Code → Unicode text → MSX bytes
        const text = Buffer.from(content).toString('utf-8');
        const encoded = encode(text, charset);

        // Ensure parent directory exists
        const dir = path.dirname(realPath);
        if (!fs.existsSync(dir)) {
            fs.mkdirSync(dir, { recursive: true });
        }

        fs.writeFileSync(realPath, Buffer.from(encoded));

        // Notify that the file changed
        this._emitter.fire([{ type: vscode.FileChangeType.Changed, uri }]);
    }

    delete(uri: vscode.Uri, _options: { recursive: boolean }): void {
        try {
            const stat = fs.statSync(uri.path);
            if (stat.isDirectory()) {
                fs.rmSync(uri.path, { recursive: true });
            } else {
                fs.unlinkSync(uri.path);
            }
        } catch {
            throw vscode.FileSystemError.FileNotFound(uri);
        }
    }

    rename(oldUri: vscode.Uri, newUri: vscode.Uri, _options: { overwrite: boolean }): void {
        try {
            fs.renameSync(oldUri.path, newUri.path);
        } catch {
            throw vscode.FileSystemError.FileNotFound(oldUri);
        }
    }

    // --- Helper methods ---

    /**
     * Create a msxenc:// URI for opening a file with a specific charset.
     */
    static createUri(filePath: string, charsetId: string): vscode.Uri {
        return vscode.Uri.from({
            scheme: 'msxenc',
            authority: charsetId,
            path: filePath,
        });
    }
}
