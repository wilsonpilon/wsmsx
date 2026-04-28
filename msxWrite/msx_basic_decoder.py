from __future__ import annotations


TOKEN_MAP = [
    "END", "FOR", "NEXT", "DATA", "INPUT", "DIM", "READ", "LET", "GOTO", "RUN",
    "IF", "RESTORE", "GOSUB", "RETURN", "REM", "STOP", "PRINT", "CLEAR", "LIST",
    "NEW", "ON", "WAIT", "DEF", "POKE", "CONT", "CSAVE", "CLOAD", "OUT", "LPRINT",
    "LLIST", "CLS", "WIDTH", "ELSE", "TRON", "TROFF", "SWAP", "ERASE", "ERROR",
    "RESUME", "DELETE", "AUTO", "RENUM", "DEFSTR", "DEFINT", "DEFSNG", "DEFDBL",
    "LINE", "OPEN", "FIELD", "GET", "PUT", "CLOSE", "LOAD", "MERGE", "FILES",
    "LSET", "RSET", "SAVE", "LFILES", "CIRCLE", "COLOR", "DRAW", "PAINT", "BEEP",
    "PLAY", "PSET", "PRESET", "SOUND", "SCREEN", "VPOKE", "SPRITE", "VDP", "BASE",
    "CALL", "TIME", "KEY", "MAX", "MOTOR", "BLOAD", "BSAVE", "DSKO$",
    "SET", "NAME", "KILL", "IPL", "COPY", "CMD", "LOCATE",
    "TO", "THEN", "TAB(", "STEP", "USR", "FN", "SPC(", "NOT", "ERL", "ERR",
    "STRING$", "USING", "INSTR", "'", "VARPTR", "CSRLIN", "ATTR$", "DSKI$", "OFF",
    "INKEY$", "POINT", ">", "=", "<", "+", "-", "*", "/", "^", "AND", "OR", "XOR",
    "EQV", "IMP", "MOD", "\\",
]

TOKEN_MAP_FF = [
    "LEFT$", "RIGHT$", "MID$", "SGN", "INT", "ABS", "SQR", "RND", "SIN", "LOG",
    "EXP", "COS", "TAN", "ATN", "FRE", "INP", "POS", "LEN", "STR$", "VAL", "ASC",
    "CHR$", "PEEK", "VPEEK", "SPACES$", "OCT$", "HEX$", "LPOS", "BIN$", "CINT",
    "CSNG", "CDBL", "FIX", "STICK", "STRIG", "PDL", "PAD", "DSKF", "FPOS", "CVI",
    "CVS", "CVD", "EOF", "LOC", "LOF", "MKI$", "MK$", "MKD$",
]


def decode_msx_basic(data: bytes) -> str:
    segments = decode_msx_basic_segments(data)
    return "".join(text for _kind, text in segments)


def decode_msx_basic_segments(data: bytes) -> list[tuple[str, str]]:
    if not data:
        raise ValueError("invalid MSX Basic file: file is empty")
    if data[0] != 0xFF:
        raise ValueError(f"invalid MSX Basic file: expected 0xFF, got 0x{data[0]:02X}")

    result: list[tuple[str, str]] = []
    offset = 1

    def add_segment(kind: str, text: str) -> None:
        if not text:
            return
        if result and result[-1][0] == kind:
            result[-1] = (kind, result[-1][1] + text)
        else:
            result.append((kind, text))

    while True:
        if offset + 4 > len(data):
            break

        offset += 2  # next line address

        line_number = data[offset] + data[offset + 1] * 256
        offset += 2
        add_segment("line_number", str(line_number))
        add_segment("plain", " ")
        comment_mode = False

        while offset < len(data) and data[offset] != 0x00:
            token = data[offset]

            if comment_mode:
                if token >= 32:
                    add_segment("comment", chr(token))
                elif 17 <= token <= 26:
                    add_segment("comment", str(token - 17))
                offset += 1
                continue

            if token == 0x0B:
                if offset + 2 >= len(data):
                    break
                value = data[offset + 1] | (data[offset + 2] << 8)
                add_segment("number", f"&O{value:o}")
                offset += 2
            elif token == 0x0C:
                if offset + 2 >= len(data):
                    break
                value = data[offset + 1] | (data[offset + 2] << 8)
                add_segment("number", f"&H{value:X}")
                offset += 2
            elif token == 0x0E or token == 0x1C:
                if offset + 2 >= len(data):
                    break
                line = data[offset + 1] | (data[offset + 2] << 8)
                add_segment("number", str(line))
                offset += 2
            elif token == 0x0F:
                if offset + 1 >= len(data):
                    break
                value = data[offset + 1]
                add_segment("number", str(value))
                offset += 1
            elif token == 0x1D:
                if offset + 4 >= len(data):
                    break
                add_segment("number", custom_bcd_to_string(data[offset + 1:offset + 5]))
                offset += 4
            elif token == 0x3A:
                if offset + 1 < len(data):
                    next_token = data[offset + 1]
                    if next_token == 0x8F:
                        offset += 1
                    elif next_token == 0xA1:
                        pass
                    else:
                        add_segment("plain", chr(token))
                else:
                    add_segment("plain", chr(token))
            elif token == 0xFF:
                offset += 1
                if offset < len(data):
                    next_token = data[offset]
                    index = next_token - 0x81
                    if 0 <= index < len(TOKEN_MAP_FF):
                        add_segment("function", TOKEN_MAP_FF[index])
                    else:
                        add_segment("plain", f"-{next_token}-")
            elif token >= 0x80:
                index = token - 0x81
                if 0 <= index < len(TOKEN_MAP):
                    keyword = TOKEN_MAP[index]
                    if keyword == "REM":
                        add_segment("command", keyword)
                        comment_mode = True
                    elif keyword == "'":
                        add_segment("comment", keyword)
                        comment_mode = True
                    else:
                        add_segment("command", keyword)
                else:
                    add_segment("plain", f"-{token}-")
            elif token == 34:
                string_value = '"'
                offset += 1
                while offset < len(data):
                    token = data[offset]
                    string_value += chr(token)
                    if token == 34:
                        break
                    if offset + 1 < len(data) and data[offset + 1] == 0:
                        break
                    offset += 1
                add_segment("string", string_value)
            elif token >= 32:
                add_segment("plain", chr(token))
            elif 17 <= token <= 26:
                add_segment("number", str(token - 17))

            offset += 1

        if offset < len(data) and data[offset] == 0x00:
            add_segment("plain", "\n")
            offset += 1

        if offset + 1 < len(data) and data[offset] == 0x00 and data[offset + 1] == 0x00:
            break

    return result


def custom_bcd_to_string(b: bytes) -> str:
    if len(b) != 4:
        return ""

    sign = "-" if b[0] & 0x80 != 0 else ""
    exponent = (b[0] & 0x7F) - 64
    mantissa = f"{b[1]:02X}{b[2]:02X}{b[3]:02X}"

    mantissa_string = insert_decimal_point(mantissa, 1)
    mantissa_string = remove_trailing_zeros(mantissa_string)

    if exponent == -64:
        return "0!"
    if -63 <= exponent < -1:
        return f"{sign}{mantissa_string}E{exponent - 1:03d}"
    if exponent == -1:
        mantissa_string = ".0" + mantissa
        mantissa_string = remove_trailing_zeros(mantissa_string)
        return f"{sign}{mantissa_string}"
    if 0 <= exponent < 15:
        mantissa_string = shift_point_right(mantissa, exponent)
        mantissa_string = remove_trailing_zeros(mantissa_string)
        return f"{sign}{mantissa_string}"
    if 15 <= exponent <= 63:
        return f"{sign}{mantissa_string}E+{exponent - 1:02d}"
    return "XXX"


def insert_decimal_point(mantissa: str, pos: int) -> str:
    return mantissa[:pos] + "." + mantissa[pos:]


def shift_point_right(mantissa: str, shift: int) -> str:
    if len(mantissa) <= shift:
        return mantissa + ("0" * (shift - len(mantissa))) + "!"
    return mantissa[:shift] + "." + mantissa[shift:]


def remove_trailing_zeros(num_str: str) -> str:
    num_str = num_str.rstrip("0")
    return num_str.rstrip(".")
