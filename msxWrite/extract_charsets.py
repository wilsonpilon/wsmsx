import os
import re
import json

def parse_ts_array(content, array_name):
    # Try to find array assignment
    pattern = rf'(?:export const|const) {array_name}: string\[\] = \[(.*?)\];'
    match = re.search(pattern, content, re.DOTALL)
    
    if not match:
        return None
    
    array_content = match.group(1)
    # Extract strings like '\u263A', '\0', '\x1FBAF', 'Ç', etc.
    chars = []
    # Simplified parser for the TS arrays
    items = re.findall(r"'((?:[^'\\]|\\.)*)'", array_content)
    for item in items:
        if item == '\\0':
            chars.append('\0')
        elif item.startswith('\\u{') and item.endswith('}'):
            hex_val = item[3:-1]
            chars.append(chr(int(hex_val, 16)))
        elif item.startswith('\\u'):
            hex_val = item[2:]
            chars.append(chr(int(hex_val, 16)))
        elif item.startswith('\\x'):
            hex_val = item[2:]
            chars.append(chr(int(hex_val, 16)))
        elif item == "\\'":
            chars.append("'")
        elif item == "\\\\":
            chars.append("\\")
        else:
            chars.append(item)
    return chars

def parse_brazilian_br(content, high_intl):
    chars = list(high_intl)
    # Extract assignments like: chars[0x84 - 0x80] = '\u00C1';
    matches = re.findall(r"chars\[(0x[0-9A-F]+) - 0x80\] = '((?:[^'\\]|\\.)*)';", content)
    for hex_pos, val in matches:
        pos = int(hex_pos, 16) - 0x80
        if val.startswith('\\u'):
            char_val = chr(int(val[2:], 16))
        else:
            char_val = val
        chars[pos] = char_val
    return chars

def extract_charsets():
    base_path = 'msx-encoding/src/charsets'
    common_content = open(os.path.join(base_path, 'common.ts'), 'r', encoding='utf-8').read()
    
    graphics = {
        'INTL': parse_ts_array(common_content, 'GRAPHIC_CHARS_INTL'),
        'JP': parse_ts_array(common_content, 'GRAPHIC_CHARS_JP'),
        'AE': parse_ts_array(common_content, 'GRAPHIC_CHARS_AE')
    }
    
    high_intl = parse_ts_array(common_content, 'HIGH_CHARS_INTERNATIONAL')
    
    charsets = {}
    
    # International
    intl_table = [''] * 256
    for i in range(32): intl_table[i] = graphics['INTL'][i]
    for i in range(0x20, 0x7F): intl_table[i] = chr(i)
    intl_table[0x7F] = '\u2302' # House override
    for i in range(min(128, len(high_intl))): intl_table[0x80 + i] = high_intl[i]
    charsets['International'] = intl_table

    # Japanese
    jp_content = open(os.path.join(base_path, 'japanese.ts'), 'r', encoding='utf-8').read()
    high_jp = parse_ts_array(jp_content, 'HIGH_CHARS_JAPANESE')
    jp_table = [''] * 256
    for i in range(32): jp_table[i] = graphics['JP'][i]
    for i in range(0x20, 0x7F): jp_table[i] = chr(i)
    jp_table[0x5C] = '\u00A5' # Yen override
    for i in range(min(128, len(high_jp))): jp_table[0x80 + i] = high_jp[i]
    charsets['Japanese'] = jp_table
    
    # Brazilian BR
    br_content = open(os.path.join(base_path, 'brazilian-br.ts'), 'r', encoding='utf-8').read()
    high_br = parse_brazilian_br(br_content, high_intl)
    br_table = [''] * 256
    for i in range(32): br_table[i] = graphics['INTL'][i]
    for i in range(0x20, 0x7F): br_table[i] = chr(i)
    br_table[0x7F] = '\u2302'
    for i in range(min(128, len(high_br))): br_table[0x80 + i] = high_br[i]
    charsets['Brazilian'] = br_table

    # Russian - reconstruct from common.ts and inline lists
    ru_content = open(os.path.join(base_path, 'russian.ts'), 'r', encoding='utf-8').read()
    block_msx = parse_ts_array(common_content, 'BLOCK_ELEMENTS_MSX')
    math_greek = parse_ts_array(common_content, 'MATH_GREEK_CHARS')
    # Build high chars 0x80-0xFF
    high_ru = ['\uFFFD'] * 128
    # 0x80-0x97 => BLOCK_ELEMENTS_MSX first 24
    for i in range(min(24, len(block_msx))):
        high_ru[0x80 - 0x80 + i] = block_msx[i]
    # 0x98-0x9F fixed symbols
    fixed = {
        0x98: '\u0394', 0x99: '\u2021', 0x9A: '\u03C9', 0x9B: '\u2588',
        0x9C: '\u2584', 0x9D: '\u258C', 0x9E: '\u2590', 0x9F: '\u2580'
    }
    # Decode escapes to actual characters
    for k, v in fixed.items():
        high_ru[k-0x80] = v.encode('utf-8').decode('unicode_escape')
    # 0xA0-0xBE => MATH_GREEK_CHARS first 31
    for i in range(min(31, len(math_greek))):
        high_ru[0xA0 - 0x80 + i] = math_greek[i]
    # 0xBF => currency sign
    high_ru[0xBF - 0x80] = '\u00A4'
    # 0xC0-0xDF cyrillic lower
    cyr_lower = [
        '\u044E','\u0430','\u0431','\u0446','\u0434','\u0435','\u0444','\u0433','\u0445','\u0438','\u0439','\u043A','\u043B','\u043C','\u043D','\u043E','\u043F','\u044F','\u0440','\u0441','\u0442','\u0443','\u0436','\u0432','\u044C','\u044B','\u0437','\u0448','\u044D','\u0449','\u0447','\u044A'
    ]
    for i, val in enumerate(cyr_lower):
        high_ru[0xC0 - 0x80 + i] = val.encode('utf-8').decode('unicode_escape')
    # 0xE0-0xFE cyrillic upper
    cyr_upper = [
        '\u042E','\u0410','\u0411','\u0426','\u0414','\u0415','\u0424','\u0413','\u0425','\u0418','\u0419','\u041A','\u041B','\u041C','\u041D','\u041E','\u041F','\u042F','\u0420','\u0421','\u0422','\u0423','\u0416','\u0412','\u042C','\u042B','\u0417','\u0428','\u042D','\u0429','\u0427'
    ]
    for i, val in enumerate(cyr_upper):
        high_ru[0xE0 - 0x80 + i] = val.encode('utf-8').decode('unicode_escape')
    # 0xFF unmapped
    high_ru[0xFF - 0x80] = '\uFFFD'

    ru_table = [''] * 256
    for i in range(32): ru_table[i] = graphics['INTL'][i]
    for i in range(0x20, 0x7F): ru_table[i] = chr(i)
    ru_table[0x7F] = '\u2302'
    for i in range(min(128, len(high_ru))):
        ru_table[0x80 + i] = high_ru[i]
    charsets['Russian'] = ru_table

    # Arabic (Ar)
    ar_content = open(os.path.join(base_path, 'arabic-ar.ts'), 'r', encoding='utf-8').read()
    high_ar = parse_ts_array(ar_content, 'HIGH_CHARS_ARABIC')
    if high_ar is None:
        # array is named const HIGH_CHARS_ARABIC in file scope, our regex supports it
        high_ar = parse_ts_array(ar_content, 'HIGH_CHARS_ARABIC')
    ar_table = [''] * 256
    graphics_ar = parse_ts_array(common_content, 'GRAPHIC_CHARS_AR')
    for i in range(32): ar_table[i] = graphics_ar[i]
    for i in range(0x20, 0x7F): ar_table[i] = chr(i)
    for i in range(min(128, len(high_ar))): ar_table[0x80 + i] = high_ar[i]
    charsets['Arabic'] = ar_table

    with open('msx_charsets.json', 'w', encoding='utf-8') as f:
        json.dump(charsets, f, ensure_ascii=False, indent=2)

if __name__ == '__main__':
    extract_charsets()
