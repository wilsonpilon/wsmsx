import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_INTL, HIGH_CHARS_INTERNATIONAL } from './common';

/**
 * MSX Brazilian BG character set.
 * Used by Gradiente Expert XP-800.
 * 
 * Differs from International by shuffling `, ~, ç, Ç positions
 * and using ₢ (Cruzeiro) + ≅ (approximately equal).
 */
function buildBrazilianBgHighChars(): string[] {
    const chars = [...HIGH_CHARS_INTERNATIONAL];
    // BG swaps: 0x80 gets ~ (tilde), 0x87 gets ` (backtick)
    chars[0x80 - 0x80] = '~';       // 0x80: ~ (was Ç — Ç moved to 0x7E in ASCII range)
    chars[0x87 - 0x80] = '`';       // 0x87: ` (was ç — ç moved to 0x60 in ASCII range)
    chars[0x9E - 0x80] = '\u20A2';  // 0x9E: ₢ Cruzeiro Sign (was ₧)
    chars[0xF0 - 0x80] = '\u2245';  // 0xF0: ≅ Approximately Equal To (was ≡)
    return chars;
}

export const msxBrazilianBg: MsxCharset = {
    id: 'msx-brazilian-bg',
    name: 'MSX Brazilian (BG)',
    description: 'Gradiente Expert XP-800',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_INTL,
        buildBrazilianBgHighChars(),
        {
            0x60: '\u00E7',  // 0x60: ç (was `)
            0x7E: '\u00C7',  // 0x7E: Ç (was ~)
            0x7F: '\u2302',  // ⌂ House
        }
    ),
};
