import { MsxCharset } from './types';
import { buildDecodeTable, GRAPHIC_CHARS_JP } from './common';

/**
 * MSX Japanese character set.
 * Based on JIS X 0201 for katakana, with hiragana and Kanji.
 * 0x5C = ¥ (Yen sign) instead of backslash.
 */
const HIGH_CHARS_JAPANESE: string[] = [
    // 0x80-0x8F: Card suits, circles, hiragana small chars
    '\u2660',   // 0x80: ♠ Black Spade Suit
    '\u2665',   // 0x81: ♥ Black Heart Suit
    '\u2663',   // 0x82: ♣ Black Club Suit
    '\u2666',   // 0x83: ♦ Black Diamond Suit
    '\u25CB',   // 0x84: ○ White Circle
    '\u25CF',   // 0x85: ● Black Circle
    '\u3092',   // 0x86: を
    '\u3041',   // 0x87: ぁ (small a)
    '\u3043',   // 0x88: ぃ (small i)
    '\u3045',   // 0x89: ぅ (small u)
    '\u3047',   // 0x8A: ぇ (small e)
    '\u3049',   // 0x8B: ぉ (small o)
    '\u3083',   // 0x8C: ゃ (small ya)
    '\u3085',   // 0x8D: ゅ (small yu)
    '\u3087',   // 0x8E: ょ (small yo)
    '\u3063',   // 0x8F: っ (small tsu)
    // 0x90-0x9F: Hiragana basic row
    '\uFFFD',   // 0x90: unmapped
    '\u3042',   // 0x91: あ a
    '\u3044',   // 0x92: い i
    '\u3046',   // 0x93: う u
    '\u3048',   // 0x94: え e
    '\u304A',   // 0x95: お o
    '\u304B',   // 0x96: か ka
    '\u304D',   // 0x97: き ki
    '\u304F',   // 0x98: く ku
    '\u3051',   // 0x99: け ke
    '\u3053',   // 0x9A: こ ko
    '\u3055',   // 0x9B: さ sa
    '\u3057',   // 0x9C: し shi
    '\u3059',   // 0x9D: す su
    '\u305B',   // 0x9E: せ se
    '\u305D',   // 0x9F: そ so
    // 0xA0-0xAF: Japanese punctuation + halfwidth katakana start
    '\u3000',   // 0xA0: Ideographic Space
    '\u3002',   // 0xA1: 。 Ideographic Full Stop
    '\u300C',   // 0xA2: 「 Left Corner Bracket
    '\u300D',   // 0xA3: 」 Right Corner Bracket
    '\u3001',   // 0xA4: 、 Ideographic Comma
    '\u30FB',   // 0xA5: ・ Katakana Middle Dot
    '\u30F2',   // 0xA6: ヲ
    '\u30A1',   // 0xA7: ァ (small a)
    '\u30A3',   // 0xA8: ィ (small i)
    '\u30A5',   // 0xA9: ゥ (small u)
    '\u30A7',   // 0xAA: ェ (small e)
    '\u30A9',   // 0xAB: ォ (small o)
    '\u30E3',   // 0xAC: ャ (small ya)
    '\u30E5',   // 0xAD: ュ (small yu)
    '\u30E7',   // 0xAE: ョ (small yo)
    '\u30C3',   // 0xAF: ッ (small tsu)
    // 0xB0-0xBF: Katakana
    '\u30FC',   // 0xB0: ー Prolonged Sound Mark
    '\u30A2',   // 0xB1: ア a
    '\u30A4',   // 0xB2: イ i
    '\u30A6',   // 0xB3: ウ u
    '\u30A8',   // 0xB4: エ e
    '\u30AA',   // 0xB5: オ o
    '\u30AB',   // 0xB6: カ ka
    '\u30AD',   // 0xB7: キ ki
    '\u30AF',   // 0xB8: ク ku
    '\u30B1',   // 0xB9: ケ ke
    '\u30B3',   // 0xBA: コ ko
    '\u30B5',   // 0xBB: サ sa
    '\u30B7',   // 0xBC: シ shi
    '\u30B9',   // 0xBD: ス su
    '\u30BB',   // 0xBE: セ se
    '\u30BD',   // 0xBF: ソ so
    // 0xC0-0xCF: More katakana
    '\u30BF',   // 0xC0: タ ta
    '\u30C1',   // 0xC1: チ chi
    '\u30C4',   // 0xC2: ツ tsu
    '\u30C6',   // 0xC3: テ te
    '\u30C8',   // 0xC4: ト to
    '\u30CA',   // 0xC5: ナ na
    '\u30CB',   // 0xC6: ニ ni
    '\u30CC',   // 0xC7: ヌ nu
    '\u30CD',   // 0xC8: ネ ne
    '\u30CE',   // 0xC9: ノ no
    '\u30CF',   // 0xCA: ハ ha
    '\u30D2',   // 0xCB: ヒ hi
    '\u30D5',   // 0xCC: フ fu
    '\u30D8',   // 0xCD: ヘ he
    '\u30DB',   // 0xCE: ホ ho
    '\u30DE',   // 0xCF: マ ma
    // 0xD0-0xDF: More katakana + voice/semi-voice marks
    '\u30DF',   // 0xD0: ミ mi
    '\u30E0',   // 0xD1: ム mu
    '\u30E1',   // 0xD2: メ me
    '\u30E2',   // 0xD3: モ mo
    '\u30E4',   // 0xD4: ヤ ya
    '\u30E6',   // 0xD5: ユ yu
    '\u30E8',   // 0xD6: ヨ yo
    '\u30E9',   // 0xD7: ラ ra
    '\u30EA',   // 0xD8: リ ri
    '\u30EB',   // 0xD9: ル ru
    '\u30EC',   // 0xDA: レ re
    '\u30ED',   // 0xDB: ロ ro
    '\u30EF',   // 0xDC: ワ wa
    '\u30F3',   // 0xDD: ン n
    '\u309B',   // 0xDE: ゛ Voiced Sound Mark (dakuten)
    '\u309C',   // 0xDF: ゜ Semi-Voiced Sound Mark (handakuten)
    // 0xE0-0xEF: Hiragana continued
    '\u305F',   // 0xE0: た ta
    '\u3061',   // 0xE1: ち chi
    '\u3064',   // 0xE2: つ tsu
    '\u3066',   // 0xE3: て te
    '\u3068',   // 0xE4: と to
    '\u306A',   // 0xE5: な na
    '\u306B',   // 0xE6: に ni
    '\u306C',   // 0xE7: ぬ nu
    '\u306D',   // 0xE8: ね ne
    '\u306E',   // 0xE9: の no
    '\u306F',   // 0xEA: は ha
    '\u3072',   // 0xEB: ひ hi
    '\u3075',   // 0xEC: ふ fu
    '\u3078',   // 0xED: へ he
    '\u307B',   // 0xEE: ほ ho
    '\u307E',   // 0xEF: ま ma
    // 0xF0-0xFF: Hiragana continued
    '\u307F',   // 0xF0: み mi
    '\u3080',   // 0xF1: む mu
    '\u3081',   // 0xF2: め me
    '\u3082',   // 0xF3: も mo
    '\u3084',   // 0xF4: や ya
    '\u3086',   // 0xF5: ゆ yu
    '\u3088',   // 0xF6: よ yo
    '\u3089',   // 0xF7: ら ra
    '\u308A',   // 0xF8: り ri
    '\u308B',   // 0xF9: る ru
    '\u308C',   // 0xFA: れ re
    '\u308D',   // 0xFB: ろ ro
    '\u308F',   // 0xFC: わ wa
    '\u3093',   // 0xFD: ん n
    '\uFFFD',   // 0xFE: unmapped
    '\uFFFD',   // 0xFF: unmapped (cursor)
];

export const msxJapanese: MsxCharset = {
    id: 'msx-japanese',
    name: 'MSX Japanese',
    description: 'Japanese MSX character set (JIS X 0201 katakana + hiragana)',
    decodeTable: buildDecodeTable(
        GRAPHIC_CHARS_JP,
        HIGH_CHARS_JAPANESE,
        { 0x5C: '\u00A5' }  // ¥ Yen sign (replaces backslash)
    ),
};
