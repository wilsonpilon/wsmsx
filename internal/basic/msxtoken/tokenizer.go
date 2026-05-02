package msxtoken

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const baseAddress = 0x8001

var (
	bigNine    = big.NewInt(9)
	big255     = big.NewInt(255)
	big32767   = big.NewInt(32767)
	big32768   = big.NewInt(32768)
	big999999  = big.NewInt(999999)
	big1000000 = big.NewInt(1000000)
	bigMax63   = mustBigInt("999999999999999999999999999999999999999999999999999999999999999")
	bigOne     = big.NewInt(1)

	reLineNumber  = regexp.MustCompile(`^\s*\d+\s?`)
	reJumpNumbers = regexp.MustCompile(`^(\s*)(\d+|,+)`)
	reAS          = regexp.MustCompile(`^(\s*)(\d{1,2})`)
	reHexDigits   = regexp.MustCompile(`^[0-9a-f]*`)
	reOctDigits   = regexp.MustCompile(`^[0-7]*`)
	reBinDigits   = regexp.MustCompile(`^[01]*`)
	reNumInitial  = regexp.MustCompile(`^(\d*)\s*(.)\s*(.?)`)
	reNumFloat    = regexp.MustCompile(`^(\d*)\s*(.)\s*(\d*)\s*(.)\s*(.?)`)
	reNumExp      = regexp.MustCompile(`^\d*\s*.\s*\d*\s*.\s*(\+|-)\s*(\d+)`)
)

// TokenizeProgram converts MSX-BASIC ASCII source into tokenized binary bytes.
func TokenizeProgram(source string) ([]byte, error) {
	t := &tokenizer{}
	lines := splitProgramLines(source)
	return t.tokenize(lines)
}

type tokenSpec struct {
	cmd string
	hex string
}

var tokens = []tokenSpec{
	{">", "ee"}, {"PAINT", "bf"}, {"=", "ef"}, {"ERROR", "a6"}, {"ERR", "e2"}, {"<", "f0"}, {"+", "f1"},
	{"FIELD", "b1"}, {"PLAY", "c1"}, {"-", "f2"}, {"FILES", "b7"}, {"POINT", "ed"}, {"*", "f3"}, {"POKE", "98"},
	{"/", "f4"}, {"FN", "de"}, {"^", "f5"}, {"FOR", "82"}, {"PRESET", "c3"}, {"\\", "fc"}, {"PRINT", "91"}, {"?", "91"},
	{"PSET", "c2"}, {"AND", "f6"}, {"GET", "b2"}, {"PUT", "b3"}, {"GOSUB", "8d"}, {"READ", "87"}, {"GOTO", "89"},
	{"ATTR$", "e9"}, {"RENUM", "aa"}, {"AUTO", "a9"}, {"IF", "8b"}, {"RESTORE", "8c"}, {"BASE", "c9"}, {"IMP", "fa"},
	{"RESUME", "a7"}, {"BEEP", "c0"}, {"INKEY$", "ec"}, {"RETURN", "8e"}, {"BLOAD", "cf"}, {"INPUT", "85"},
	{"BSAVE", "d0"}, {"INSTR", "e5"}, {"RSET", "b9"}, {"CALL", "ca"}, {"_", "5f"}, {"RUN", "8a"}, {"IPL", "d5"}, {"SAVE", "ba"},
	{"KEY", "cc"}, {"SCREEN", "c5"}, {"KILL", "d4"}, {"SET", "d2"}, {"CIRCLE", "bc"}, {"CLEAR", "92"}, {"CLOAD", "9b"},
	{"LET", "88"}, {"SOUND", "c4"}, {"CLOSE", "b4"}, {"LFILES", "bb"}, {"CLS", "9f"}, {"LINE", "af"}, {"SPC(", "df"},
	{"CMD", "d7"}, {"LIST", "93"}, {"SPRITE", "c7"}, {"COLOR", "bd"}, {"LLIST", "9e"}, {"CONT", "99"}, {"LOAD", "b5"},
	{"STEP", "dc"}, {"COPY", "d6"}, {"LOCATE", "d8"}, {"STOP", "90"}, {"CSAVE", "9a"}, {"CSRLIN", "e8"},
	{"STRING$", "e3"}, {"LPRINT", "9d"}, {"SWAP", "a4"}, {"LSET", "b8"}, {"TAB(", "db"}, {"MAX", "cd"}, {"DATA", "84"},
	{"MERGE", "b6"}, {"THEN", "da"}, {"TIME", "cb"}, {"TO", "d9"}, {"DEFDBL", "ae"}, {"DEFINT", "ac"}, {"DEFSTR", "ab"},
	{"TROFF", "a3"}, {"DEFSNG", "ad"}, {"TRON", "a2"}, {"DEF", "97"}, {"MOD", "fb"}, {"USING", "e4"},
	{"DELETE", "a8"}, {"MOTOR", "ce"}, {"USR", "dd"}, {"DIM", "86"}, {"NAME", "d3"}, {"DRAW", "be"}, {"NEW", "94"},
	{"VARPTR", "e7"}, {"NEXT", "83"}, {"VDP", "c8"}, {"DSKI$", "ea"}, {"NOT", "e0"}, {"DSKO$", "d1"}, {"VPOKE", "c6"},
	{"OFF", "eb"}, {"WAIT", "96"}, {"END", "81"}, {"ON", "95"}, {"WIDTH", "a0"}, {"OPEN", "b0"}, {"XOR", "f8"},
	{"EQV", "f9"}, {"OR", "f7"}, {"ERASE", "a5"}, {"OUT", "9c"}, {"ERL", "e1"}, {"REM", "8f"},
	{"PDL", "ffa4"}, {"EXP", "ff8b"}, {"PEEK", "ff97"}, {"FIX", "ffa1"}, {"POS", "ff91"}, {"FPOS", "ffa7"},
	{"ABS", "ff86"}, {"FRE", "ff8f"}, {"ASC", "ff95"}, {"ATN", "ff8e"}, {"HEX$", "ff9b"}, {"BIN$", "ff9d"},
	{"INP", "ff90"}, {"RIGHT$", "ff82"}, {"RND", "ff88"}, {"INT", "ff85"}, {"CDBL", "ffa0"}, {"CHR$", "ff96"},
	{"CINT", "ff9e"}, {"LEFT$", "ff81"}, {"SGN", "ff84"}, {"LEN", "ff92"}, {"SIN", "ff89"}, {"SPACE$", "ff99"},
	{"SQR", "ff87"}, {"LOC(", "ffac28"}, {"STICK", "ffa2"}, {"COS", "ff8c"}, {"LOF", "ffad"}, {"STR$", "ff93"},
	{"CSNG", "ff9f"}, {"LOG", "ff8a"}, {"STRIG", "ffa3"}, {"LPOS", "ff9c"}, {"CVD", "ffaa"}, {"CVI", "ffa8"},
	{"CVS", "ffa9"}, {"TAN", "ff8d"}, {"MID$", "ff83"}, {"MKD$", "ffb0"}, {"MKI$", "ffae"}, {"MKS$", "ffaf"},
	{"VAL", "ff94"}, {"DSKF", "ffa6"}, {"VPEEK", "ff98"}, {"OCT$", "ff9a"}, {"EOF", "ffab"}, {"PAD", "ffa5"},
	{"'", "3a8fe6"}, {"ELSE", "3aa1"}, {"AS", "4153"},
}

var jumps = map[string]struct{}{
	"RESTORE": {}, "AUTO": {}, "RENUM": {}, "DELETE": {}, "RESUME": {}, "ERL": {}, "ELSE": {},
	"RUN": {}, "LIST": {}, "LLIST": {}, "GOTO": {}, "RETURN": {}, "THEN": {}, "GOSUB": {},
}

type tokenizer struct {
	source       int
	compiled     string
	lineCompiled string
	lineSource   string
}

func splitProgramLines(source string) []string {
	raw := strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		trim := strings.TrimSpace(line)
		if trim == "" || isAllDigits(trim) {
			continue
		}
		lines = append(lines, strings.TrimRight(line, " \t")+"\r\n")
	}
	return lines
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func mustBigInt(s string) *big.Int {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid big int literal")
	}
	return v
}

func parseBig(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return big.NewInt(0)
	}
	return v
}

func (t *tokenizer) updateLines() {
	if len(t.lineSource) > 2 {
		t.lineSource = t.lineSource[t.source:]
		t.lineCompiled += t.compiled
	}
}

func (t *tokenizer) tokenize(lines []string) ([]byte, error) {
	lineAddress := baseAddress
	lineOrder := 0
	tokenizedHex := []string{"ff"}

	for idx, line := range lines {
		t.lineSource = line
		t.lineCompiled = ""

		if len(t.lineSource) == 0 {
			continue
		}
		if t.lineSource[0] < '0' || t.lineSource[0] > '9' {
			if t.lineSource[0] == 26 {
				continue
			}
			return nil, fmt.Errorf("line %d: line not starting with number", idx+1)
		}

		lineNumberPart := reLineNumber.FindString(t.lineSource)
		if lineNumberPart == "" {
			return nil, fmt.Errorf("line %d: missing line number", idx+1)
		}
		lineNumberStr := strings.TrimSpace(lineNumberPart)
		lineNumber, err := strconv.Atoi(lineNumberStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid line number %q", idx+1, lineNumberStr)
		}
		if lineNumber <= lineOrder {
			return nil, fmt.Errorf("line %d: line number out of order: %d", idx+1, lineNumber)
		}
		if lineNumber > 65529 {
			return nil, fmt.Errorf("line %d: line number too high: %d", idx+1, lineNumber)
		}
		lineOrder = lineNumber
		t.lineSource = t.lineSource[len(lineNumberPart):]
		t.lineCompiled += littleEndianWordHex(lineNumber)

		for len(t.lineSource) > 2 {
			if ok, err := t.tryToken(idx + 1); err != nil {
				return nil, err
			} else if ok {
				continue
			}
			if err := t.tryAtom(idx + 1); err != nil {
				return nil, err
			}
		}

		lineAddress += (len(t.lineCompiled) + 6) / 2
		tokenizedHex = append(tokenizedHex, littleEndianWordHex(lineAddress)+t.lineCompiled+"00")
	}

	tokenizedHex = append(tokenizedHex, "0000")
	out := make([]byte, 0, len(tokenizedHex)*16)
	for _, h := range tokenizedHex {
		b, err := hex.DecodeString(h)
		if err != nil {
			return nil, fmt.Errorf("invalid generated token stream: %w", err)
		}
		out = append(out, b...)
	}
	return out, nil
}

func (t *tokenizer) tryToken(lineNum int) (bool, error) {
	upper := strings.ToUpper(t.lineSource)
	for _, tk := range tokens {
		if !strings.HasPrefix(upper, tk.cmd) {
			continue
		}
		t.compiled = tk.hex
		t.source = len(tk.cmd)
		t.updateLines()

		if tk.cmd == "AS" {
			if m := reAS.FindStringSubmatch(t.lineSource); len(m) == 3 {
				spaces := strings.Repeat("20", len(m[1]))
				n, _ := strconv.Atoi(m[2])
				t.compiled = spaces + fmt.Sprintf("%02x", byte(n))
				t.source = len(m[1]) + len(m[2])
				t.updateLines()
			}
		}

		if _, isJump := jumps[tk.cmd]; isJump {
			for {
				m := reJumpNumbers.FindStringSubmatch(t.lineSource)
				if len(m) != 3 {
					break
				}
				spaces := strings.Repeat("20", len(m[1]))
				val := m[2]
				if isAllDigits(val) {
					jump, _ := strconv.Atoi(val)
					if jump > 65529 {
						return false, fmt.Errorf("line %d: jump line too high: %s", lineNum, val)
					}
					t.compiled = spaces + "0e" + littleEndianWordHex(jump)
				} else {
					t.compiled = spaces + strings.Repeat("2c", len(val))
				}
				t.source = len(m[1]) + len(val)
				t.updateLines()
			}
		}

		if tk.cmd == "DATA" || tk.cmd == "REM" || tk.cmd == "'" || tk.cmd == "CALL" || tk.cmd == "_" {
			for {
				if len(t.lineSource) == 0 {
					break
				}
				ch := t.lineSource[0]
				if tk.cmd == "CALL" || tk.cmd == "_" {
					ch = toUpperASCII(ch)
				}
				t.compiled = fmt.Sprintf("%02x", ch)
				t.source = 1
				t.updateLines()

				if len(t.lineSource) <= 2 ||
					(tk.cmd == "DATA" && len(t.lineSource) > 0 && t.lineSource[0] == ':') ||
					((tk.cmd == "_" || tk.cmd == "CALL") && len(t.lineSource) > 0 && (t.lineSource[0] == ':' || t.lineSource[0] == '(')) {
					break
				}
			}
		}
		return true, nil
	}
	return false, nil
}

func (t *tokenizer) tryAtom(lineNum int) error {
	if len(t.lineSource) == 0 {
		return nil
	}
	first := toUpperASCII(t.lineSource[0])

	if (first >= '0' && first <= '9') || first == '.' {
		if ok, err := t.tryNumber(lineNum); ok || err != nil {
			return err
		}
	}
	if first == '&' {
		return t.tryBaseNumber(lineNum)
	}

	if first == '"' {
		numQuotes := 0
		for {
			if len(t.lineSource) == 0 {
				break
			}
			if t.lineSource[0] == '"' {
				numQuotes++
			}
			t.compiled = fmt.Sprintf("%02x", t.lineSource[0])
			t.source = 1
			t.updateLines()
			if numQuotes > 1 || len(t.lineSource) <= 2 {
				break
			}
		}
		return nil
	}

	if first >= 'A' && first <= 'Z' {
		for {
			if len(t.lineSource) == 0 {
				break
			}
			n := toUpperASCII(t.lineSource[0])
			if !((n >= '0' && n <= '9') || (n >= 'A' && n <= 'Z')) {
				break
			}
			upper := strings.ToUpper(t.lineSource)
			isVar := true
			for _, tk := range tokens {
				if strings.HasPrefix(upper, tk.cmd) {
					isVar = false
					break
				}
			}
			if !isVar {
				break
			}
			t.compiled = fmt.Sprintf("%02x", n)
			t.source = 1
			t.updateLines()
		}
		return nil
	}

	t.compiled = fmt.Sprintf("%02x", first)
	t.source = 1
	t.updateLines()
	return nil
}

func (t *tokenizer) tryNumber(lineNum int) (bool, error) {
	m := reNumInitial.FindStringSubmatch(t.lineSource)
	if len(m) != 4 {
		return false, nil
	}
	nuggetNumber := m[1]
	nuggetInteger := m[1]
	nuggetFractional := ""
	nuggetSignal := m[2]
	nuggetNotifConfirm := m[3]
	group1Orig := m[1]

	if nuggetSignal == "." {
		fm := reNumFloat.FindStringSubmatch(t.lineSource)
		if len(fm) == 6 {
			group1 := fm[1]
			if group1 == "" {
				group1 = "0"
			}
			nuggetNumber = group1 + fm[3]
			nuggetInteger = group1
			nuggetFractional = "." + fm[3]
			nuggetSignal = fm[4]
			nuggetNotifConfirm = fm[5]
			group1Orig = fm[1]
		}
	}
	if nuggetNumber == "" {
		nuggetNumber = "0"
	}

	isInt := false
	if nuggetSignal == "%" {
		nuggetNumber = nuggetInteger
		isInt = true
		if parseBig(nuggetNumber).Cmp(big32768) >= 0 {
			return false, fmt.Errorf("line %d: integer overflow: %s", lineNum, nuggetNumber)
		}
	} else if nuggetSignal != "%" && nuggetSignal != "!" && nuggetSignal != "#" &&
		((strings.ToLower(nuggetSignal) != "e" && strings.ToLower(nuggetSignal) != "d") ||
			(nuggetNotifConfirm != "-" && nuggetNotifConfirm != "+")) {
		nuggetSignal = ""
		if nuggetFractional == "" {
			isInt = true
		}
	}

	numberBig := parseBig(nuggetNumber)
	var hexa string

	if (strings.EqualFold(nuggetSignal, "e") || strings.EqualFold(nuggetSignal, "d")) && (nuggetNotifConfirm == "-" || nuggetNotifConfirm == "+") {
		em := reNumExp.FindStringSubmatch(t.lineSource)
		if len(em) != 3 {
			return false, fmt.Errorf("line %d: invalid scientific notation", lineNum)
		}
		expAbs := parseBig(em[2])
		expVal := new(big.Int).Set(expAbs)
		if em[1] == "-" {
			expVal.Neg(expVal)
		}
		trimInt := strings.TrimLeft(nuggetInteger, "0")
		nuggetExpSize := len(trimInt) + int(expVal.Int64())
		nuggetManSize := nuggetExpSize - len(strings.TrimPrefix(nuggetFractional, ".")) - len(trimInt)
		if nuggetExpSize > 63 || nuggetExpSize < -64 {
			return false, fmt.Errorf("line %d: float overflow: %s", lineNum, nuggetNumber)
		}

		notationInteger, notationFractional := scientificNotation(numberBig, nuggetManSize)
		notationNumber := notationInteger + strings.TrimPrefix(notationFractional, ".")
		notationSize := strings.TrimLeft(nuggetNumber, "0")

		if strings.EqualFold(nuggetSignal, "e") && len(notationSize) < 7 {
			hexa, _ = parseSgnDbl("1d", 6, notationInteger, notationFractional, group1Orig, notationNumber)
			hexa += strings.Repeat("0", max(0, 10-len(hexa)))
		} else {
			hexa, _ = parseSgnDbl("1f", 14, notationInteger, notationFractional, group1Orig, notationNumber)
			hexa += strings.Repeat("0", max(0, 18-len(hexa)))
			hexa = hexa[:18]
		}
		if strings.TrimLeft(nuggetInteger, "0") == "" {
			nuggetInteger = group1Orig
		}
		nuggetSignal += em[1] + em[2]
	} else if ((numberBig.Cmp(big32768) >= 0 && numberBig.Cmp(big999999) <= 0) && nuggetSignal != "#") ||
		(nuggetSignal == "!" && numberBig.Cmp(bigMax63) <= 0) ||
		(!isInt && numberBig.Cmp(big999999) <= 0 && nuggetSignal != "#") {
		hexa, nuggetInteger = parseSgnDbl("1d", 6, nuggetInteger, nuggetFractional, group1Orig, nuggetNumber)
		hexa += strings.Repeat("0", max(0, 10-len(hexa)))
	} else if (numberBig.Cmp(big1000000) >= 0 && numberBig.Cmp(bigMax63) <= 0) ||
		(nuggetSignal == "#" && numberBig.Cmp(bigMax63) <= 0) ||
		(!isInt && numberBig.Cmp(bigMax63) <= 0) {
		hexa, nuggetInteger = parseSgnDbl("1f", 14, nuggetInteger, nuggetFractional, group1Orig, nuggetNumber)
		hexa += strings.Repeat("0", max(0, 18-len(hexa)))
		hexa = hexa[:18]
	} else if numberBig.Cmp(bigNine) <= 0 {
		hexa = fmt.Sprintf("%02x", numberBig.Int64()+17)
	} else if numberBig.Cmp(big255) <= 0 {
		hexa = "0f" + fmt.Sprintf("%02x", numberBig.Int64())
	} else if numberBig.Cmp(big32767) <= 0 {
		hexa = "1c" + littleEndianWordHex(int(numberBig.Int64()))
	} else {
		return false, fmt.Errorf("line %d: number too high: %s", lineNum, strings.TrimLeft(nuggetNumber, "0"))
	}

	consumed := len(nuggetInteger) + len(nuggetFractional) + len(nuggetSignal)
	if consumed <= 0 {
		return false, nil
	}
	t.compiled = hexa
	t.source = consumed
	t.updateLines()
	return true, nil
}

func scientificNotation(value *big.Int, shift int) (string, string) {
	digits := value.Text(10)
	if shift >= 0 {
		return digits + strings.Repeat("0", shift), ""
	}
	fracSize := -shift
	if len(digits) <= fracSize {
		return "0", "." + strings.Repeat("0", fracSize-len(digits)) + digits
	}
	split := len(digits) - fracSize
	return digits[:split], "." + digits[split:]
}

func parseSgnDbl(header string, precision int, nuggetInteger, nuggetFractional, group1Orig, nuggetNumber string) (string, string) {
	stripped := strings.TrimLeft(nuggetInteger, "0")
	hexPrecision := ""
	if stripped == "" {
		if nuggetFractional == "" || allZeros(strings.TrimPrefix(nuggetFractional, ".")+"0") {
			stripped = "0"
			hexPrecision = "00"
		} else {
			nuggetInteger = group1Orig
			fracDigits := strings.TrimPrefix(nuggetFractional, ".")
			if len(fracDigits) > 0 && fracDigits[0] == '0' {
				zeros := strings.TrimRight(fracDigits, "0")
				leading := len(zeros) - len(strings.TrimLeft(zeros, "0"))
				hexPrecision = fmt.Sprintf("%02x", 64-leading)
			} else {
				hexPrecision = "40"
			}
		}
	} else {
		hexPrecision = fmt.Sprintf("%02x", len(stripped)+64)
	}

	hexa := header + hexPrecision
	cropped := parseBig(nuggetNumber).Text(10)
	roundDigit := 0
	if len(cropped) > precision {
		roundDigit = int(cropped[precision] - '0')
	}
	prefix := cropped
	if len(prefix) > precision {
		prefix = prefix[:precision]
	}
	if roundDigit >= 5 {
		prefix = new(big.Int).Add(parseBig(prefix), bigOne).Text(10)
	}
	if prefix == "" {
		prefix = "0"
	}
	hexa += prefix
	return hexa, nuggetInteger
}

func allZeros(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

func (t *tokenizer) tryBaseNumber(lineNum int) error {
	if len(t.lineSource) < 2 {
		t.compiled = "26"
		t.source = 1
		t.updateLines()
		return nil
	}
	prefix := strings.ToUpper(t.lineSource[:2])
	remain := ""
	if len(t.lineSource) > 2 {
		remain = t.lineSource[2:]
	}

	var digits string
	var token string
	var base int
	switch prefix {
	case "&H":
		token = "0c"
		base = 16
		digits = reHexDigits.FindString(strings.ToLower(remain))
	case "&O":
		token = "0b"
		base = 8
		digits = reOctDigits.FindString(remain)
	case "&B":
		token = "2642"
		digits = reBinDigits.FindString(remain)
		hexChars := strings.Builder{}
		hexChars.WriteString(token)
		for i := 0; i < len(digits); i++ {
			hexChars.WriteString(fmt.Sprintf("%02x", digits[i]))
		}
		t.compiled = hexChars.String()
		t.source = len(prefix) + len(digits)
		t.updateLines()
		return nil
	default:
		t.compiled = "26"
		t.source = 1
		t.updateLines()
		return nil
	}

	if digits == "" {
		t.compiled = token + "0000"
		t.source = len(prefix)
		t.updateLines()
		return nil
	}
	value, err := strconv.ParseInt(digits, base, 32)
	if err != nil {
		return fmt.Errorf("line %d: invalid base number %q", lineNum, digits)
	}
	if value > 65535 {
		return fmt.Errorf("line %d: number overflow: %s", lineNum, digits)
	}
	t.compiled = token + littleEndianWordHex(int(value))
	t.source = len(prefix) + len(digits)
	t.updateLines()
	return nil
}

func littleEndianWordHex(v int) string {
	hexa := fmt.Sprintf("%04x", v&0xffff)
	return hexa[2:] + hexa[:2]
}

func toUpperASCII(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
