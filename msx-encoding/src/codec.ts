import { MsxCharset } from './charsets';

/**
 * MSX Text Codec.
 * 
 * Handles encoding/decoding between MSX byte sequences and Unicode text.
 * 
 * Escape mechanism (0x01):
 * - In MSX interchange format, byte 0x01 acts as an escape prefix for
 *   graphic characters in the 0x00-0x1F VRAM range.
 * - The byte following 0x01 has 0x40 added as offset:
 *   0x01 0x4N → graphic character at VRAM position N (0x4N - 0x40 = 0x0N).
 *   Example: 0x01 0x54 → position 0x14 = ├
 * - Without 0x01 prefix, bytes 0x00-0x1F are treated as control codes.
 * 
 * Encoding direction:
 * - Unicode graphic chars that map to bytes 0x00-0x1F → emit [0x01, byte + 0x40]
 * - Unicode chars that map to bytes 0x20-0xFF → emit [byte]
 * - Common control codes (TAB, LF, CR) → emit as-is
 * - Unknown characters → emit 0x3F ('?')
 */

/** Escape byte used in MSX files for graphic characters in 0x00-0x1F range */
const ESCAPE_BYTE = 0x01;

/** Common control codes that should be preserved as-is (not escaped) */
const CONTROL_CODES = new Set([
    0x00,  // NULL
    0x07,  // BEL
    0x08,  // BS
    0x09,  // TAB
    0x0A,  // LF
    0x0D,  // CR
    0x1A,  // EOF (SUB)
    0x1B,  // ESC
]);

/**
 * Build a reverse mapping from Unicode character → byte value.
 * Used for encoding (Unicode → MSX bytes).
 * 
 * Only maps characters from 0x20-0xFF (high range).
 * Graphic chars (0x00-0x1F) are handled separately with escape prefix.
 */
export function buildEncodeMap(charset: MsxCharset): Map<string, number> {
    const map = new Map<string, number>();

    // Map all displayable chars (0x20-0xFF)
    // Iterate in order so later entries overwrite earlier ones
    // (for Arabic where multiple bytes map to same char, the later byte wins)
    for (let byte = 0x20; byte <= 0xFE; byte++) {
        const char = charset.decodeTable[byte];
        if (char && char !== '\uFFFD') {
            map.set(char, byte);
        }
    }

    return map;
}

/**
 * Build a reverse mapping for graphic characters (0x00-0x1F).
 * These require the 0x01 escape prefix when encoding.
 */
export function buildGraphicEncodeMap(charset: MsxCharset): Map<string, number> {
    const map = new Map<string, number>();

    for (let byte = 0x00; byte < 0x20; byte++) {
        const char = charset.decodeTable[byte];
        if (char && char !== '\uFFFD' && char !== '\0') {
            map.set(char, byte);
        }
    }

    return map;
}

/**
 * Decode MSX bytes to Unicode string.
 * 
 * @param data - Raw bytes from an MSX file
 * @param charset - The MSX charset to use for decoding
 * @returns Decoded Unicode string
 */
export function decode(data: Uint8Array, charset: MsxCharset): string {
    const result: string[] = [];
    let i = 0;

    while (i < data.length) {
        const byte = data[i];

        if (byte === ESCAPE_BYTE && i + 1 < data.length) {
            // Escape sequence: 0x01 + (charPos + 0x40) → graphic character
            // Subtract 0x40 offset to get the VRAM position (0x00-0x1F)
            const nextByte = data[i + 1];
            const charPos = nextByte - 0x40;
            if (charPos >= 0x00 && charPos <= 0x1F) {
                const char = charset.decodeTable[charPos];
                if (char && char !== '\uFFFD') {
                    result.push(char);
                } else {
                    result.push('\uFFFD');
                }
            } else {
                // Invalid escape sequence — emit replacement
                result.push('\uFFFD');
            }
            i += 2;
        } else if (byte < 0x20) {
            // Control code (no escape prefix) → pass through as Unicode control char
            result.push(String.fromCharCode(byte));
            i++;
        } else {
            // Regular byte (0x20-0xFF) → look up in decode table
            const char = charset.decodeTable[byte];
            if (char && char !== '\uFFFD') {
                result.push(char);
            } else {
                result.push('\uFFFD');
            }
            i++;
        }
    }

    return result.join('');
}

/**
 * Encode Unicode string to MSX bytes.
 * 
 * @param text - Unicode string to encode
 * @param charset - The MSX charset to use for encoding
 * @returns Encoded MSX bytes
 */
export function encode(text: string, charset: MsxCharset): Uint8Array {
    const encodeMap = buildEncodeMap(charset);
    const graphicMap = buildGraphicEncodeMap(charset);
    const result: number[] = [];

    // Iterate over Unicode codepoints (handles surrogate pairs correctly)
    for (const char of text) {
        const codePoint = char.codePointAt(0)!;

        // Check if it's a control code (0x00-0x1F)
        if (codePoint < 0x20) {
            if (codePoint === ESCAPE_BYTE) {
                // Skip bare 0x01 (SOH) — it's our escape byte and can't be
                // represented as a standalone byte without corruption
                continue;
            }
            // Common control codes pass through directly
            result.push(codePoint);
            continue;
        }

        // Check regular encode map first (0x20-0xFF)
        const byte = encodeMap.get(char);
        if (byte !== undefined) {
            result.push(byte);
            continue;
        }

        // Check graphic chars map (0x00-0x1F, needs escape prefix + 0x40 offset)
        const graphicByte = graphicMap.get(char);
        if (graphicByte !== undefined) {
            result.push(ESCAPE_BYTE);
            result.push(graphicByte + 0x40);
            continue;
        }

        // Unknown character → emit '?'
        result.push(0x3F);
    }

    return new Uint8Array(result);
}
