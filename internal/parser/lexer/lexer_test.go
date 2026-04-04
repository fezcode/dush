package lexer

import (
	"dush/internal/parser/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let @five = 5
let @ten = 10

proc add(@x, @y) {
	@x + @y
}

let @result = add(@five, @ten)
!-/*5
5 < 10 > 5

if (@x == 10) {
	return true
} else {
	return false
}

10 == 10
10 != 9
"hello world"
'raw string'
loop (@x < 10) {}
loop (@i : items) {}
@x && @y || @z
echo "hello @name"
const @PI = 3.14
pub @KEY = "value"
@name.upper()
echo file.txt
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.AT, "@"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, "\n"},

		{token.LET, "let"},
		{token.AT, "@"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, "\n"},

		{token.SEMICOLON, "\n"},

		{token.PROC, "proc"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.AT, "@"},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.AT, "@"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, "\n"},

		{token.AT, "@"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.AT, "@"},
		{token.IDENT, "y"},
		{token.SEMICOLON, "\n"},

		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},

		{token.SEMICOLON, "\n"},

		{token.LET, "let"},
		{token.AT, "@"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.AT, "@"},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.AT, "@"},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, "\n"},

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, "\n"},

		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, "\n"},

		{token.SEMICOLON, "\n"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.AT, "@"},
		{token.IDENT, "x"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, "\n"},

		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, "\n"},

		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, "\n"},

		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, "\n"},

		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},

		{token.SEMICOLON, "\n"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, "\n"},

		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, "\n"},

		{token.STRING, "hello world"},
		{token.SEMICOLON, "\n"},

		{token.RAW_STRING, "raw string"},
		{token.SEMICOLON, "\n"},

		{token.LOOP, "loop"},
		{token.LPAREN, "("},
		{token.AT, "@"},
		{token.IDENT, "x"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},

		{token.LOOP, "loop"},
		{token.LPAREN, "("},
		{token.AT, "@"},
		{token.IDENT, "i"},
		{token.COLON, ":"},
		{token.IDENT, "items"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},

		{token.AT, "@"},
		{token.IDENT, "x"},
		{token.AND, "&&"},
		{token.AT, "@"},
		{token.IDENT, "y"},
		{token.OR, "||"},
		{token.AT, "@"},
		{token.IDENT, "z"},
		{token.SEMICOLON, "\n"},

		{token.IDENT, "echo"},
		{token.STRING, "hello @name"},
		{token.SEMICOLON, "\n"},

		{token.CONST, "const"},
		{token.AT, "@"},
		{token.IDENT, "PI"},
		{token.ASSIGN, "="},
		{token.FLOAT, "3.14"},
		{token.SEMICOLON, "\n"},

		{token.PUB, "pub"},
		{token.AT, "@"},
		{token.IDENT, "KEY"},
		{token.ASSIGN, "="},
		{token.STRING, "value"},
		{token.SEMICOLON, "\n"},

		{token.AT, "@"},
		{token.IDENT, "name"},
		{token.DOT, "."},
		{token.IDENT, "upper"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.SEMICOLON, "\n"},

		{token.IDENT, "echo"},
		{token.IDENT, "file"},
		{token.DOT, "."},
		{token.IDENT, "txt"},
		{token.SEMICOLON, "\n"},

		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestPrecededBySpace(t *testing.T) {
	input := `echo hello @name`
	l := New(input)

	tok := l.NextToken() // echo
	if tok.PrecededBySpace != false {
		t.Errorf("echo: expected PrecededBySpace=false, got=%v", tok.PrecededBySpace)
	}

	tok = l.NextToken() // hello
	if tok.PrecededBySpace != true {
		t.Errorf("hello: expected PrecededBySpace=true, got=%v", tok.PrecededBySpace)
	}

	tok = l.NextToken() // @
	if tok.PrecededBySpace != true {
		t.Errorf("@: expected PrecededBySpace=true, got=%v", tok.PrecededBySpace)
	}

	tok = l.NextToken() // name
	if tok.PrecededBySpace != false {
		t.Errorf("name: expected PrecededBySpace=false, got=%v", tok.PrecededBySpace)
	}
}

func TestSingleQuoteString(t *testing.T) {
	input := `'hello @world'`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.RAW_STRING {
		t.Errorf("expected RAW_STRING, got %s", tok.Type)
	}
	if tok.Literal != "hello @world" {
		t.Errorf("expected 'hello @world', got %q", tok.Literal)
	}
}

func TestAmpersandToken(t *testing.T) {
	input := `sleep 10 &`
	l := New(input)

	tok := l.NextToken() // sleep
	if tok.Type != token.IDENT || tok.Literal != "sleep" {
		t.Errorf("expected IDENT 'sleep', got %s %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken() // 10
	if tok.Type != token.INT || tok.Literal != "10" {
		t.Errorf("expected INT '10', got %s %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken() // &
	if tok.Type != token.AMPERSAND || tok.Literal != "&" {
		t.Errorf("expected AMPERSAND '&', got %s %q", tok.Type, tok.Literal)
	}
}

func TestAmpersandVsAnd(t *testing.T) {
	input := `a & b && c`
	l := New(input)

	l.NextToken() // a
	tok := l.NextToken()
	if tok.Type != token.AMPERSAND {
		t.Errorf("expected AMPERSAND for single &, got %s", tok.Type)
	}

	l.NextToken() // b
	tok = l.NextToken()
	if tok.Type != token.AND {
		t.Errorf("expected AND for &&, got %s", tok.Type)
	}
}

func TestMatchCaseTokens(t *testing.T) {
	input := `match case`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.MATCH {
		t.Errorf("expected MATCH, got %s", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != token.CASE {
		t.Errorf("expected CASE, got %s", tok.Type)
	}
}

func TestDotToken(t *testing.T) {
	input := `file.txt`
	l := New(input)

	tok := l.NextToken() // file
	if tok.Type != token.IDENT || tok.Literal != "file" {
		t.Errorf("expected IDENT 'file', got %s %q", tok.Type, tok.Literal)
	}

	tok = l.NextToken() // .
	if tok.Type != token.DOT || tok.Literal != "." {
		t.Errorf("expected DOT '.', got %s %q", tok.Type, tok.Literal)
	}
	if tok.PrecededBySpace {
		t.Error("DOT after 'file' should not be preceded by space")
	}

	tok = l.NextToken() // txt
	if tok.Type != token.IDENT || tok.Literal != "txt" {
		t.Errorf("expected IDENT 'txt', got %s %q", tok.Type, tok.Literal)
	}
}

func TestDotWithSpace(t *testing.T) {
	input := `cd .config`
	l := New(input)

	l.NextToken() // cd
	tok := l.NextToken() // .
	if tok.Type != token.DOT {
		t.Errorf("expected DOT, got %s", tok.Type)
	}
	if !tok.PrecededBySpace {
		t.Error("DOT after 'cd ' should be preceded by space")
	}

	tok = l.NextToken() // config
	if tok.Type != token.IDENT || tok.Literal != "config" {
		t.Errorf("expected IDENT 'config', got %s %q", tok.Type, tok.Literal)
	}
	if tok.PrecededBySpace {
		t.Error("'config' should not be preceded by space (attached to dot)")
	}
}

func TestFloatLiteral(t *testing.T) {
	input := `3.14`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.FLOAT {
		t.Errorf("expected FLOAT, got %s", tok.Type)
	}
	if tok.Literal != "3.14" {
		t.Errorf("expected '3.14', got %q", tok.Literal)
	}
}

func TestShellOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.TokenType
	}{
		{"|", []token.TokenType{token.PIPE}},
		{">>", []token.TokenType{token.APPEND}},
		{"&&", []token.TokenType{token.AND}},
		{"||", []token.TokenType{token.OR}},
		{"==", []token.TokenType{token.EQ}},
		{"!=", []token.TokenType{token.NOT_EQ}},
		{"&", []token.TokenType{token.AMPERSAND}},
	}

	for _, tt := range tests {
		l := New(tt.input)
		for i, expectedType := range tt.expected {
			tok := l.NextToken()
			if tok.Type != expectedType {
				t.Errorf("input %q, token %d: expected %s, got %s", tt.input, i, expectedType, tok.Type)
			}
		}
	}
}

func TestCommentSkipping(t *testing.T) {
	input := "// this is a comment\n42"
	l := New(input)

	// Comment produces a newline (SEMICOLON) first, then 42
	tok := l.NextToken()
	if tok.Type == token.SEMICOLON {
		tok = l.NextToken() // skip the newline
	}
	if tok.Type != token.INT || tok.Literal != "42" {
		t.Errorf("expected INT '42' after comment, got %s %q", tok.Type, tok.Literal)
	}
}

func TestAllDelimiters(t *testing.T) {
	input := `(){}[],;:`
	expected := []token.TokenType{
		token.LPAREN, token.RPAREN,
		token.LBRACE, token.RBRACE,
		token.LBRACKET, token.RBRACKET,
		token.COMMA, token.SEMICOLON, token.COLON,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("delimiter %d: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestAllArithmeticOperators(t *testing.T) {
	input := `+ - * / %`
	expected := []token.TokenType{
		token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.MODULO,
	}

	l := New(input)
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp {
			t.Errorf("operator %d: expected %s, got %s (%q)", i, exp, tok.Type, tok.Literal)
		}
	}
}

func TestIdentifiersWithUnderscores(t *testing.T) {
	input := `my_var _private __dunder`
	l := New(input)

	tests := []string{"my_var", "_private", "__dunder"}
	for _, expected := range tests {
		tok := l.NextToken()
		if tok.Type != token.IDENT || tok.Literal != expected {
			t.Errorf("expected IDENT %q, got %s %q", expected, tok.Type, tok.Literal)
		}
	}
}

func TestEmptyString(t *testing.T) {
	input := `""`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.STRING {
		t.Errorf("expected STRING, got %s", tok.Type)
	}
	if tok.Literal != "" {
		t.Errorf("expected empty string, got %q", tok.Literal)
	}
}

func TestEmptyRawString(t *testing.T) {
	input := `''`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.RAW_STRING {
		t.Errorf("expected RAW_STRING, got %s", tok.Type)
	}
	if tok.Literal != "" {
		t.Errorf("expected empty raw string, got %q", tok.Literal)
	}
}

func TestBackslashToken(t *testing.T) {
	input := `\`
	l := New(input)

	tok := l.NextToken()
	if tok.Type != token.BACKSLASH {
		t.Errorf("expected BACKSLASH, got %s", tok.Type)
	}
}

func TestAllKeywords(t *testing.T) {
	keywords := map[string]token.TokenType{
		"proc":   token.PROC,
		"let":    token.LET,
		"const":  token.CONST,
		"true":   token.TRUE,
		"false":  token.FALSE,
		"if":     token.IF,
		"else":   token.ELSE,
		"return": token.RETURN,
		"loop":   token.LOOP,
		"with":   token.WITH,
		"pub":    token.PUB,
		"match":  token.MATCH,
		"case":   token.CASE,
	}

	for word, expectedType := range keywords {
		l := New(word)
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Errorf("keyword %q: expected %s, got %s", word, expectedType, tok.Type)
		}
	}
}

func TestNonKeywordIdentifiers(t *testing.T) {
	// Words that look keyword-ish but aren't
	words := []string{"letx", "iffy", "proclaim", "trueish", "matcher"}
	for _, word := range words {
		l := New(word)
		tok := l.NextToken()
		if tok.Type != token.IDENT {
			t.Errorf("%q should be IDENT, got %s", word, tok.Type)
		}
	}
}
