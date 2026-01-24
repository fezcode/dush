package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT"  // add, foobar, x, y, ...
	INT    = "INT"    // 1343456
	FLOAT  = "FLOAT"  // 12.34
	STRING = "STRING" // "foobar"

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="
	AND    = "&&"
	OR     = "||"

	// Shell Operators
	PIPE     = "|"
	REDIRECT = ">" // Same as GT, context matters, but lexer just sees symbols
	APPEND   = ">>"
	INPUT    = "<" // Same as LT

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	PROC   = "PROC"
	LET    = "LET"
	CONST  = "CONST"
	TRUE   = "TRUE"
	FALSE  = "FALSE"
	IF     = "IF"
	ELSE   = "ELSE"
	RETURN = "RETURN"
	LOOP   = "LOOP"
	WITH   = "WITH"
	SAVE   = "SAVE"
)

var keywords = map[string]TokenType{
	"proc":   PROC,
	"let":    LET,
	"const":  CONST,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"loop":   LOOP,
	"with":   WITH,
	"save":   SAVE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
