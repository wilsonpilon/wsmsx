import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_INTL, BLOCK_ELEMENTS_MSX, MATH_GREEK_CHARS } from './common';

/**
 * MSX Russian character set.
 * Shares graphic chars 0x00-0x1F with International.
 * 0x80-0x97: Relocated MSX block elements (from International 0xC0-0xD7)
 * 0x98-0x9F: Δ ‡ ω █ ▄ ▌ ▐ ▀
 * 0xA0-0xBE: Relocated math/Greek (from International 0xE0-0xFE)
 * 0xBF: ¤ Currency Sign
 * 0xC0-0xDF: Cyrillic lowercase (JCUKEN order)
 * 0xE0-0xFE: Cyrillic uppercase (JCUKEN order)
 */
function buildRussianHighChars(): string[] {
    const chars: string[] = new Array(128).fill('\uFFFD');

    // 0x80-0x97 (indices 0-23): MSX block elements (same as International 0xC0-0xD7)
    for (let i = 0; i < BLOCK_ELEMENTS_MSX.length && i < 24; i++) {
        chars[i] = BLOCK_ELEMENTS_MSX[i];
    }

    // 0x98-0x9F (indices 24-31): Δ ‡ ω █ ▄ ▌ ▐ ▀
    chars[0x98 - 0x80] = '\u0394';   // Δ Greek Capital Delta
    chars[0x99 - 0x80] = '\u2021';   // ‡ Double Dagger
    chars[0x9A - 0x80] = '\u03C9';   // ω Greek Small Omega
    chars[0x9B - 0x80] = '\u2588';   // █ Full Block
    chars[0x9C - 0x80] = '\u2584';   // ▄ Lower Half Block
    chars[0x9D - 0x80] = '\u258C';   // ▌ Left Half Block
    chars[0x9E - 0x80] = '\u2590';   // ▐ Right Half Block
    chars[0x9F - 0x80] = '\u2580';   // ▀ Upper Half Block

    // 0xA0-0xBE (indices 32-62): Math/Greek (same as International 0xE0-0xFE)
    for (let i = 0; i < MATH_GREEK_CHARS.length && i < 31; i++) {
        chars[0xA0 - 0x80 + i] = MATH_GREEK_CHARS[i];
    }

    // 0xBF (index 63): ¤ Currency Sign
    chars[0xBF - 0x80] = '\u00A4';   // ¤

    // 0xC0-0xDF (indices 64-95): Cyrillic lowercase (JCUKEN order)
    const cyrillicLower = [
        '\u044E',   // 0xC0: ю
        '\u0430',   // 0xC1: а
        '\u0431',   // 0xC2: б
        '\u0446',   // 0xC3: ц
        '\u0434',   // 0xC4: д
        '\u0435',   // 0xC5: е
        '\u0444',   // 0xC6: ф
        '\u0433',   // 0xC7: г
        '\u0445',   // 0xC8: х
        '\u0438',   // 0xC9: и
        '\u0439',   // 0xCA: й
        '\u043A',   // 0xCB: к
        '\u043B',   // 0xCC: л
        '\u043C',   // 0xCD: м
        '\u043D',   // 0xCE: н
        '\u043E',   // 0xCF: о
        '\u043F',   // 0xD0: п
        '\u044F',   // 0xD1: я
        '\u0440',   // 0xD2: р
        '\u0441',   // 0xD3: с
        '\u0442',   // 0xD4: т
        '\u0443',   // 0xD5: у
        '\u0436',   // 0xD6: ж
        '\u0432',   // 0xD7: в
        '\u044C',   // 0xD8: ь
        '\u044B',   // 0xD9: ы
        '\u0437',   // 0xDA: з
        '\u0448',   // 0xDB: ш
        '\u044D',   // 0xDC: э
        '\u0449',   // 0xDD: щ
        '\u0447',   // 0xDE: ч
        '\u044A',   // 0xDF: ъ
    ];
    for (let i = 0; i < cyrillicLower.length; i++) {
        chars[0xC0 - 0x80 + i] = cyrillicLower[i];
    }

    // 0xE0-0xFE (indices 96-126): Cyrillic uppercase (JCUKEN order)
    const cyrillicUpper = [
        '\u042E',   // 0xE0: Ю
        '\u0410',   // 0xE1: А
        '\u0411',   // 0xE2: Б
        '\u0426',   // 0xE3: Ц
        '\u0414',   // 0xE4: Д
        '\u0415',   // 0xE5: Е
        '\u0424',   // 0xE6: Ф
        '\u0413',   // 0xE7: Г
        '\u0425',   // 0xE8: Х
        '\u0418',   // 0xE9: И
        '\u0419',   // 0xEA: Й
        '\u041A',   // 0xEB: К
        '\u041B',   // 0xEC: Л
        '\u041C',   // 0xED: М
        '\u041D',   // 0xEE: Н
        '\u041E',   // 0xEF: О
        '\u041F',   // 0xF0: П
        '\u042F',   // 0xF1: Я
        '\u0420',   // 0xF2: Р
        '\u0421',   // 0xF3: С
        '\u0422',   // 0xF4: Т
        '\u0423',   // 0xF5: У
        '\u0416',   // 0xF6: Ж
        '\u0412',   // 0xF7: В
        '\u042C',   // 0xF8: Ь
        '\u042B',   // 0xF9: Ы
        '\u0417',   // 0xFA: З
        '\u0428',   // 0xFB: Ш
        '\u042D',   // 0xFC: Э
        '\u0429',   // 0xFD: Щ
        '\u0427',   // 0xFE: Ч
    ];
    for (let i = 0; i < cyrillicUpper.length; i++) {
        chars[0xE0 - 0x80 + i] = cyrillicUpper[i];
    }

    // 0xFF (index 127): unmapped (cursor)
    chars[0xFF - 0x80] = '\uFFFD';

    return chars;
}

export const msxRussian: MsxCharset = {
    id: 'msx-russian',
    name: 'MSX Russian',
    description: 'Russian MSX character set (Cyrillic in JCUKEN layout order)',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_INTL,
        buildRussianHighChars(),
        { 0x7F: '\u2302' }  // ⌂ House
    ),
};
