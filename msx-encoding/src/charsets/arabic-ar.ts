import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_AR } from './common';

/**
 * MSX Arabic AR character set.
 * Used by Bawareth Perfect MSX1 and Yamaha AX500.
 * 
 * Arabic letters use simplified mapping: each byte decodes to the abstract
 * Unicode letter. Multiple presentation forms (initial, medial, final, isolated)
 * that share the same abstract letter decode to the same Unicode character.
 * 
 * The Arabic Shaping Algorithm is NOT implemented — encoding from Unicode
 * will use a default byte for each abstract letter.
 */
const HIGH_CHARS_ARABIC: string[] = [
    // 0x80-0x8F: RTL punctuation + symbols
    ' ',        // 0x80: Space (RTL context)
    '!',        // 0x81
    '"',        // 0x82
    '#',        // 0x83
    '$',        // 0x84
    '\u066A',   // 0x85: ٪ Arabic Percent Sign
    '&',        // 0x86
    "'",        // 0x87
    '(',        // 0x88
    ')',        // 0x89
    '*',        // 0x8A
    '+',        // 0x8B
    '\u060C',   // 0x8C: ، Arabic Comma
    '-',        // 0x8D
    '.',        // 0x8E
    '/',        // 0x8F
    // 0x90-0x9F: Arabic-Indic digits + punctuation
    '\u0660',   // 0x90: ٠ Arabic-Indic Digit Zero
    '\u0661',   // 0x91: ١
    '\u0662',   // 0x92: ٢
    '\u0663',   // 0x93: ٣
    '\u0664',   // 0x94: ٤
    '\u0665',   // 0x95: ٥
    '\u0666',   // 0x96: ٦
    '\u0667',   // 0x97: ٧
    '\u0668',   // 0x98: ٨
    '\u0669',   // 0x99: ٩
    ':',        // 0x9A
    '\u061B',   // 0x9B: ؛ Arabic Semicolon
    '<',        // 0x9C
    '=',        // 0x9D
    '>',        // 0x9E
    '\u061F',   // 0x9F: ؟ Arabic Question Mark
    // 0xA0-0xAF: Arabic letters (abstract Unicode forms)
    '@',        // 0xA0: @ (RTL)
    '\u0626',   // 0xA1: ئ Yeh With Hamza Above (initial/medial)
    '\u0626',   // 0xA2: ئ Yeh With Hamza Above (isolated/final)
    '\u0628',   // 0xA3: ب Beh (initial/medial)
    '\u0628',   // 0xA4: ب Beh (isolated/final)
    '\u062A',   // 0xA5: ت Teh (initial/medial)
    '\u062A',   // 0xA6: ت Teh (isolated/final)
    '\u062B',   // 0xA7: ث Theh (initial/medial)
    '\u062B',   // 0xA8: ث Theh (isolated/final)
    '\u062C',   // 0xA9: ج Jeem (initial/medial)
    '\u062C',   // 0xAA: ج Jeem (isolated/final)
    '\u062D',   // 0xAB: ح Hah (initial/medial)
    '\u062D',   // 0xAC: ح Hah (isolated/final)
    '\u062E',   // 0xAD: خ Khah (initial/medial)
    '\u062E',   // 0xAE: خ Khah (isolated/final)
    '\u0633',   // 0xAF: س Seen (initial/medial)
    // 0xB0-0xBF: More Arabic letters + RTL brackets
    '\u0633',   // 0xB0: س Seen (isolated/final)
    '\u0634',   // 0xB1: ش Sheen (initial/medial)
    '\u0634',   // 0xB2: ش Sheen (isolated/final)
    '\u0635',   // 0xB3: ص Sad (initial/medial)
    '\u0635',   // 0xB4: ص Sad (isolated/final)
    '\u0636',   // 0xB5: ض Dad (initial/medial)
    '\u0636',   // 0xB6: ض Dad (isolated/final)
    '\u0637',   // 0xB7: ط Tah (all forms)
    '\u0638',   // 0xB8: ظ Zah (all forms)
    '\u0639',   // 0xB9: ع Ain (initial)
    '\u0639',   // 0xBA: ع Ain (isolated/final)
    '[',        // 0xBB: [ (RTL)
    '\\',       // 0xBC: \ (RTL)
    ']',        // 0xBD: ] (RTL)
    '^',        // 0xBE: ^ (RTL)
    '_',        // 0xBF: _ (RTL)
    // 0xC0-0xCF: More Arabic letters
    '\u0639',   // 0xC0: ع Ain (medial)
    '\u0639',   // 0xC1: ع Ain (final)
    '\u063A',   // 0xC2: غ Ghain (initial)
    '\u063A',   // 0xC3: غ Ghain (isolated)
    '\u063A',   // 0xC4: غ Ghain (medial)
    '\u063A',   // 0xC5: غ Ghain (final)
    '\u0641',   // 0xC6: ف Feh (initial/medial)
    '\u0641',   // 0xC7: ف Feh (isolated/final)
    '\u0642',   // 0xC8: ق Qaf (initial/medial)
    '\u0642',   // 0xC9: ق Qaf (isolated/final)
    '\u0643',   // 0xCA: ك Kaf (initial/medial)
    '\u0643',   // 0xCB: ك Kaf (isolated/final)
    '\u0644',   // 0xCC: ل Lam (initial/medial)
    '\u0644',   // 0xCD: ل Lam (isolated/final)
    '\u0645',   // 0xCE: م Meem (initial/medial)
    '\u0645',   // 0xCF: م Meem (isolated/final)
    // 0xD0-0xDF: More Arabic letters + RTL brackets
    '\u0646',   // 0xD0: ن Noon (initial/medial)
    '\u0646',   // 0xD1: ن Noon (isolated/final)
    '\u0647',   // 0xD2: ه Heh (initial/medial)
    '\u0647',   // 0xD3: ه Heh (isolated/final)
    '\u064A',   // 0xD4: ي Yeh (initial/medial)
    '\u064A',   // 0xD5: ي Yeh (isolated)
    '\u064A',   // 0xD6: ي Yeh (final)
    '\u0622',   // 0xD7: آ Alef With Madda Above (isolated)
    '\u0622',   // 0xD8: آ Alef With Madda Above (final)
    '\u0623',   // 0xD9: أ Alef With Hamza Above (isolated)
    '\u0623',   // 0xDA: أ Alef With Hamza Above (final)
    '{',        // 0xDB: { (RTL)
    '|',        // 0xDC: | (RTL)
    '}',        // 0xDD: } (RTL)
    '~',        // 0xDE: ~ (RTL)
    '\u0624',   // 0xDF: ؤ Waw With Hamza Above
    // 0xE0-0xEF: More Arabic letters + ligatures
    '\u0625',   // 0xE0: إ Alef With Hamza Below (isolated)
    '\u0625',   // 0xE1: إ Alef With Hamza Below (final)
    '\u0627',   // 0xE2: ا Alef (isolated)
    '\u0627',   // 0xE3: ا Alef (final)
    '\u0629',   // 0xE4: ة Teh Marbuta
    '\u062F',   // 0xE5: د Dal
    '\u0630',   // 0xE6: ذ Thal
    '\u0631',   // 0xE7: ر Reh
    '\u0632',   // 0xE8: ز Zain
    '\u0648',   // 0xE9: و Waw
    '\u0649',   // 0xEA: ى Alef Maksura (isolated)
    '\u0649',   // 0xEB: ى Alef Maksura (final)
    '\uFEFB',   // 0xEC: ﻻ Lam-Alef ligature (isolated)
    '\uFEF7',   // 0xED: ﻷ Lam-Alef With Hamza Above (isolated)
    '\uFEF5',   // 0xEE: ﻵ Lam-Alef With Madda Above (isolated)
    '\uFEF9',   // 0xEF: ﻹ Lam-Alef With Hamza Below (isolated)
    // 0xF0-0xFF: Standalone characters + diacritical marks
    '\u0621',   // 0xF0: ء Hamza
    '\u0640',   // 0xF1: ـ Tatweel (Kashida)
    '\u064B',   // 0xF2: ً Fathatan
    '\u064C',   // 0xF3: ٌ Dammatan
    '\u064D',   // 0xF4: ٍ Kasratan
    '\u064E',   // 0xF5: َ Fatha (isolated)
    '\u064E',   // 0xF6: َ Fatha (medial)
    '\u064F',   // 0xF7: ُ Damma (isolated)
    '\u064F',   // 0xF8: ُ Damma (medial)
    '\u0650',   // 0xF9: ِ Kasra (isolated)
    '\u0650',   // 0xFA: ِ Kasra (medial)
    '\u0651',   // 0xFB: ّ Shadda (isolated)
    '\u0651',   // 0xFC: ّ Shadda (medial)
    '\u0652',   // 0xFD: ْ Sukun (isolated)
    '\u0652',   // 0xFE: ْ Sukun (medial)
    '\uFFFD',   // 0xFF: unmapped (cursor)
];

export const msxArabicAr: MsxCharset = {
    id: 'msx-arabic-ar',
    name: 'MSX Arabic (AR)',
    description: 'Bawareth Perfect MSX1 / Yamaha AX500',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_AR,
        HIGH_CHARS_ARABIC,
    ),
};
