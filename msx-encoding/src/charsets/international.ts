import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_INTL, HIGH_CHARS_INTERNATIONAL } from './common';

/**
 * MSX International character set.
 * Default charset used by most MSX computers worldwide.
 * Based on CP437 with MSX-specific modifications in 0xB0-0xDF.
 */
export const msxInternational: MsxCharset = {
    id: 'msx-international',
    name: 'MSX International',
    description: 'Default MSX character set (CP437-based with MSX block elements)',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_INTL,
        HIGH_CHARS_INTERNATIONAL,
        { 0x7F: '\u2302' }  // ⌂ House
    ),
};
