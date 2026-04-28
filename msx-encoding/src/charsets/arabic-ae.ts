import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_AE } from './common';
import { msxArabicAr } from './arabic-ar';

/**
 * MSX Arabic AE character set (Al Alamiah AX-170).
 * 
 * Identical to Arabic AR in 0x80-0xFF range.
 * Differs in 0x10-0x1D: uses accented Latin characters instead of box drawings.
 */
export const msxArabicAe: MsxCharset = {
    id: 'msx-arabic-ae',
    name: 'MSX Arabic (AE)',
    description: 'Al Alamiah AX-170',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_AE,
        // High chars (0x80-0xFF) are identical to Arabic AR
        msxArabicAr.decodeTable.slice(0x80),
    ),
};
