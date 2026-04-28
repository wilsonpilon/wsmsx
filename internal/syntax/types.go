package syntax

import "ws7/internal/syntax/core"

type TokenKind = core.TokenKind

const (
	TokenPlain    TokenKind = core.TokenPlain
	TokenKeyword  TokenKind = core.TokenKeyword
	TokenFunction TokenKind = core.TokenFunction
	TokenComment  TokenKind = core.TokenComment
	TokenString   TokenKind = core.TokenString
	TokenNumber   TokenKind = core.TokenNumber
	TokenOperator TokenKind = core.TokenOperator
	TokenIdent    TokenKind = core.TokenIdent
)

type Token = core.Token

// Highlighter classifies source text by line.
type Highlighter interface {
	ID() string
	Name() string
	HighlightLine(line string) []core.Token
}
