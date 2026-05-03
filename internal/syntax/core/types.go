package core

// TokenKind classifies lexical tokens for syntax highlighting.
type TokenKind string

const (
	TokenPlain    TokenKind = "plain"
	TokenKeyword  TokenKind = "keyword"  // instructions  (PRINT, CLS, COLOR…)
	TokenJump     TokenKind = "jump"     // jump commands  (GOTO, GOSUB, THEN…)
	TokenFunction TokenKind = "function" // functions      (LEFT$, INT, SIN…)
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
