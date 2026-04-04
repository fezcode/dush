package token

type TokenType string

type Token struct {
	Type            TokenType
	Literal         string
	PrecededBySpace bool
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT      = "IDENT"      // add, foobar, x, y, ...
	INT        = "INT"        // 1343456
	FLOAT      = "FLOAT"      // 12.34
	STRING     = "STRING"     // "foobar"
	RAW_STRING = "RAW_STRING" // 'foobar' (no interpolation)

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	MODULO   = "%"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="
	AND    = "&&"
	OR     = "||"

	// Shell Operators
	PIPE      = "|"
	REDIRECT  = ">" // Same as GT, context matters
	APPEND    = ">>"
	INPUT     = "<" // Same as LT
	AMPERSAND = "&"

	// Variable sigil
	AT = "@"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	DOT       = "."
	BACKSLASH = "\\"

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
	PUB    = "PUB"
	MATCH  = "MATCH"
	CASE   = "CASE"

	// Legacy (kept as constant, removed from keywords map)
	SAVE = "SAVE"
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
	"pub":    PUB,
	"match":  MATCH,
	"case":   CASE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
