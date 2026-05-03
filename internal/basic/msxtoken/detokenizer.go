package msxtoken

import (
	"fmt"
	"strconv"
	"strings"
)

var tokenMap = []string{
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
}

var tokenMapFF = []string{
	"LEFT$", "RIGHT$", "MID$", "SGN", "INT", "ABS", "SQR", "RND", "SIN", "LOG",
	"EXP", "COS", "TAN", "ATN", "FRE", "INP", "POS", "LEN", "STR$", "VAL", "ASC",
	"CHR$", "PEEK", "VPEEK", "SPACE$", "OCT$", "HEX$", "LPOS", "BIN$", "CINT",
	"CSNG", "CDBL", "FIX", "STICK", "STRIG", "PDL", "PAD", "DSKF", "FPOS", "CVI",
	"CVS", "CVD", "EOF", "LOC", "LOF", "MKI$", "MKS$", "MKD$",
}

// IsTokenizedProgram reports whether a file likely is an MSX tokenized BASIC file.
func IsTokenizedProgram(data []byte) bool {
	return len(data) > 0 && data[0] == 0xff
}

// DecodeProgramText decodes tokenized MSX-BASIC to editable text, or returns
// the original bytes as text for non-tokenized inputs.
func DecodeProgramText(data []byte) (string, bool, error) {
	if !IsTokenizedProgram(data) {
		return string(data), false, nil
	}
	text, err := detokenize(data)
	if err != nil {
		return "", true, err
	}
	return text, true, nil
}

func detokenize(data []byte) (string, error) {
	if len(data) == 0 || data[0] != 0xff {
		return "", fmt.Errorf("invalid tokenized program header")
	}

	offset := 1
	var out strings.Builder

	for {
		if offset+1 >= len(data) {
			break
		}
		if data[offset] == 0x00 && data[offset+1] == 0x00 {
			break
		}
		if offset+3 >= len(data) {
			return "", fmt.Errorf("unexpected end while reading line header")
		}

		_ = uint16(data[offset]) | uint16(data[offset+1])<<8 // next address
		lineNumber := uint16(data[offset+2]) | uint16(data[offset+3])<<8
		offset += 4

		out.WriteString(strconv.Itoa(int(lineNumber)))
		out.WriteByte(' ')
		commentMode := false

		for offset < len(data) && data[offset] != 0x00 {
			tok := data[offset]

			if commentMode {
				if tok >= 32 {
					out.WriteString(decodeMSXByte(tok))
				} else if tok >= 17 && tok <= 26 {
					out.WriteByte(byte('0' + (tok - 17)))
				}
				offset++
				continue
			}

			switch tok {
			case 0x0b: // octal
				if offset+2 >= len(data) {
					return "", fmt.Errorf("invalid octal token payload")
				}
				v := uint16(data[offset+1]) | uint16(data[offset+2])<<8
				out.WriteString(fmt.Sprintf("&O%o", v))
				offset += 3
				continue
			case 0x0c: // hex
				if offset+2 >= len(data) {
					return "", fmt.Errorf("invalid hex token payload")
				}
				v := uint16(data[offset+1]) | uint16(data[offset+2])<<8
				out.WriteString(fmt.Sprintf("&H%X", v))
				offset += 3
				continue
			case 0x0e, 0x1c: // line/integer word
				if offset+2 >= len(data) {
					return "", fmt.Errorf("invalid word token payload")
				}
				v := uint16(data[offset+1]) | uint16(data[offset+2])<<8
				out.WriteString(strconv.Itoa(int(v)))
				offset += 3
				continue
			case 0x0f:
				if offset+1 >= len(data) {
					return "", fmt.Errorf("invalid byte token payload")
				}
				out.WriteString(strconv.Itoa(int(data[offset+1])))
				offset += 2
				continue
			case 0x1d: // single precision (4-byte packed form)
				if offset+4 >= len(data) {
					return "", fmt.Errorf("invalid single token payload")
				}
				out.WriteString(customBCDToString(data[offset+1 : offset+5]))
				offset += 5
				continue
			case 0x1f: // double precision (8-byte packed form)
				if offset+8 >= len(data) {
					return "", fmt.Errorf("invalid double token payload")
				}
				out.WriteString(customBCDToString(data[offset+1 : offset+9]))
				offset += 9
				continue
			case 0xff:
				if offset+1 >= len(data) {
					return "", fmt.Errorf("invalid ff token payload")
				}
				idx := int(data[offset+1]) - 0x81
				if idx >= 0 && idx < len(tokenMapFF) {
					out.WriteString(tokenMapFF[idx])
				} else {
					out.WriteString(fmt.Sprintf("{FF%02X}", data[offset+1]))
				}
				offset += 2
				continue
			case 0x22:
				out.WriteByte('"')
				offset++
				for offset < len(data) {
					b := data[offset]
					out.WriteString(decodeMSXByte(b))
					offset++
					if b == 0x22 || offset >= len(data) || data[offset] == 0x00 {
						break
					}
				}
				continue
			default:
				if tok >= 0x80 {
					idx := int(tok) - 0x81
					if idx >= 0 && idx < len(tokenMap) {
						kw := tokenMap[idx]
						out.WriteString(kw)
						if kw == "REM" || kw == "'" {
							commentMode = true
						}
					} else {
						out.WriteString(fmt.Sprintf("{T%02X}", tok))
					}
				} else if tok >= 32 {
					out.WriteString(decodeMSXByte(tok))
				} else if tok >= 17 && tok <= 26 {
					out.WriteByte(byte('0' + (tok - 17)))
				}
				offset++
			}
		}

		if offset < len(data) && data[offset] == 0x00 {
			offset++
		}
		out.WriteString("\r\n")
	}

	return out.String(), nil
}

func customBCDToString(b []byte) string {
	if len(b) != 4 && len(b) != 8 {
		return "0"
	}

	isSingle := len(b) == 4
	expMark := 'E'
	suffix := "#"
	if isSingle {
		expMark = 'E'
		suffix = "!"
	} else {
		expMark = 'D'
		suffix = "#"
	}

	sign := ""
	if b[0]&0x80 != 0 {
		sign = "-"
	}

	exponent := int(b[0]&0x7f) - 64
	if exponent == -64 {
		return "0" + suffix
	}

	var mantissa strings.Builder
	for i := 1; i < len(b); i++ {
		mantissa.WriteString(fmt.Sprintf("%02X", b[i]))
	}
	digits := strings.TrimRight(mantissa.String(), "0")
	if digits == "" {
		return "0" + suffix
	}

	// Scientific canonical form (significant digits with explicit exponent)
	// when decimal point cannot be positioned cleanly in fixed form.
	if exponent <= 0 || exponent > len(digits) {
		sig := digits
		if len(sig) > 1 {
			sig = sig[:1] + "." + sig[1:]
			sig = trimTrailingDecimalZeros(sig)
		}
		return fmt.Sprintf("%s%s%c%+d", sign, sig, expMark, exponent-1)
	}

	fixed := digits
	if exponent < len(fixed) {
		fixed = fixed[:exponent] + "." + fixed[exponent:]
	} else if exponent > len(fixed) {
		fixed = fixed + strings.Repeat("0", exponent-len(fixed))
	}
	fixed = trimTrailingDecimalZeros(fixed)
	return sign + fixed + suffix
}

func trimTrailingDecimalZeros(v string) string {
	v = strings.TrimRight(v, "0")
	v = strings.TrimRight(v, ".")
	if v == "" || v == "-" {
		return "0"
	}
	return v
}
