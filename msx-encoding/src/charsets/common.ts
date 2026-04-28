/**
 * Common helpers and shared character arrays for MSX charset definitions.
 */

/**
 * Builds a 256-entry decode table from component parts.
 * 
 * @param graphicChars - 32 entries for positions 0x00-0x1F (graphic characters)
 * @param highChars - 128 entries for positions 0x80-0xFF (charset-specific)
 * @param overrides - Optional overrides for any position (e.g., 0x7F, 0x5C)
 * @returns Complete 256-entry decode table
 */
export function buildDecodeTable(
    graphicChars: string[],
    highChars: string[],
    overrides?: { [byte: number]: string }
): string[] {
    const table: string[] = new Array(256).fill('\uFFFD');

    // 0x00-0x1F: graphic characters
    for (let i = 0; i < 32 && i < graphicChars.length; i++) {
        table[i] = graphicChars[i];
    }

    // 0x20-0x7E: standard ASCII
    for (let i = 0x20; i <= 0x7E; i++) {
        table[i] = String.fromCharCode(i);
    }

    // 0x80-0xFF: charset-specific high chars
    for (let i = 0; i < 128 && i < highChars.length; i++) {
        table[0x80 + i] = highChars[i];
    }

    // Apply overrides
    if (overrides) {
        for (const [byteStr, char] of Object.entries(overrides)) {
            table[Number(byteStr)] = char;
        }
    }

    return table;
}

// ============================================================================
// Shared graphic character sets (0x00-0x1F)
// ============================================================================

/**
 * Graphic chars for International / Brazilian / Russian charsets.
 * CP437-derived symbols + MSX box drawings.
 */
export const GRAPHIC_CHARS_INTL: string[] = [
    // 0x00-0x07
    '\0',       // 0x00: NULL
    '\u263A',   // 0x01: ☺ White Smiling Face
    '\u263B',   // 0x02: ☻ Black Smiling Face
    '\u2665',   // 0x03: ♥ Black Heart Suit
    '\u2666',   // 0x04: ♦ Black Diamond Suit
    '\u2663',   // 0x05: ♣ Black Club Suit
    '\u2660',   // 0x06: ♠ Black Spade Suit
    '\u2022',   // 0x07: • Bullet
    // 0x08-0x0F
    '\u25D8',   // 0x08: ◘ Inverse Bullet
    '\u25CB',   // 0x09: ○ White Circle
    '\u25D9',   // 0x0A: ◙ Inverse White Circle
    '\u2642',   // 0x0B: ♂ Male Sign
    '\u2640',   // 0x0C: ♀ Female Sign
    '\u266A',   // 0x0D: ♪ Eighth Note
    '\u266B',   // 0x0E: ♫ Beamed Eighth Notes
    '\u263C',   // 0x0F: ☼ White Sun With Rays
    // 0x10-0x17: Box drawings
    '\u25BA',   // 0x10: ► Black Right-Pointing Pointer
    '\u2534',   // 0x11: ┴ Box Light Up And Horizontal
    '\u252C',   // 0x12: ┬ Box Light Down And Horizontal
    '\u2524',   // 0x13: ┤ Box Light Vertical And Left
    '\u251C',   // 0x14: ├ Box Light Vertical And Right
    '\u253C',   // 0x15: ┼ Box Light Vertical And Horizontal
    '\u2502',   // 0x16: │ Box Light Vertical
    '\u2500',   // 0x17: ─ Box Light Horizontal
    // 0x18-0x1F
    '\u250C',   // 0x18: ┌ Box Light Down And Right
    '\u2510',   // 0x19: ┐ Box Light Down And Left
    '\u2514',   // 0x1A: └ Box Light Up And Right
    '\u2518',   // 0x1B: ┘ Box Light Up And Left
    '\u2573',   // 0x1C: ╳ Box Light Diagonal Cross
    '\u2571',   // 0x1D: ╱ Box Light Diagonal Upper Right To Lower Left
    '\u2572',   // 0x1E: ╲ Box Light Diagonal Upper Left To Lower Right
    '\u{1FBAF}', // 0x1F: 🮯 Box Light Horizontal With Vertical Stroke
];

/**
 * Graphic chars for Japanese charset (0x00-0x1F).
 * Kanji for days/units + box drawings.
 */
export const GRAPHIC_CHARS_JP: string[] = [
    // 0x00-0x07: Kanji for days of week
    '\0',       // 0x00: NULL
    '\u6708',   // 0x01: 月 Monday/month
    '\u706B',   // 0x02: 火 Tuesday/fire
    '\u6C34',   // 0x03: 水 Wednesday/water
    '\u6728',   // 0x04: 木 Thursday/wood
    '\u91D1',   // 0x05: 金 Friday/metal
    '\u571F',   // 0x06: 土 Saturday/earth
    '\u65E5',   // 0x07: 日 Sunday/day
    // 0x08-0x0F: Kanji for units
    '\u5E74',   // 0x08: 年 year
    '\u5186',   // 0x09: 円 yen
    '\u6642',   // 0x0A: 時 hour
    '\u5206',   // 0x0B: 分 minute
    '\u79D2',   // 0x0C: 秒 second
    '\u767E',   // 0x0D: 百 hundred
    '\u5343',   // 0x0E: 千 thousand
    '\u4E07',   // 0x0F: 万 ten thousand
    // 0x10-0x17: π + box drawings (shared with intl 0x11-0x17)
    '\u03C0',   // 0x10: π Greek Small Letter Pi
    '\u2534',   // 0x11: ┴
    '\u252C',   // 0x12: ┬
    '\u2524',   // 0x13: ┤
    '\u251C',   // 0x14: ├
    '\u253C',   // 0x15: ┼
    '\u2502',   // 0x16: │
    '\u2500',   // 0x17: ─
    // 0x18-0x1F: More box drawings + Kanji
    '\u250C',   // 0x18: ┌
    '\u2510',   // 0x19: ┐
    '\u2514',   // 0x1A: └
    '\u2518',   // 0x1B: ┘
    '\u2573',   // 0x1C: ╳
    '\u5927',   // 0x1D: 大 big
    '\u4E2D',   // 0x1E: 中 middle
    '\u5C0F',   // 0x1F: 小 small
];

/**
 * Graphic chars for Arabic AR charset (0x00-0x1F).
 * 0x00-0x0F unmapped, 0x10-0x1C box drawings, 0x1D-0x1F unmapped.
 */
export const GRAPHIC_CHARS_AR: string[] = [
    // 0x00-0x0F: unmapped
    '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD',
    '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD',
    // 0x10-0x17: Box drawings (same as intl)
    '\u25BA',   // 0x10: ►
    '\u2534',   // 0x11: ┴
    '\u252C',   // 0x12: ┬
    '\u2524',   // 0x13: ┤
    '\u251C',   // 0x14: ├
    '\u253C',   // 0x15: ┼
    '\u2502',   // 0x16: │
    '\u2500',   // 0x17: ─
    // 0x18-0x1F
    '\u250C',   // 0x18: ┌
    '\u2510',   // 0x19: ┐
    '\u2514',   // 0x1A: └
    '\u2518',   // 0x1B: ┘
    '\u2573',   // 0x1C: ╳
    '\uFFFD',   // 0x1D: unmapped
    '\uFFFD',   // 0x1E: unmapped
    '\uFFFD',   // 0x1F: unmapped
];

/**
 * Graphic chars for Arabic AE charset (Al Alamiah AX-170) (0x00-0x1F).
 * 0x00-0x0F unmapped, 0x10-0x1D accented Latin, 0x1E-0x1F unmapped.
 */
export const GRAPHIC_CHARS_AE: string[] = [
    // 0x00-0x0F: unmapped
    '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD',
    '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD', '\uFFFD',
    // 0x10-0x1D: Accented Latin chars
    '\u00E9',   // 0x10: é
    '\u00E2',   // 0x11: â
    '\u00E0',   // 0x12: à
    '\u00E7',   // 0x13: ç
    '\u00EA',   // 0x14: ê
    '\u00EB',   // 0x15: ë
    '\u00E8',   // 0x16: è
    '\u00EF',   // 0x17: ï
    '\u00EE',   // 0x18: î
    '\u00F4',   // 0x19: ô
    '\u00FB',   // 0x1A: û
    '\u00F9',   // 0x1B: ù
    '\u00A7',   // 0x1C: §
    '\u00B0',   // 0x1D: °
    // 0x1E-0x1F: unmapped
    '\uFFFD',   // 0x1E: unmapped
    '\uFFFD',   // 0x1F: unmapped
];

// ============================================================================
// Shared high character ranges (0x80-0xFF)
// ============================================================================

/**
 * High chars for MSX International charset (0x80-0xFF).
 * CP437-based with MSX-specific modifications in 0xB0-0xDF.
 */
export const HIGH_CHARS_INTERNATIONAL: string[] = [
    // 0x80-0x8F: Accented Latin (same as CP437)
    '\u00C7',   // 0x80: Ç
    '\u00FC',   // 0x81: ü
    '\u00E9',   // 0x82: é
    '\u00E2',   // 0x83: â
    '\u00E4',   // 0x84: ä
    '\u00E0',   // 0x85: à
    '\u00E5',   // 0x86: å
    '\u00E7',   // 0x87: ç
    '\u00EA',   // 0x88: ê
    '\u00EB',   // 0x89: ë
    '\u00E8',   // 0x8A: è
    '\u00EF',   // 0x8B: ï
    '\u00EE',   // 0x8C: î
    '\u00EC',   // 0x8D: ì
    '\u00C4',   // 0x8E: Ä
    '\u00C5',   // 0x8F: Å
    // 0x90-0x9F: More Latin + symbols
    '\u00C9',   // 0x90: É
    '\u00E6',   // 0x91: æ
    '\u00C6',   // 0x92: Æ
    '\u00F4',   // 0x93: ô
    '\u00F6',   // 0x94: ö
    '\u00F2',   // 0x95: ò
    '\u00FB',   // 0x96: û
    '\u00F9',   // 0x97: ù
    '\u00FF',   // 0x98: ÿ
    '\u00D6',   // 0x99: Ö
    '\u00DC',   // 0x9A: Ü
    '\u00A2',   // 0x9B: ¢
    '\u00A3',   // 0x9C: £
    '\u00A5',   // 0x9D: ¥
    '\u20A7',   // 0x9E: ₧ Peseta Sign
    '\u0192',   // 0x9F: ƒ
    // 0xA0-0xAF: Latin + punctuation
    '\u00E1',   // 0xA0: á
    '\u00ED',   // 0xA1: í
    '\u00F3',   // 0xA2: ó
    '\u00FA',   // 0xA3: ú
    '\u00F1',   // 0xA4: ñ
    '\u00D1',   // 0xA5: Ñ
    '\u00AA',   // 0xA6: ª
    '\u00BA',   // 0xA7: º
    '\u00BF',   // 0xA8: ¿
    '\u2310',   // 0xA9: ⌐ Reversed Not Sign
    '\u00AC',   // 0xAA: ¬
    '\u00BD',   // 0xAB: ½
    '\u00BC',   // 0xAC: ¼
    '\u00A1',   // 0xAD: ¡
    '\u00AB',   // 0xAE: «
    '\u00BB',   // 0xAF: »
    // 0xB0-0xBF: MSX-specific (differs from CP437)
    '\u00C3',   // 0xB0: Ã
    '\u00E3',   // 0xB1: ã
    '\u0128',   // 0xB2: Ĩ
    '\u0129',   // 0xB3: ĩ
    '\u00D5',   // 0xB4: Õ
    '\u00F5',   // 0xB5: õ
    '\u0168',   // 0xB6: Ũ
    '\u0169',   // 0xB7: ũ
    '\u0132',   // 0xB8: Ĳ
    '\u0133',   // 0xB9: ĳ
    '\u00BE',   // 0xBA: ¾
    '\u223D',   // 0xBB: ∽ Reversed Tilde
    '\u25C7',   // 0xBC: ◇ White Diamond
    '\u2030',   // 0xBD: ‰ Per Mille Sign
    '\u00B6',   // 0xBE: ¶ Pilcrow
    '\u00A7',   // 0xBF: § Section Sign
    // 0xC0-0xCF: MSX block elements
    '\u2582',   // 0xC0: ▂ Lower One Quarter Block
    '\u259A',   // 0xC1: ▚ Quadrant Upper Left And Lower Right
    '\u2586',   // 0xC2: ▆ Lower Three Quarters Block
    '\u{1FB82}', // 0xC3: 🮂 Upper One Quarter Block
    '\u25AC',   // 0xC4: ▬ Black Rectangle
    '\u{1FB85}', // 0xC5: 🮅 Upper Three Quarters Block
    '\u258E',   // 0xC6: ▎ Left One Quarter Block
    '\u259E',   // 0xC7: ▞ Quadrant Upper Right And Lower Left
    '\u258A',   // 0xC8: ▊ Left Three Quarters Block
    '\u{1FB87}', // 0xC9: 🮇 Right One Quarter Block
    '\u{1FB8A}', // 0xCA: 🮊 Right Three Quarters Block
    '\u{1FB99}', // 0xCB: 🮙 Upper Right To Lower Left Fill
    '\u{1FB98}', // 0xCC: 🮘 Upper Left To Lower Right Fill
    '\u{1FB6D}', // 0xCD: 🭭 Upper Triangular One Quarter Block
    '\u{1FB6F}', // 0xCE: 🭯 Lower Triangular One Quarter Block
    '\u{1FB6C}', // 0xCF: 🭬 Left Triangular One Quarter Block
    // 0xD0-0xDF: More block elements
    '\u{1FB6E}', // 0xD0: 🭮 Right Triangular One Quarter Block
    '\u{1FB9A}', // 0xD1: 🮚 Upper And Lower Triangular Half Block
    '\u{1FB9B}', // 0xD2: 🮛 Left And Right Triangular Half Block
    '\u2598',   // 0xD3: ▘ Quadrant Upper Left
    '\u2597',   // 0xD4: ▗ Quadrant Lower Right
    '\u259D',   // 0xD5: ▝ Quadrant Upper Right
    '\u2596',   // 0xD6: ▖ Quadrant Lower Left
    '\u{1FB96}', // 0xD7: 🮖 Inverse Checker Board Fill
    '\u0394',   // 0xD8: Δ Greek Capital Delta
    '\u2021',   // 0xD9: ‡ Double Dagger
    '\u03C9',   // 0xDA: ω Greek Small Omega
    '\u2588',   // 0xDB: █ Full Block
    '\u2584',   // 0xDC: ▄ Lower Half Block
    '\u258C',   // 0xDD: ▌ Left Half Block
    '\u2590',   // 0xDE: ▐ Right Half Block
    '\u2580',   // 0xDF: ▀ Upper Half Block
    // 0xE0-0xEF: Math/Greek symbols
    '\u03B1',   // 0xE0: α alpha
    '\u00DF',   // 0xE1: ß sharp s
    '\u0393',   // 0xE2: Γ Gamma
    '\u03C0',   // 0xE3: π pi
    '\u03A3',   // 0xE4: Σ Sigma
    '\u03C3',   // 0xE5: σ sigma
    '\u00B5',   // 0xE6: µ micro
    '\u03C4',   // 0xE7: τ tau
    '\u03A6',   // 0xE8: Φ Phi
    '\u0398',   // 0xE9: Θ Theta
    '\u03A9',   // 0xEA: Ω Omega
    '\u03B4',   // 0xEB: δ delta
    '\u221E',   // 0xEC: ∞ Infinity
    '\u2205',   // 0xED: ∅ Empty Set
    '\u2208',   // 0xEE: ∈ Element Of
    '\u2229',   // 0xEF: ∩ Intersection
    // 0xF0-0xFF: More math symbols
    '\u2261',   // 0xF0: ≡ Identical To
    '\u00B1',   // 0xF1: ± Plus-Minus
    '\u2265',   // 0xF2: ≥ Greater-Than Or Equal To
    '\u2264',   // 0xF3: ≤ Less-Than Or Equal To
    '\u2320',   // 0xF4: ⌠ Top Half Integral
    '\u2321',   // 0xF5: ⌡ Bottom Half Integral
    '\u00F7',   // 0xF6: ÷ Division
    '\u2248',   // 0xF7: ≈ Almost Equal To
    '\u00B0',   // 0xF8: ° Degree
    '\u2219',   // 0xF9: ∙ Bullet Operator
    '\u00B7',   // 0xFA: · Middle Dot
    '\u221A',   // 0xFB: √ Square Root
    '\u207F',   // 0xFC: ⁿ Superscript N
    '\u00B2',   // 0xFD: ² Superscript Two
    '\u25A0',   // 0xFE: ■ Black Square
    '\uFFFD',   // 0xFF: unmapped (cursor)
];

/**
 * MSX block elements (0xC0-0xDF range from International).
 * Reused by Russian charset at positions 0x80-0x97.
 */
export const BLOCK_ELEMENTS_MSX: string[] = [
    '\u2582',   // ▂ Lower One Quarter Block
    '\u259A',   // ▚ Quadrant Upper Left And Lower Right
    '\u2586',   // ▆ Lower Three Quarters Block
    '\u{1FB82}', // 🮂 Upper One Quarter Block
    '\u25AC',   // ▬ Black Rectangle
    '\u{1FB85}', // 🮅 Upper Three Quarters Block
    '\u258E',   // ▎ Left One Quarter Block
    '\u259E',   // ▞ Quadrant Upper Right And Lower Left
    '\u258A',   // ▊ Left Three Quarters Block
    '\u{1FB87}', // 🮇 Right One Quarter Block
    '\u{1FB8A}', // 🮊 Right Three Quarters Block
    '\u{1FB99}', // 🮙 Upper Right To Lower Left Fill
    '\u{1FB98}', // 🮘 Upper Left To Lower Right Fill
    '\u{1FB6D}', // 🭭 Upper Triangular One Quarter Block
    '\u{1FB6F}', // 🭯 Lower Triangular One Quarter Block
    '\u{1FB6C}', // 🭬 Left Triangular One Quarter Block
    '\u{1FB6E}', // 🭮 Right Triangular One Quarter Block
    '\u{1FB9A}', // 🮚 Upper And Lower Triangular Half Block
    '\u{1FB9B}', // 🮛 Left And Right Triangular Half Block
    '\u2598',   // ▘ Quadrant Upper Left
    '\u2597',   // ▗ Quadrant Lower Right
    '\u259D',   // ▝ Quadrant Upper Right
    '\u2596',   // ▖ Quadrant Lower Left
    '\u{1FB96}', // 🮖 Inverse Checker Board Fill
];

/**
 * Math/Greek characters (0xE0-0xFE range from International).
 * Reused by Russian charset at positions 0xA0-0xBE.
 */
export const MATH_GREEK_CHARS: string[] = [
    '\u03B1',   // α alpha
    '\u00DF',   // ß sharp s
    '\u0393',   // Γ Gamma
    '\u03C0',   // π pi
    '\u03A3',   // Σ Sigma
    '\u03C3',   // σ sigma
    '\u00B5',   // µ micro
    '\u03C4',   // τ tau
    '\u03A6',   // Φ Phi
    '\u0398',   // Θ Theta
    '\u03A9',   // Ω Omega
    '\u03B4',   // δ delta
    '\u221E',   // ∞ Infinity
    '\u2205',   // ∅ Empty Set
    '\u2208',   // ∈ Element Of
    '\u2229',   // ∩ Intersection
    '\u2261',   // ≡ Identical To
    '\u00B1',   // ± Plus-Minus
    '\u2265',   // ≥ Greater-Than Or Equal
    '\u2264',   // ≤ Less-Than Or Equal
    '\u2320',   // ⌠ Top Half Integral
    '\u2321',   // ⌡ Bottom Half Integral
    '\u00F7',   // ÷ Division
    '\u2248',   // ≈ Almost Equal To
    '\u00B0',   // ° Degree
    '\u2219',   // ∙ Bullet Operator
    '\u00B7',   // · Middle Dot
    '\u221A',   // √ Square Root
    '\u207F',   // ⁿ Superscript N
    '\u00B2',   // ² Superscript Two
    '\u25A0',   // ■ Black Square
];
