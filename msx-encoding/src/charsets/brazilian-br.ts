import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_INTL, HIGH_CHARS_INTERNATIONAL } from './common';

/**
 * MSX Brazilian BR character set.
 * Used by Gradiente Expert DDPlus and Sharp Hotbit 1.2.
 * Based on International with Portuguese-specific accented characters.
 */
function buildBrazilianBrHighChars(): string[] {
    const chars = [...HIGH_CHARS_INTERNATIONAL];
    // Differences from International:
    chars[0x84 - 0x80] = '\u00C1';   // 0x84: Á (was ä)
    chars[0x86 - 0x80] = '\u00A8';   // 0x86: ¨ Diaeresis (was å)
    chars[0x89 - 0x80] = '\u00CD';   // 0x89: Í (was ë)
    chars[0x8A - 0x80] = '\u00D3';   // 0x8A: Ó (was è)
    chars[0x8B - 0x80] = '\u00DA';   // 0x8B: Ú (was ï)
    chars[0x8C - 0x80] = '\u00C2';   // 0x8C: Â (was î)
    chars[0x8D - 0x80] = '\u00CA';   // 0x8D: Ê (was ì)
    chars[0x8E - 0x80] = '\u00D4';   // 0x8E: Ô (was Ä)
    chars[0x8F - 0x80] = '\u00C0';   // 0x8F: À (was Å)
    chars[0x9E - 0x80] = '\u20A2';   // 0x9E: ₢ Cruzeiro Sign (was ₧)
    return chars;
}

export const msxBrazilianBr: MsxCharset = {
    id: 'msx-brazilian-br',
    name: 'MSX Brazilian (BR)',
    description: 'Gradiente Expert DDPlus / Sharp Hotbit 1.2',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_INTL,
        buildBrazilianBrHighChars(),
        { 0x7F: '\u2302' }  // ⌂ House
    ),
};
