package calc

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"unicode"
)

// Result contains the evaluated value and formatted outputs.
type Result struct {
	Value   float64
	Decimal string
	Hex     string
	Binary  string
}

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokNumber
	tokIdent
	tokPlus
	tokMinus
	tokMul
	tokDiv
	tokPow
	tokLParen
	tokRParen
	tokComma
	tokShiftLeft
	tokShiftRight
)

type token struct {
	kind tokenKind
	text string
	num  float64
}

type lexer struct {
	s       string
	pos     int
	hasLast bool
	last    float64
}

func (l *lexer) nextToken() (token, error) {
	l.skipSpaces()
	if l.pos >= len(l.s) {
		return token{kind: tokEOF}, nil
	}

	ch := l.s[l.pos]
	switch ch {
	case '+':
		l.pos++
		return token{kind: tokPlus, text: "+"}, nil
	case '-':
		l.pos++
		return token{kind: tokMinus, text: "-"}, nil
	case '*':
		l.pos++
		return token{kind: tokMul, text: "*"}, nil
	case '/':
		l.pos++
		return token{kind: tokDiv, text: "/"}, nil
	case '^':
		l.pos++
		return token{kind: tokPow, text: "^"}, nil
	case '(':
		l.pos++
		return token{kind: tokLParen, text: "("}, nil
	case ')':
		l.pos++
		return token{kind: tokRParen, text: ")"}, nil
	case ',':
		l.pos++
		return token{kind: tokComma, text: ","}, nil
	case '<':
		if l.peek("<<") {
			l.pos += 2
			return token{kind: tokShiftLeft, text: "<<"}, nil
		}
	case '>':
		if l.peek(">>") {
			l.pos += 2
			return token{kind: tokShiftRight, text: ">>"}, nil
		}
	case '&':
		if len(l.s)-l.pos >= 2 {
			prefix := unicode.ToUpper(rune(l.s[l.pos+1]))
			if prefix == 'H' || prefix == 'B' {
				return l.readPrefixedNumber(prefix)
			}
		}
	}

	if unicode.IsLetter(rune(ch)) || ch == '_' {
		start := l.pos
		for l.pos < len(l.s) {
			r := rune(l.s[l.pos])
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				break
			}
			l.pos++
		}
		text := l.s[start:l.pos]
		return token{kind: tokIdent, text: strings.ToUpper(text)}, nil
	}

	if unicode.IsDigit(rune(ch)) || ch == '.' {
		if ch == '.' && (l.pos+1 >= len(l.s) || !unicode.IsDigit(rune(l.s[l.pos+1]))) {
			if !l.hasLast {
				return token{}, fmt.Errorf("no previous result available for '.'")
			}
			l.pos++
			return token{kind: tokNumber, text: ".", num: l.last}, nil
		}
		return l.readDecimalNumber()
	}

	return token{}, fmt.Errorf("invalid token at position %d", l.pos+1)
}

func (l *lexer) readDecimalNumber() (token, error) {
	start := l.pos
	hasDot := false
	for l.pos < len(l.s) {
		ch := l.s[l.pos]
		if ch == '.' {
			if hasDot {
				break
			}
			hasDot = true
			l.pos++
			continue
		}
		if !unicode.IsDigit(rune(ch)) {
			break
		}
		l.pos++
	}
	text := l.s[start:l.pos]
	if text == "." {
		return token{}, fmt.Errorf("invalid number at position %d", start+1)
	}
	n, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number %q", text)
	}
	return token{kind: tokNumber, text: text, num: n}, nil
}

func (l *lexer) readPrefixedNumber(prefix rune) (token, error) {
	start := l.pos
	l.pos += 2 // &H / &B
	digitsStart := l.pos
	for l.pos < len(l.s) {
		ch := rune(l.s[l.pos])
		if prefix == 'H' {
			if !isHexDigit(ch) {
				break
			}
		} else {
			if ch != '0' && ch != '1' {
				break
			}
		}
		l.pos++
	}
	if l.pos == digitsStart {
		return token{}, fmt.Errorf("missing digits after %s at position %d", l.s[start:start+2], start+1)
	}
	text := l.s[start:l.pos]
	base := 16
	if prefix == 'B' {
		base = 2
	}
	parsed, err := strconv.ParseInt(text[2:], base, 64)
	if err != nil {
		return token{}, fmt.Errorf("invalid number %q", text)
	}
	return token{kind: tokNumber, text: text, num: float64(parsed)}, nil
}

func (l *lexer) skipSpaces() {
	for l.pos < len(l.s) && unicode.IsSpace(rune(l.s[l.pos])) {
		l.pos++
	}
}

func (l *lexer) peek(s string) bool {
	return len(l.s)-l.pos >= len(s) && l.s[l.pos:l.pos+len(s)] == s
}

func isHexDigit(ch rune) bool {
	return unicode.IsDigit(ch) || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')
}

type parser struct {
	lx  *lexer
	cur token
}

func Evaluate(expr string) (Result, error) {
	return EvaluateWithLast(expr, 0, false)
}

// EvaluateWithLast evaluates an expression with optional bc-style last result support.
// When hasLast is true, a standalone '.' token resolves to last.
func EvaluateWithLast(expr string, last float64, hasLast bool) (Result, error) {
	p := &parser{lx: &lexer{s: strings.TrimSpace(expr), last: last, hasLast: hasLast}}
	if err := p.next(); err != nil {
		return Result{}, err
	}
	if p.cur.kind == tokEOF {
		return Result{}, fmt.Errorf("empty expression")
	}
	v, err := p.parseExpr()
	if err != nil {
		return Result{}, err
	}
	if p.cur.kind != tokEOF {
		return Result{}, fmt.Errorf("unexpected token %q", p.cur.text)
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return Result{}, fmt.Errorf("invalid numeric result")
	}

	iv := int64(math.Trunc(v))
	return Result{
		Value:   v,
		Decimal: formatDecimal(v),
		Hex:     formatHex(iv),
		Binary:  formatBinary(iv),
	}, nil
}

func formatDecimal(v float64) string {
	if math.Abs(v-math.Trunc(v)) < 1e-9 {
		return strconv.FormatInt(int64(math.Trunc(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatHex(v int64) string {
	if v < 0 {
		return "-&H" + strings.ToUpper(strconv.FormatUint(uint64(-v), 16))
	}
	return "&H" + strings.ToUpper(strconv.FormatUint(uint64(v), 16))
}

func formatBinary(v int64) string {
	if v < 0 {
		return "-&B" + strconv.FormatUint(uint64(-v), 2)
	}
	return "&B" + strconv.FormatUint(uint64(v), 2)
}

func (p *parser) next() error {
	tok, err := p.lx.nextToken()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *parser) parseExpr() (float64, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (float64, error) {
	left, err := p.parseXor()
	if err != nil {
		return 0, err
	}
	for p.isIdent("OR") {
		if err := p.next(); err != nil {
			return 0, err
		}
		right, err := p.parseXor()
		if err != nil {
			return 0, err
		}
		left = float64(asInt(left) | asInt(right))
	}
	return left, nil
}

func (p *parser) parseXor() (float64, error) {
	left, err := p.parseAnd()
	if err != nil {
		return 0, err
	}
	for p.isIdent("XOR") {
		if err := p.next(); err != nil {
			return 0, err
		}
		right, err := p.parseAnd()
		if err != nil {
			return 0, err
		}
		left = float64(asInt(left) ^ asInt(right))
	}
	return left, nil
}

func (p *parser) parseAnd() (float64, error) {
	left, err := p.parseShift()
	if err != nil {
		return 0, err
	}
	for p.isIdent("AND") {
		if err := p.next(); err != nil {
			return 0, err
		}
		right, err := p.parseShift()
		if err != nil {
			return 0, err
		}
		left = float64(asInt(left) & asInt(right))
	}
	return left, nil
}

func (p *parser) parseShift() (float64, error) {
	left, err := p.parseAdd()
	if err != nil {
		return 0, err
	}
	for {
		switch {
		case p.cur.kind == tokShiftLeft || p.isIdent("SHL"):
			if err := p.next(); err != nil {
				return 0, err
			}
			right, err := p.parseAdd()
			if err != nil {
				return 0, err
			}
			shift := uint(asInt(right) & 63)
			left = float64(asInt(left) << shift)
		case p.cur.kind == tokShiftRight || p.isIdent("SHR"):
			if err := p.next(); err != nil {
				return 0, err
			}
			right, err := p.parseAdd()
			if err != nil {
				return 0, err
			}
			shift := uint(asInt(right) & 63)
			left = float64(asInt(left) >> shift)
		default:
			return left, nil
		}
	}
}

func (p *parser) parseAdd() (float64, error) {
	left, err := p.parseMul()
	if err != nil {
		return 0, err
	}
	for p.cur.kind == tokPlus || p.cur.kind == tokMinus {
		op := p.cur.kind
		if err := p.next(); err != nil {
			return 0, err
		}
		right, err := p.parseMul()
		if err != nil {
			return 0, err
		}
		if op == tokPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *parser) parseMul() (float64, error) {
	left, err := p.parsePow()
	if err != nil {
		return 0, err
	}
	for p.cur.kind == tokMul || p.cur.kind == tokDiv {
		op := p.cur.kind
		if err := p.next(); err != nil {
			return 0, err
		}
		right, err := p.parsePow()
		if err != nil {
			return 0, err
		}
		if op == tokMul {
			left *= right
			continue
		}
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		left /= right
	}
	return left, nil
}

func (p *parser) parsePow() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	if p.cur.kind != tokPow {
		return left, nil
	}
	if err := p.next(); err != nil {
		return 0, err
	}
	right, err := p.parsePow()
	if err != nil {
		return 0, err
	}
	return math.Pow(left, right), nil
}

func (p *parser) parseUnary() (float64, error) {
	switch {
	case p.cur.kind == tokPlus:
		if err := p.next(); err != nil {
			return 0, err
		}
		return p.parseUnary()
	case p.cur.kind == tokMinus:
		if err := p.next(); err != nil {
			return 0, err
		}
		v, err := p.parseUnary()
		return -v, err
	case p.isIdent("NOT"):
		if err := p.next(); err != nil {
			return 0, err
		}
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return float64(^asInt(v)), nil
	case p.isIdent("INT"):
		if err := p.next(); err != nil {
			return 0, err
		}
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return math.Trunc(v), nil
	case p.isIdent("SQR"):
		if err := p.next(); err != nil {
			return 0, err
		}
		v, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		if v < 0 {
			return 0, fmt.Errorf("sqr of negative value")
		}
		return math.Sqrt(v), nil
	default:
		return p.parsePrimary()
	}
}

func (p *parser) parsePrimary() (float64, error) {
	switch p.cur.kind {
	case tokNumber:
		v := p.cur.num
		if err := p.next(); err != nil {
			return 0, err
		}
		return v, nil
	case tokLParen:
		if err := p.next(); err != nil {
			return 0, err
		}
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.cur.kind != tokRParen {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		if err := p.next(); err != nil {
			return 0, err
		}
		return v, nil
	case tokIdent:
		name := p.cur.text
		if err := p.next(); err != nil {
			return 0, err
		}
		if p.cur.kind != tokLParen {
			return 0, fmt.Errorf("unknown identifier %q", name)
		}
		if err := p.next(); err != nil {
			return 0, err
		}
		args := make([]float64, 0, 2)
		if p.cur.kind != tokRParen {
			for {
				arg, err := p.parseExpr()
				if err != nil {
					return 0, err
				}
				args = append(args, arg)
				if p.cur.kind == tokRParen {
					break
				}
				if p.cur.kind != tokComma {
					return 0, fmt.Errorf("expected ',' in %s(...)", name)
				}
				if err := p.next(); err != nil {
					return 0, err
				}
			}
		}
		if p.cur.kind != tokRParen {
			return 0, fmt.Errorf("missing closing ')' in %s(...)", name)
		}
		if err := p.next(); err != nil {
			return 0, err
		}
		return applyFunc(name, args)
	default:
		return 0, fmt.Errorf("unexpected token %q", p.cur.text)
	}
}

func applyFunc(name string, args []float64) (float64, error) {
	require := func(n int) error {
		if len(args) != n {
			return fmt.Errorf("%s expects %d argument(s)", strings.ToLower(name), n)
		}
		return nil
	}

	switch name {
	case "SQR":
		if err := require(1); err != nil {
			return 0, err
		}
		if args[0] < 0 {
			return 0, fmt.Errorf("sqr of negative value")
		}
		return math.Sqrt(args[0]), nil
	case "INT":
		if err := require(1); err != nil {
			return 0, err
		}
		return math.Trunc(args[0]), nil
	case "HEX", "BIN", "DEC":
		if err := require(1); err != nil {
			return 0, err
		}
		return float64(asInt(args[0])), nil
	case "ROL":
		if err := require(2); err != nil {
			return 0, err
		}
		shift := int(asInt(args[1]) & 63)
		return float64(int64(bits.RotateLeft64(uint64(asInt(args[0])), shift))), nil
	case "ROR":
		if err := require(2); err != nil {
			return 0, err
		}
		shift := int(asInt(args[1]) & 63)
		return float64(int64(bits.RotateLeft64(uint64(asInt(args[0])), -shift))), nil
	case "SHL":
		if err := require(2); err != nil {
			return 0, err
		}
		return float64(asInt(args[0]) << uint(asInt(args[1])&63)), nil
	case "SHR":
		if err := require(2); err != nil {
			return 0, err
		}
		return float64(asInt(args[0]) >> uint(asInt(args[1])&63)), nil
	default:
		return 0, fmt.Errorf("unknown function %q", strings.ToLower(name))
	}
}

func asInt(v float64) int64 {
	return int64(math.Trunc(v))
}

func (p *parser) isIdent(name string) bool {
	return p.cur.kind == tokIdent && p.cur.text == name
}
