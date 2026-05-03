package input

import (
	"fmt"
	"strings"
)

var allowedShortcutTokens = map[string]struct{}{
	"?":     {},
	"[":     {},
	"]":     {},
	"'":     {},
	"\"":    {},
	".":     {},
	"=":     {},
	"<":     {},
	"ENTER": {},
	"DEL":   {},
}

// NormalizeShortcut parses and canonicalizes shortcut text.
// Canonical format is: Ctrl+X,Y,Z
func NormalizeShortcut(raw string) (string, error) {
	tokens, err := parseShortcutTokens(raw)
	if err != nil {
		return "", err
	}
	if len(tokens) == 0 {
		return "", nil
	}
	if len(tokens) == 1 {
		return "Ctrl+" + tokens[0], nil
	}
	return "Ctrl+" + tokens[0] + "," + strings.Join(tokens[1:], ","), nil
}

// ShortcutToResolverChord converts shortcut text into resolver key tokens.
func ShortcutToResolverChord(raw string) ([]string, error) {
	tokens, err := parseShortcutTokens(raw)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		switch token {
		case "ENTER":
			out = append(out, "\r")
		case "DEL":
			out = append(out, "\b")
		default:
			out = append(out, token)
		}
	}
	return out, nil
}

func parseShortcutTokens(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	tokens := make([]string, 0, len(parts))
	for i, part := range parts {
		piece := strings.TrimSpace(part)
		if piece == "" {
			return nil, fmt.Errorf("invalid shortcut segment")
		}
		upper := strings.ToUpper(piece)
		if strings.HasPrefix(upper, "CTRL+") {
			piece = strings.TrimSpace(piece[5:])
		} else if i == 0 {
			return nil, fmt.Errorf("shortcut must start with Ctrl+")
		}
		tok, err := normalizeShortcutToken(piece)
		if err != nil {
			return nil, err
		}
		if i == 0 && (len(tok) != 1 || tok[0] < 'A' || tok[0] > 'Z') {
			return nil, fmt.Errorf("first key must be A-Z")
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func normalizeShortcutToken(token string) (string, error) {
	t := strings.ToUpper(strings.TrimSpace(token))
	switch t {
	case "RETURN":
		return "ENTER", nil
	case "DELETE", "BACKSPACE":
		return "DEL", nil
	}
	if len(t) == 1 {
		c := t[0]
		if c >= 'A' && c <= 'Z' {
			return t, nil
		}
		if c >= '0' && c <= '9' {
			return t, nil
		}
		if _, ok := allowedShortcutTokens[t]; ok {
			return t, nil
		}
	}
	if _, ok := allowedShortcutTokens[t]; ok {
		return t, nil
	}
	return "", fmt.Errorf("unsupported key token %q", token)
}
