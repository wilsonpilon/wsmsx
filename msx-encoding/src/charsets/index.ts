import { MsxCharset } from './types';

// Import all charset definitions
import { msxInternational } from './international';
import { msxBrazilianBr } from './brazilian-br';
import { msxBrazilianBg } from './brazilian-bg';
import { msxBrazilianBh } from './brazilian-bh';
import { msxArabicAr } from './arabic-ar';
import { msxArabicAe } from './arabic-ae';
import { msxJapanese } from './japanese';
import { msxRussian } from './russian';

export { MsxCharset } from './types';

/**
 * All available MSX charsets, indexed by ID.
 */
export const MSX_CHARSETS: Record<string, MsxCharset> = {
    [msxInternational.id]: msxInternational,
    [msxJapanese.id]: msxJapanese,
    [msxBrazilianBr.id]: msxBrazilianBr,
    [msxBrazilianBg.id]: msxBrazilianBg,
    [msxBrazilianBh.id]: msxBrazilianBh,
    [msxRussian.id]: msxRussian,
    [msxArabicAr.id]: msxArabicAr,
    [msxArabicAe.id]: msxArabicAe,
};

/**
 * Get a charset by its ID.
 * @returns The charset definition, or undefined if not found.
 */
export function getCharset(id: string): MsxCharset | undefined {
    return MSX_CHARSETS[id];
}

/**
 * Get all available charsets as an array of [id, charset] pairs.
 */
export function getAllCharsets(): MsxCharset[] {
    return Object.values(MSX_CHARSETS);
}
