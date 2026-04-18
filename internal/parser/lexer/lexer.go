package lexer

import "dush/internal/parser/token"

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	errors       []string
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// Errors returns any errors the lexer recorded while tokenizing (e.g.
// unterminated string literals). Consumed by the parser's Errors().
func (l *Lexer) Errors() []string {
	return l.errors
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	hadSpace := l.skipWhitespace()
	tok.PrecededBySpace = hadSpace

	switch l.ch {
	case '\n':
		tok = newToken(token.SEMICOLON, l.ch)
		tok.PrecededBySpace = hadSpace
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal, PrecededBySpace: hadSpace}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
		tok.PrecededBySpace = hadSpace
	case '-':
		tok = newToken(token.MINUS, l.ch)
		tok.PrecededBySpace = hadSpace
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal, PrecededBySpace: hadSpace}
		} else {
			tok = newToken(token.BANG, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipSingleLineComment()
			return l.NextToken()
		}
		tok = newToken(token.SLASH, l.ch)
		tok.PrecededBySpace = hadSpace
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
		tok.PrecededBySpace = hadSpace
	case '%':
		tok = newToken(token.MODULO, l.ch)
		tok.PrecededBySpace = hadSpace
	case '<':
		if l.peekChar() == '<' {
			l.readChar() // consume second <
			if l.peekChar() == '<' {
				l.readChar() // consume third <
				tok = token.Token{Type: token.HERESTRING, Literal: "<<<", PrecededBySpace: hadSpace}
			} else {
				tok = token.Token{Type: token.HEREDOC, Literal: "<<", PrecededBySpace: hadSpace}
			}
		} else {
			tok = newToken(token.LT, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.APPEND, Literal: literal, PrecededBySpace: hadSpace}
		} else {
			tok = newToken(token.GT, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal, PrecededBySpace: hadSpace}
		} else if l.peekChar() == '>' {
			l.readChar() // consume >
			tok = token.Token{Type: token.ALL_GT, Literal: "&>", PrecededBySpace: hadSpace}
		} else {
			tok = newToken(token.AMPERSAND, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal, PrecededBySpace: hadSpace}
		} else {
			tok = newToken(token.PIPE, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '@':
		tok = newToken(token.AT, l.ch)
		tok.PrecededBySpace = hadSpace
	case '.':
		if l.peekChar() == '.' {
			l.readChar() // consume second .
			if l.peekChar() == '=' {
				l.readChar() // consume =
				tok = token.Token{Type: token.RANGE_EQ, Literal: "..=", PrecededBySpace: hadSpace}
			} else {
				tok = token.Token{Type: token.RANGE, Literal: "..", PrecededBySpace: hadSpace}
			}
		} else {
			tok = newToken(token.DOT, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	case '\\':
		tok = newToken(token.BACKSLASH, l.ch)
		tok.PrecededBySpace = hadSpace
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
		tok.PrecededBySpace = hadSpace
	case ':':
		tok = newToken(token.COLON, l.ch)
		tok.PrecededBySpace = hadSpace
	case ',':
		tok = newToken(token.COMMA, l.ch)
		tok.PrecededBySpace = hadSpace
	case '{':
		tok = newToken(token.LBRACE, l.ch)
		tok.PrecededBySpace = hadSpace
	case '}':
		tok = newToken(token.RBRACE, l.ch)
		tok.PrecededBySpace = hadSpace
	case '(':
		tok = newToken(token.LPAREN, l.ch)
		tok.PrecededBySpace = hadSpace
	case ')':
		tok = newToken(token.RPAREN, l.ch)
		tok.PrecededBySpace = hadSpace
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
		tok.PrecededBySpace = hadSpace
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
		tok.PrecededBySpace = hadSpace
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.PrecededBySpace = hadSpace
		return tok
	case '\'':
		tok.Type = token.RAW_STRING
		tok.Literal = l.readRawString()
		tok.PrecededBySpace = hadSpace
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.PrecededBySpace = hadSpace
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.PrecededBySpace = hadSpace
			return tok
		} else if isDigit(l.ch) {
			tok.PrecededBySpace = hadSpace
			tok.Literal = l.readNumber()
			// Check for stderr redirect: 2> or 2>>
			if tok.Literal == "2" && l.ch == '>' {
				l.readChar() // consume >
				if l.ch == '>' {
					l.readChar() // consume second >
					tok.Type = token.STDERR_APPEND
					tok.Literal = "2>>"
				} else {
					tok.Type = token.STDERR_GT
					tok.Literal = "2>"
				}
				return tok
			}
			if l.ch == '.' && isDigit(l.peekChar()) {
				tok.Literal += "."
				l.readChar()
				tok.Literal += l.readNumber()
				tok.Type = token.FLOAT
			} else {
				tok.Type = token.INT
			}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
			tok.PrecededBySpace = hadSpace
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace skips spaces, tabs, carriage returns and returns whether any were skipped.
func (l *Lexer) skipWhitespace() bool {
	skipped := false
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		skipped = true
		l.readChar()
	}
	return skipped
}

func (l *Lexer) skipSingleLineComment() {
	// Skip the initial //
	l.readChar()
	l.readChar()
	// Skip until end of line
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.skipWhitespace()
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readString reads a double-quoted string (interpolation handled by parser).
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	result := l.input[position:l.position]
	if l.ch == 0 {
		l.errors = append(l.errors, "unterminated double-quoted string")
		return result
	}
	l.readChar()
	return result
}

// readRawString reads a single-quoted string with no interpolation.
func (l *Lexer) readRawString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	result := l.input[position:l.position]
	if l.ch == 0 {
		l.errors = append(l.errors, "unterminated single-quoted string")
		return result
	}
	l.readChar()
	return result
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// isLetter defines valid identifier characters: a-z, A-Z, _
// Note: `.`, `:`, `\` are NOT included — they are separate tokens (DOT, COLON, BACKSLASH)
// for method calls and path handling in command args.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
