package lexer

import (
	"dush/internal/parser/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5
let ten = 10

proc add(x, y) {
  x + y
}

let result = add(five, ten)
!-/*5
5 < 10 > 5

if (5 < 10) {
	return true
} else {
	return false
}

10 == 10
10 != 9
"foobar"
"foo bar"
loop (i < 10) { i = i + 1 }
loop (x : items) { echo x }
git status && go build || echo fail
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, "\n"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, "\n"},
		{token.SEMICOLON, "\n"},
		{token.PROC, "proc"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.SEMICOLON, "\n"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, "\n"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},
		{token.SEMICOLON, "\n"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
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
		{token.INT, "5"},
		{token.LT, "<"},
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
		{token.STRING, "foobar"},
		{token.SEMICOLON, "\n"},
		{token.STRING, "foo bar"},
		{token.SEMICOLON, "\n"},
		{token.LOOP, "loop"},
		{token.LPAREN, "("},
		{token.IDENT, "i"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "i"},
		{token.ASSIGN, "="},
		{token.IDENT, "i"},
		{token.PLUS, "+"},
		{token.INT, "1"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},
		{token.LOOP, "loop"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COLON, ":"},
		{token.IDENT, "items"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "echo"},
		{token.IDENT, "x"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, "\n"},
		{token.IDENT, "git"},
		{token.IDENT, "status"},
		{token.AND, "&&"},
		{token.IDENT, "go"},
		{token.IDENT, "build"},
		{token.OR, "||"},
		{token.IDENT, "echo"},
		{token.IDENT, "fail"},
		{token.SEMICOLON, "\n"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {

		tok := l.NextToken()

		if tok.Type != tt.expectedType {

			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",

				i, tt.expectedType, tok.Type)

		}

		if tok.Literal != tt.expectedLiteral {

			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",

				i, tt.expectedLiteral, tok.Literal)

		}

	}

}
