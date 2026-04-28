import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_INTL, HIGH_CHARS_INTERNATIONAL } from './common';

/**
 * MSX Brazilian BH character set.
 * Used by Sharp Hotbit 1.1.
 * 
 * Same Portuguese accented chars as BR (0x84-0x8F) but keeps ₧ (Peseta)
 * at 0x9E like International (BR uses ₢ Cruzeiro instead).
 */
function buildBrazilianBhHighChars(): string[] {
    const chars = [...HIGH_CHARS_INTERNATIONAL];
    // Same changes as BR at 0x84-0x8F
    chars[0x84 - 0x80] = '\u00C1';   // 0x84: Á (was ä)
    chars[0x86 - 0x80] = '\u00A8';   // 0x86: ¨ Diaeresis (was å)
    chars[0x89 - 0x80] = '\u00CD';   // 0x89: Í (was ë)
    chars[0x8A - 0x80] = '\u00D3';   // 0x8A: Ó (was è)
    chars[0x8B - 0x80] = '\u00DA';   // 0x8B: Ú (was ï)
    chars[0x8C - 0x80] = '\u00C2';   // 0x8C: Â (was î)
    chars[0x8D - 0x80] = '\u00CA';   // 0x8D: Ê (was ì)
    chars[0x8E - 0x80] = '\u00D4';   // 0x8E: Ô (was Ä)
    chars[0x8F - 0x80] = '\u00C0';   // 0x8F: À (was Å)
    // NOTE: 0x9E stays as ₧ (Peseta) — unlike BR which changes to ₢
    return chars;
}

export const msxBrazilianBh: MsxCharset = {
    id: 'msx-brazilian-bh',
    name: 'MSX Brazilian (BH)',
    description: 'Sharp Hotbit 1.1',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_INTL,
        buildBrazilianBhHighChars(),
        { 0x7F: '\u2302' }  // ⌂ House
    ),
};
