/**
 * MSX Character Set definition.
 */
export interface MsxCharset {
    /** Unique identifier (e.g., 'msx-international') */
    id: string;
    /** Human-readable name */
    name: string;
    /** Description of the charset variant */
    description: string;
    /**
     * 256-entry decode table mapping byte values (0x00-0xFF) to Unicode characters.
     * 
     * - Positions 0x00-0x1F contain GRAPHIC characters (used via 0x01 escape in files).
     * - Positions 0x20-0x7E contain standard ASCII (or variant, e.g., ¥ at 0x5C for Japanese).
     * - Positions 0x80-0xFF contain charset-specific characters.
     * - Unmapped positions use '\uFFFD' (REPLACEMENT CHARACTER).
     */
    decodeTable: string[];
}
