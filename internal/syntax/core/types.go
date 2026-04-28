package core

// TokenKind classifies lexical tokens for syntax highlighting.
type TokenKind string

const (
	TokenPlain    TokenKind = "plain"
	TokenKeyword  TokenKind = "keyword"
	TokenFunction TokenKind = "function"
	TokenComment  TokenKind = "comment"
	TokenString   TokenKind = "string"
	TokenNumber   TokenKind = "number"
	TokenOperator TokenKind = "operator"
	TokenIdent    TokenKind = "identifier"
)

// Token is a classified text chunk produced by a highlighter.
type Token struct {
	Kind  TokenKind
	Value string
}
