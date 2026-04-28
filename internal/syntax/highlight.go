package syntax

import "strings"

// HighlightDocument tokenizes each line preserving line order.
func HighlightDocument(dialectID, text string) [][]Token {
	h := HighlighterFor(dialectID)
	lines := strings.Split(text, "\n")
	out := make([][]Token, 0, len(lines))
	for _, line := range lines {
		out = append(out, h.HighlightLine(line))
	}
	return out
}
