package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"ws7/internal/syntax"
)

// syntaxHighlightEntry combines a RichText display with an Entry for editing,
// syncing the content between them and providing syntax highlighting in real-time.
type syntaxHighlightEntry struct {
	widget.BaseWidget
	entry        *cursorEntry
	richText     *widget.RichText
	dialectID    string
	highlights   [][]syntax.Token
}

func newSyntaxHighlightEntry(dialectID string) *syntaxHighlightEntry {
	entry := newCursorEntry()
	entry.hideText = true
	richText := widget.NewRichText()
	richText.Wrapping = fyne.TextWrapOff

	e := &syntaxHighlightEntry{
		entry:     entry,
		richText:  richText,
		dialectID: dialectID,
	}

	// Sync text changes
	entry.OnChanged = func(_ string) {
		e.updateHighlights()
		e.Refresh()
	}

	e.ExtendBaseWidget(e)
	return e
}

// Text returns the current text content
func (e *syntaxHighlightEntry) Text() string {
	return e.entry.Text
}

// SetText updates the text content
func (e *syntaxHighlightEntry) SetText(text string) {
	e.entry.SetText(text)
	e.updateHighlights()
}

// SetDialect changes the syntax dialect
func (e *syntaxHighlightEntry) SetDialect(dialectID string) {
	if e.dialectID == dialectID {
		return
	}
	e.dialectID = dialectID
	e.updateHighlights()
	e.Refresh()
}

// updateHighlights tokenizes the current text and updates the RichText display
func (e *syntaxHighlightEntry) updateHighlights() {
	if e.dialectID == "" {
		e.highlights = nil
		e.richText.Segments = nil
		return
	}

	e.highlights = syntax.HighlightDocument(e.dialectID, e.entry.Text)
	e.richText.Segments = syntaxPreviewSegments(e.highlights)
}

// CreateRenderer creates a custom renderer that composites RichText and Entry
func (e *syntaxHighlightEntry) CreateRenderer() fyne.WidgetRenderer {
	return &syntaxHighlightRenderer{
		entry:    e.entry,
		richText: e.richText,
		objects: []fyne.CanvasObject{
			e.richText,
			e.entry,
		},
	}
}

// syntaxHighlightRenderer renders the syntax-highlighted text and editable entry
type syntaxHighlightRenderer struct {
	entry    *cursorEntry
	richText *widget.RichText
	objects  []fyne.CanvasObject
}

func (r *syntaxHighlightRenderer) Destroy() {
	// No cleanup needed
}

func (r *syntaxHighlightRenderer) Layout(size fyne.Size) {
	// Position both the RichText (behind) and Entry (in front) at the same location
	r.richText.Move(fyne.NewPos(0, 0))
	r.richText.Resize(size)

	r.entry.Move(fyne.NewPos(0, 0))
	r.entry.Resize(size)
}

func (r *syntaxHighlightRenderer) MinSize() fyne.Size {
	// Use the Entry's min size
	return r.entry.MinSize()
}

func (r *syntaxHighlightRenderer) Refresh() {
	// Refresh both components
	r.richText.Refresh()
	r.entry.Refresh()
}

func (r *syntaxHighlightRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

// Expose Entry properties
func (e *syntaxHighlightEntry) CursorRow() int {
	return e.entry.CursorRow
}

func (e *syntaxHighlightEntry) CursorColumn() int {
	return e.entry.CursorColumn
}

// Delegate input methods to the entry
func (e *syntaxHighlightEntry) TypedKey(key *fyne.KeyEvent) {
	e.entry.TypedKey(key)
}

func (e *syntaxHighlightEntry) TypedRune(r rune) {
	e.entry.TypedRune(r)
}

func (e *syntaxHighlightEntry) TypedShortcut(shortcut fyne.Shortcut) {
	e.entry.TypedShortcut(shortcut)
}

func (e *syntaxHighlightEntry) Tapped(ev *fyne.PointEvent) {
	e.entry.Tapped(ev)
}

func (e *syntaxHighlightEntry) TappedSecondary(ev *fyne.PointEvent) {
	e.entry.TappedSecondary(ev)
}








