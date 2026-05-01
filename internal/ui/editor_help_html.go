package ui

import (
	"bytes"
	stdhtml "html"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	xhtml "golang.org/x/net/html"
)

// htmlToRichTextSegments converts raw HTML to Fyne RichText segments
// for in-app rendering without any external browser.
func htmlToRichTextSegments(rawHTML []byte, baseURL string) []widget.RichTextSegment {
	doc, err := xhtml.Parse(bytes.NewReader(rawHTML))
	if err != nil {
		return []widget.RichTextSegment{
			&widget.TextSegment{Text: "Error parsing page: " + err.Error()},
		}
	}

	root := helpFindRoot(doc)
	if root == nil {
		root = doc
	}

	var out []widget.RichTextSegment
	helpWalkBlock(&out, root, baseURL)
	if len(out) == 0 {
		out = append(out, &widget.TextSegment{Text: "No content found."})
	}
	return out
}

// helpFindRoot locates the best content container: article → main → body.
func helpFindRoot(doc *xhtml.Node) *xhtml.Node {
	for _, tag := range []string{"article", "main", "body"} {
		if n := helpFindElem(doc, tag); n != nil {
			return n
		}
	}
	return nil
}

func helpFindElem(root *xhtml.Node, tag string) *xhtml.Node {
	if root == nil {
		return nil
	}
	if root.Type == xhtml.ElementNode && strings.EqualFold(root.Data, tag) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if n := helpFindElem(c, tag); n != nil {
			return n
		}
	}
	return nil
}

func helpAttrVal(n *xhtml.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func helpHasClass(n *xhtml.Node, cls string) bool {
	for _, token := range strings.Fields(helpAttrVal(n, "class")) {
		if strings.EqualFold(token, cls) {
			return true
		}
	}
	return false
}

// helpWalkBlock processes block-level HTML nodes into RichText segments.
func helpWalkBlock(out *[]widget.RichTextSegment, n *xhtml.Node, baseURL string) {
	if n == nil {
		return
	}

	switch n.Type {
	case xhtml.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			helpWalkBlock(out, c, baseURL)
		}

	case xhtml.TextNode:
		text := helpInlineText(n.Data)
		if text != "" {
			*out = append(*out, &widget.TextSegment{
				Style: widget.RichTextStyleParagraph,
				Text:  text,
			})
		}

	case xhtml.ElementNode:
		tag := strings.ToLower(n.Data)

		// Skip non-content or navigation elements.
		switch tag {
		case "script", "style", "noscript", "nav", "footer", "head",
			"button", "input", "form", "select", "textarea":
			return
		}
		// Skip table-of-contents blocks.
		if helpHasClass(n, "toc") {
			return
		}

		switch tag {
		// ── Structural containers ──────────────────────────────────────
		case "html", "body", "div", "section", "article", "main",
			"header", "aside", "details", "summary", "figure", "figcaption":
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				helpWalkBlock(out, c, baseURL)
			}

		// ── Headings ──────────────────────────────────────────────────
		case "h1":
			text := helpCollectPlain(n)
			if text != "" {
				*out = append(*out,
					&widget.TextSegment{Style: widget.RichTextStyleHeading, Text: text},
					&widget.SeparatorSegment{},
				)
			}

		case "h2":
			text := helpCollectPlain(n)
			if text != "" {
				*out = append(*out, &widget.TextSegment{Style: widget.RichTextStyleSubHeading, Text: text})
			}

		case "h3", "h4", "h5", "h6":
			text := helpCollectPlain(n)
			if text != "" {
				*out = append(*out, &widget.TextSegment{
					Style: widget.RichTextStyle{
						TextStyle: fyne.TextStyle{Bold: true},
						Inline:    false,
					},
					Text: text,
				})
			}

		// ── Paragraph ─────────────────────────────────────────────────
		case "p":
			inline := helpBuildInline(n, baseURL)
			if len(inline) > 0 {
				*out = append(*out, &widget.ParagraphSegment{Texts: inline})
			}

		// ── Unordered / ordered list ──────────────────────────────────
		case "ul":
			items := helpCollectListItems(n, baseURL)
			if len(items) > 0 {
				*out = append(*out, &widget.ListSegment{Items: items})
			}

		case "ol":
			items := helpCollectListItems(n, baseURL)
			if len(items) > 0 {
				*out = append(*out, &widget.ListSegment{Items: items, Ordered: true})
			}

		// ── Definition list ───────────────────────────────────────────
		case "dl":
			helpWalkDL(out, n, baseURL)

		// ── Pre-formatted / code block ────────────────────────────────
		case "pre":
			text := helpCollectAllText(n)
			if text != "" {
				*out = append(*out, &widget.TextSegment{
					Style: widget.RichTextStyleCodeBlock,
					Text:  strings.TrimRight(text, "\n"),
				})
			}

		case "code":
			// Top-level <code> not inside <pre>.
			text := helpCollectPlain(n)
			if text != "" {
				*out = append(*out, &widget.TextSegment{
					Style: widget.RichTextStyleCodeBlock,
					Text:  text,
				})
			}

		// ── Table ─────────────────────────────────────────────────────
		case "table":
			helpWalkTable(out, n)

		// ── Horizontal rule ───────────────────────────────────────────
		case "hr":
			*out = append(*out, &widget.SeparatorSegment{})

		// ── Block-level line break ────────────────────────────────────
		case "br":
			*out = append(*out, &widget.TextSegment{
				Style: widget.RichTextStyleParagraph,
				Text:  "",
			})

		// ── Blockquote ────────────────────────────────────────────────
		case "blockquote":
			inline := helpBuildInline(n, baseURL)
			if len(inline) > 0 {
				*out = append(*out, &widget.ParagraphSegment{Texts: inline})
			}

		// ── Anything else: try to walk children ───────────────────────
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				helpWalkBlock(out, c, baseURL)
			}
		}
	}
}

// helpCollectListItems returns one TextSegment per <li>.
func helpCollectListItems(ul *xhtml.Node, baseURL string) []widget.RichTextSegment {
	var items []widget.RichTextSegment
	for c := ul.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != xhtml.ElementNode || !strings.EqualFold(c.Data, "li") {
			continue
		}
		// Use inline content if possible, fall back to plain text.
		inline := helpBuildInline(c, baseURL)
		if len(inline) > 0 {
			// ListSegment items must be a single segment; use ParagraphSegment.
			items = append(items, &widget.ParagraphSegment{Texts: inline})
		} else {
			text := helpCollectPlain(c)
			if text != "" {
				items = append(items, &widget.TextSegment{Text: text})
			}
		}
	}
	return items
}

// helpWalkDL renders <dl>/<dt>/<dd>.
func helpWalkDL(out *[]widget.RichTextSegment, dl *xhtml.Node, baseURL string) {
	for c := dl.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != xhtml.ElementNode {
			continue
		}
		switch strings.ToLower(c.Data) {
		case "dt":
			text := helpCollectPlain(c)
			if text != "" {
				*out = append(*out, &widget.TextSegment{
					Style: widget.RichTextStyleStrong,
					Text:  text,
				})
			}
		case "dd":
			inline := helpBuildInline(c, baseURL)
			if len(inline) > 0 {
				// Indent first text segment.
				if ts, ok := inline[0].(*widget.TextSegment); ok {
					ts.Text = "  " + ts.Text
				}
				*out = append(*out, &widget.ParagraphSegment{Texts: inline})
			}
		}
	}
}

// helpWalkTable renders a <table> as a monospaced code block.
func helpWalkTable(out *[]widget.RichTextSegment, table *xhtml.Node) {
	var rows [][]string
	helpCollectTableRows(table, &rows)
	if len(rows) == 0 {
		return
	}

	// Find column widths.
	widths := make([]int, 0)
	for _, row := range rows {
		for i, cell := range row {
			for len(widths) <= i {
				widths = append(widths, 0)
			}
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder
	for i, row := range rows {
		for j, cell := range row {
			w := 0
			if j < len(widths) {
				w = widths[j]
			}
			pad := w - len(cell)
			if pad < 0 {
				pad = 0
			}
			sb.WriteString(cell + strings.Repeat(" ", pad))
			if j < len(row)-1 {
				sb.WriteString("  |  ")
			}
		}
		sb.WriteString("\n")
		if i == 0 {
			total := 0
			for _, w := range widths {
				total += w + 5
			}
			sb.WriteString(strings.Repeat("-", total) + "\n")
		}
	}
	*out = append(*out, &widget.TextSegment{
		Style: widget.RichTextStyleCodeBlock,
		Text:  strings.TrimRight(sb.String(), "\n"),
	})
}

func helpCollectTableRows(n *xhtml.Node, rows *[][]string) {
	if n == nil {
		return
	}
	if n.Type == xhtml.ElementNode && strings.EqualFold(n.Data, "tr") {
		var cells []string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type != xhtml.ElementNode {
				continue
			}
			tag := strings.ToLower(c.Data)
			if tag != "td" && tag != "th" {
				continue
			}
			cells = append(cells, helpCollectPlain(c))
		}
		if len(cells) > 0 {
			*rows = append(*rows, cells)
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		helpCollectTableRows(c, rows)
	}
}

// helpBuildInline builds a slice of inline RichTextSegments from node children.
func helpBuildInline(n *xhtml.Node, baseURL string) []widget.RichTextSegment {
	var out []widget.RichTextSegment
	helpWalkInline(&out, n, baseURL, widget.RichTextStyleInline)
	return out
}

func helpWalkInline(out *[]widget.RichTextSegment, n *xhtml.Node, baseURL string, style widget.RichTextStyle) {
	if n == nil {
		return
	}
	switch n.Type {
	case xhtml.TextNode:
		text := helpInlineText(n.Data)
		if text != "" {
			*out = append(*out, &widget.TextSegment{Style: style, Text: text})
		}

	case xhtml.ElementNode:
		tag := strings.ToLower(n.Data)
		switch tag {
		case "script", "style", "noscript":
			return

		case "br":
			*out = append(*out, &widget.TextSegment{Style: style, Text: " "})
			return

		case "img":
			alt := strings.TrimSpace(helpAttrVal(n, "alt"))
			if alt != "" {
				*out = append(*out, &widget.TextSegment{Style: style, Text: "[" + alt + "]"})
			}
			return

		case "a":
			href := strings.TrimSpace(helpAttrVal(n, "href"))
			text := helpCollectPlain(n)
			if text == "" {
				text = href
			}
			// Only create hyperlinks for non-anchor absolute/relative URLs.
			if href != "" && !strings.HasPrefix(href, "#") {
				full := helpResolveURL(baseURL, href)
				if u, err := url.Parse(full); err == nil && u.Host != "" {
					*out = append(*out, &widget.HyperlinkSegment{
						Alignment: fyne.TextAlignLeading,
						Text:      text,
						URL:       u,
					})
					return
				}
			}
			if text != "" {
				*out = append(*out, &widget.TextSegment{Style: style, Text: text})
			}
			return

		case "strong", "b":
			bold := style
			bold.TextStyle.Bold = true
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				helpWalkInline(out, c, baseURL, bold)
			}
			return

		case "em", "i":
			italic := style
			italic.TextStyle.Italic = true
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				helpWalkInline(out, c, baseURL, italic)
			}
			return

		case "code", "tt", "kbd", "samp", "var":
			*out = append(*out, &widget.TextSegment{
				Style: widget.RichTextStyleCodeInline,
				Text:  helpCollectPlain(n),
			})
			return
		}
		// Default: recurse with the same style.
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			helpWalkInline(out, c, baseURL, style)
		}
	}
}

// helpCollectPlain extracts all visible text as normalized plain text.
func helpCollectPlain(n *xhtml.Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.TextNode {
			b.WriteString(stdhtml.UnescapeString(node.Data))
			return
		}
		if node.Type == xhtml.ElementNode {
			switch strings.ToLower(node.Data) {
			case "script", "style", "noscript":
				return
			case "br":
				b.WriteString(" ")
				return
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

// helpCollectAllText extracts text verbatim (for <pre> blocks).
func helpCollectAllText(n *xhtml.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == xhtml.TextNode {
		return stdhtml.UnescapeString(n.Data)
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(helpCollectAllText(c))
	}
	return b.String()
}

// helpInlineText normalizes whitespace, preserving one leading/trailing space
// for word separation between adjacent inline elements.
func helpInlineText(raw string) string {
	unescaped := stdhtml.UnescapeString(raw)
	if strings.TrimSpace(unescaped) == "" {
		if strings.ContainsAny(unescaped, " \t\n\r") {
			return " "
		}
		return ""
	}
	fields := strings.Fields(unescaped)
	result := strings.Join(fields, " ")
	if unescaped[0] == ' ' || unescaped[0] == '\t' || unescaped[0] == '\n' || unescaped[0] == '\r' {
		result = " " + result
	}
	last := unescaped[len(unescaped)-1]
	if last == ' ' || last == '\t' || last == '\n' || last == '\r' {
		result += " "
	}
	return result
}

// helpResolveURL resolves href relative to baseURL.
func helpResolveURL(baseRaw, href string) string {
	base, bErr := url.Parse(strings.TrimSpace(baseRaw))
	ref, rErr := url.Parse(strings.TrimSpace(href))
	if bErr != nil || rErr != nil {
		return strings.TrimSpace(href)
	}
	return base.ResolveReference(ref).String()
}
