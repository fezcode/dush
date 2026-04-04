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
