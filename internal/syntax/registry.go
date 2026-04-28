package syntax

import "ws7/internal/syntax/msxbasic"

const (
	DialectMSXBasicOfficial = "msx-basic"
	DialectMSXBas2ROM       = "msxbas2rom"
	DialectBasicDignified   = "basic-dignified"
)

// DialectOption describes an available BASIC dialect in the UI.
type DialectOption struct {
	ID      string
	Label   string
	Enabled bool
}

var dialectOptions = []DialectOption{
	{ID: DialectMSXBasicOfficial, Label: "MSX-BASIC Official", Enabled: true},
	{ID: DialectMSXBas2ROM, Label: "MSXBAS2ROM", Enabled: false},
	{ID: DialectBasicDignified, Label: "BASIC Dignified", Enabled: false},
}

var highlighters = map[string]Highlighter{
	DialectMSXBasicOfficial: msxbasic.NewHighlighter(),
}

func DialectOptions() []DialectOption {
	out := make([]DialectOption, len(dialectOptions))
	copy(out, dialectOptions)
	return out
}

func DefaultDialect() DialectOption {
	for _, opt := range dialectOptions {
		if opt.Enabled {
			return opt
		}
	}
	return DialectOption{ID: DialectMSXBasicOfficial, Label: "MSX-BASIC Official", Enabled: true}
}

func HighlighterFor(dialectID string) Highlighter {
	h, ok := highlighters[dialectID]
	if ok {
		return h
	}
	return highlighters[DefaultDialect().ID]
}
