package parser

import (
	"dush/internal/parser/ast"
	"dush/internal/parser/lexer"
	"dush/internal/parser/token"
	"fmt"
	"strconv"
	"strings"
)

// Precedence constants
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	LOGICAL     // && or ||
	EQUALS      // ==
	LESSGREATER // > or <
	RANGE_PREC  // .. or ..=
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	DOT_PREC    // method calls .method()
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:   ASSIGN,
	token.AND:      LOGICAL,
	token.OR:       LOGICAL,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.APPEND:   LESSGREATER,
	token.PIPE:     LESSGREATER,
	token.RANGE:    RANGE_PREC,
	token.RANGE_EQ: RANGE_PREC,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MODULO:   PRODUCT,
	token.STDERR_GT:     LESSGREATER,
	token.STDERR_APPEND: LESSGREATER,
	token.ALL_GT:        LESSGREATER,
	token.HERESTRING:    LESSGREATER,
	token.LPAREN:        CALL,
	token.LBRACKET: CALL,
	token.DOT:      DOT_PREC,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.AT, p.parseVarExpression)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.RAW_STRING, p.parseRawStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.WITH, p.parseWithExpression)
	p.registerPrefix(token.PROC, p.parseProcLiteral)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.APPEND, p.parseInfixExpression)
	p.registerInfix(token.PIPE, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseInfixExpression)
	p.registerInfix(token.RANGE, p.parseRangeExpression)
	p.registerInfix(token.RANGE_EQ, p.parseRangeExpression)
	p.registerInfix(token.STDERR_GT, p.parseInfixExpression)
	p.registerInfix(token.STDERR_APPEND, p.parseInfixExpression)
	p.registerInfix(token.ALL_GT, p.parseInfixExpression)
	p.registerInfix(token.HERESTRING, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMethodCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	lexErrs := p.l.Errors()
	if len(lexErrs) == 0 {
		return p.errors
	}
	out := make([]string, 0, len(p.errors)+len(lexErrs))
	out = append(out, lexErrs...)
	out = append(out, p.errors...)
	return out
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		// Skip empty statements (semicolons/newlines)
		if p.curToken.Type == token.SEMICOLON {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.PUB:
		return p.parsePubStatement()
	case token.PROC:
		return p.parseProcStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.BREAK:
		return &ast.BreakStatement{Token: p.curToken}
	case token.CONTINUE:
		return &ast.ContinueStatement{Token: p.curToken}
	case token.STRICT, token.TRACE, token.PIPEFAIL:
		return p.parseModeStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement: let @x = expr
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.AT) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.VarExpression{Token: p.curToken, Name: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseConstStatement: const @x = expr
func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}

	if !p.expectPeek(token.AT) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.VarExpression{Token: p.curToken, Name: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parsePubStatement: pub @x = expr, pub const @x = expr, or pub @x
func (p *Parser) parsePubStatement() *ast.PubStatement {
	stmt := &ast.PubStatement{Token: p.curToken}

	// Check for pub const
	if p.peekTokenIs(token.CONST) {
		p.nextToken() // consume const
		stmt.IsConst = true
	}

	if !p.expectPeek(token.AT) {
		return nil
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.VarExpression{Token: p.curToken, Name: p.curToken.Literal}

	// Optional assignment
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken() // consume =
		p.nextToken() // move to value

		stmt.Value = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	// Background operator: command &
	if p.peekTokenIs(token.AMPERSAND) {
		p.nextToken()
		stmt.Expression = &ast.BackgroundExpression{
			Token:      p.curToken,
			Expression: stmt.Expression,
		}
	}

	// Optional semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// --- Prefix Parsing Functions ---

// parseVarExpression: @name
func (p *Parser) parseVarExpression() ast.Expression {
	atToken := p.curToken // the @ token

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	return &ast.VarExpression{Token: atToken, Name: p.curToken.Literal}
}

func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// In the @ system, bare identifiers are ALWAYS commands (unless followed by
	// an operator that makes them an expression — like function calls or infix ops).
	// Check if this looks like a command: if peek is NOT an operator/terminator/paren, it's a command.
	// If peek is LPAREN with NO space → function call (return ident, Pratt handles call infix)
	// If peek is LPAREN WITH space → command with grouped arg (enter command mode)
	if p.peekTokenIs(token.LPAREN) && !p.peekToken.PrecededBySpace {
		return ident // Will be picked up as CallExpression by Pratt infix
	}

	// If peek is an operator or terminator, this identifier is part of an expression
	// (not a command). But SLASH/ASTERISK/MODULO without space are path separators
	// or glob patterns (build/dush.exe, ls *.go), not arithmetic — since bare
	// identifiers are never variables, `ident/something` can't be division.
	if p.peekTokenIs(token.ASSIGN) ||
		p.peekTokenIs(token.EQ) || p.peekTokenIs(token.NOT_EQ) ||
		p.peekTokenIs(token.LT) || p.peekTokenIs(token.GT) ||
		p.peekTokenIs(token.APPEND) || p.peekTokenIs(token.PIPE) ||
		p.peekTokenIs(token.AND) || p.peekTokenIs(token.OR) ||
		p.peekTokenIs(token.SEMICOLON) ||
		p.peekTokenIs(token.EOF) || p.peekTokenIs(token.RPAREN) ||
		p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.COMMA) ||
		p.peekTokenIs(token.PLUS) ||
		(p.peekTokenIs(token.ASTERISK) && p.peekToken.PrecededBySpace) ||
		(p.peekTokenIs(token.SLASH) && p.peekToken.PrecededBySpace) ||
		(p.peekTokenIs(token.MODULO) && p.peekToken.PrecededBySpace) ||
		(p.peekTokenIs(token.LBRACKET) && p.peekToken.PrecededBySpace) {
		return ident
	}

	// It's a command!
	return p.parseCommandExpression(ident)
}

// parseCommandExpression builds a CommandExpression from an identifier that starts a command.
func (p *Parser) parseCommandExpression(ident *ast.Identifier) ast.Expression {
	// Absorb path-like command names: build/dush.exe, ./cmd, C:\Users, atlas.ed, etc.
	// Concatenate adjacent non-space tokens that are part of paths (DOT, SLASH, BACKSLASH, IDENT, etc.)
	name := ident.Value
	if !p.peekToken.PrecededBySpace && isCommandNamePart(p.peekToken.Type) {
		var b strings.Builder
		b.WriteString(name)
		for !p.peekToken.PrecededBySpace && isCommandNamePart(p.peekToken.Type) {
			p.nextToken()
			b.WriteString(p.curToken.Literal)
		}
		name = b.String()
	}

	cmd := &ast.CommandExpression{
		Token: ident.Token,
		Name:  name,
		Args:  []ast.Expression{},
	}

	// Consume arguments until we hit a command terminator in PEEK
	for !isCommandTerminator(p.peekToken.Type) {
		p.nextToken()
		arg := p.parseCommandArg()
		if arg != nil {
			cmd.Args = append(cmd.Args, arg)
		}
	}

	return cmd
}

// isCommandNamePart returns true for tokens that can be part of a command name/path.
func isCommandNamePart(t token.TokenType) bool {
	switch t {
	case token.DOT, token.SLASH, token.BACKSLASH, token.IDENT,
		token.MINUS, token.COLON, token.INT, token.ASTERISK:
		return true
	}
	return false
}

// isCommandTerminator returns true for tokens that end command argument parsing.
func isCommandTerminator(t token.TokenType) bool {
	switch t {
	case token.SEMICOLON, token.EOF,
		token.AND, token.OR, token.AMPERSAND,
		token.PIPE, token.GT, token.APPEND, token.LT,
		token.STDERR_GT, token.STDERR_APPEND, token.ALL_GT,
		token.HEREDOC, token.HERESTRING,
		token.RPAREN, token.RBRACE:
		return true
	}
	return false
}

// parseCommandArg parses a single command argument.
// Absorbs adjacent (no-space) pieces into one arg so things like
// `cd "Primer (2004)"/` or `foo"bar"baz` form a single argument.
func (p *Parser) parseCommandArg() ast.Expression {
	first := p.parseCommandArgPiece()
	if first == nil {
		return nil
	}

	if isCommandTerminator(p.peekToken.Type) || p.peekToken.PrecededBySpace || p.peekToken.Type == token.EOF {
		return first
	}

	parts := []ast.Expression{first}
	for !isCommandTerminator(p.peekToken.Type) && !p.peekToken.PrecededBySpace && p.peekToken.Type != token.EOF {
		p.nextToken()
		piece := p.parseCommandArgPiece()
		if piece != nil {
			parts = append(parts, piece)
		}
	}

	return combineArgParts(parts)
}

// parseCommandArgPiece parses one piece of a command argument (string, var, word, etc.).
func (p *Parser) parseCommandArgPiece() ast.Expression {
	switch p.curToken.Type {
	case token.AT:
		if !p.peekTokenIs(token.IDENT) || p.peekToken.PrecededBySpace {
			return p.parseCommandWord()
		}
		return p.parseVarExpressionWithMethods()

	case token.STRING:
		return p.parseStringLiteral()

	case token.RAW_STRING:
		return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	case token.LPAREN:
		return p.parseGroupedExpression()

	default:
		return p.parseCommandWord()
	}
}

// combineArgParts merges multiple adjacent arg pieces into a single expression.
// Adjacent StringLiterals are concatenated; mixed pieces become an
// InterpolatedStringExpression so evaluation stringifies and joins them.
func combineArgParts(parts []ast.Expression) ast.Expression {
	if len(parts) == 1 {
		return parts[0]
	}

	var flat []ast.Expression
	for _, part := range parts {
		if ise, ok := part.(*ast.InterpolatedStringExpression); ok {
			flat = append(flat, ise.Parts...)
		} else {
			flat = append(flat, part)
		}
	}

	var merged []ast.Expression
	for _, part := range flat {
		if sl, ok := part.(*ast.StringLiteral); ok && len(merged) > 0 {
			if prev, ok := merged[len(merged)-1].(*ast.StringLiteral); ok {
				prev.Value += sl.Value
				continue
			}
		}
		merged = append(merged, part)
	}

	if len(merged) == 1 {
		return merged[0]
	}

	var tok token.Token
	switch f := parts[0].(type) {
	case *ast.StringLiteral:
		tok = f.Token
	case *ast.InterpolatedStringExpression:
		tok = f.Token
	default:
		tok = token.Token{Type: token.STRING}
	}
	return &ast.InterpolatedStringExpression{Token: tok, Parts: merged}
}

// parseVarExpressionWithMethods parses @name and then any .method() chains,
// used in command argument context where we can't rely on Pratt infix dispatch.
func (p *Parser) parseVarExpressionWithMethods() ast.Expression {
	atToken := p.curToken

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	var expr ast.Expression = &ast.VarExpression{Token: atToken, Name: p.curToken.Literal}

	// Check for method chains: .method()
	for p.peekTokenIs(token.DOT) && !p.peekToken.PrecededBySpace {
		p.nextToken() // consume DOT
		dotToken := p.curToken

		if !p.expectPeek(token.IDENT) {
			return expr
		}
		methodName := p.curToken.Literal

		var args []ast.Expression
		if p.peekTokenIs(token.LPAREN) {
			p.nextToken() // consume (
			args = p.parseCallArguments()
		}

		expr = &ast.MethodCallExpression{
			Token:     dotToken,
			Object:    expr,
			Method:    methodName,
			Arguments: args,
		}
	}

	return expr
}

// parseCommandWord concatenates adjacent non-space tokens into a single string literal.
// Handles paths (file.txt, C:\Users), flags (-la, --verbose), globs (*.go), etc.
func (p *Parser) parseCommandWord() ast.Expression {
	word := p.curToken.Literal

	for !isCommandTerminator(p.peekToken.Type) &&
		!p.peekToken.PrecededBySpace &&
		p.peekToken.Type != token.AT &&
		p.peekToken.Type != token.STRING &&
		p.peekToken.Type != token.RAW_STRING &&
		p.peekToken.Type != token.LPAREN &&
		p.peekToken.Type != token.EOF {
		p.nextToken()
		word += p.curToken.Literal
	}

	return &ast.StringLiteral{Token: p.curToken, Value: word}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral handles double-quoted strings with @ interpolation.
func (p *Parser) parseStringLiteral() ast.Expression {
	raw := p.curToken.Literal
	tok := p.curToken

	// Fast path: no interpolation and no escapes.
	if !strings.ContainsAny(raw, "@\\") {
		return &ast.StringLiteral{Token: tok, Value: raw}
	}

	// Parse interpolation: split on @ patterns
	parts := p.parseInterpolationParts(raw, tok)
	if len(parts) == 1 {
		// Optimization: if only one part, return it directly
		if sl, ok := parts[0].(*ast.StringLiteral); ok {
			return sl
		}
	}

	return &ast.InterpolatedStringExpression{Token: tok, Parts: parts}
}

// parseInterpolationParts scans a string for @name and @{expr} patterns and
// processes backslash escapes (\n, \t, \r, \0, \\, \", \', \@, \xHH, \u{...}).
func (p *Parser) parseInterpolationParts(raw string, tok token.Token) []ast.Expression {
	var parts []ast.Expression
	var lit strings.Builder
	flushLit := func() {
		if lit.Len() > 0 {
			parts = append(parts, &ast.StringLiteral{Token: tok, Value: lit.String()})
			lit.Reset()
		}
	}
	i := 0

	for i < len(raw) {
		if raw[i] == '\\' && i+1 < len(raw) {
			esc := raw[i+1]
			switch esc {
			case 'n':
				lit.WriteByte('\n')
				i += 2
				continue
			case 't':
				lit.WriteByte('\t')
				i += 2
				continue
			case 'r':
				lit.WriteByte('\r')
				i += 2
				continue
			case '0':
				lit.WriteByte(0)
				i += 2
				continue
			case '\\', '"', '\'', '@':
				lit.WriteByte(esc)
				i += 2
				continue
			case 'x':
				if i+3 < len(raw) {
					if b, err := strconv.ParseUint(raw[i+2:i+4], 16, 8); err == nil {
						lit.WriteByte(byte(b))
						i += 4
						continue
					}
				}
				p.errors = append(p.errors, fmt.Sprintf("invalid \\x escape in string literal at position %d", i))
				lit.WriteByte('\\')
				lit.WriteByte('x')
				i += 2
				continue
			case 'u':
				if i+2 < len(raw) && raw[i+2] == '{' {
					end := i + 3
					for end < len(raw) && raw[end] != '}' {
						end++
					}
					if end < len(raw) && end > i+3 {
						if cp, err := strconv.ParseUint(raw[i+3:end], 16, 32); err == nil && cp <= 0x10FFFF {
							lit.WriteRune(rune(cp))
							i = end + 1
							continue
						}
					}
				}
				p.errors = append(p.errors, fmt.Sprintf("invalid \\u{...} escape in string literal at position %d", i))
				lit.WriteByte('\\')
				lit.WriteByte('u')
				i += 2
				continue
			}
			// Unknown escape — keep the backslash literal (forgiving).
			lit.WriteByte('\\')
			lit.WriteByte(esc)
			i += 2
			continue
		}

		if raw[i] == '@' {
			i++
			if i >= len(raw) {
				// Trailing @, treat as literal
				lit.WriteByte('@')
				break
			}

			if raw[i] == '{' {
				flushLit()
				// @{expr} — find matching }
				i++ // skip {
				start := i
				depth := 1
				for i < len(raw) && depth > 0 {
					if raw[i] == '{' {
						depth++
					} else if raw[i] == '}' {
						depth--
					}
					if depth > 0 {
						i++
					}
				}
				exprStr := raw[start:i]
				if i < len(raw) {
					i++ // skip closing }
				}
				// Sub-parse the expression
				subL := lexer.New(exprStr)
				subP := New(subL)
				expr := subP.parseExpression(LOWEST)
				if expr != nil {
					parts = append(parts, expr)
				}
			} else if isIdentStart(raw[i]) {
				flushLit()
				// @name — read identifier, then check for .method()
				start := i
				for i < len(raw) && isIdentContinue(raw[i]) {
					i++
				}
				varName := raw[start:i]
				var expr ast.Expression = &ast.VarExpression{Token: tok, Name: varName}

				// Check for .method() chains
				for i < len(raw) && raw[i] == '.' {
					i++ // skip .
					mStart := i
					for i < len(raw) && isIdentContinue(raw[i]) {
						i++
					}
					if mStart == i {
						// Dot with no method name — treat the dot as literal
						i = mStart // back up past the dot
						i--        // include the dot
						break
					}
					methodName := raw[mStart:i]

					var args []ast.Expression
					if i < len(raw) && raw[i] == '(' {
						i++ // skip (
						argStart := i
						depth := 1
						for i < len(raw) && depth > 0 {
							if raw[i] == '(' {
								depth++
							} else if raw[i] == ')' {
								depth--
							}
							if depth > 0 {
								i++
							}
						}
						argStr := raw[argStart:i]
						if i < len(raw) {
							i++ // skip )
						}
						if argStr != "" {
							// Parse arguments
							subL := lexer.New(argStr)
							subP := New(subL)
							for {
								arg := subP.parseExpression(LOWEST)
								if arg != nil {
									args = append(args, arg)
								}
								if !subP.peekTokenIs(token.COMMA) {
									break
								}
								subP.nextToken() // skip comma
								subP.nextToken() // next arg
							}
						}
					}

					expr = &ast.MethodCallExpression{
						Token:     tok,
						Object:    expr,
						Method:    methodName,
						Arguments: args,
					}
				}

				parts = append(parts, expr)
			} else {
				// @ followed by non-ident char — treat @ as literal
				lit.WriteByte('@')
				lit.WriteByte(raw[i])
				i++
			}
		} else {
			// Plain text char — accumulate into the literal buffer
			lit.WriteByte(raw[i])
			i++
		}
	}

	flushLit()
	return parts
}

func isIdentStart(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isIdentContinue(ch byte) bool {
	return isIdentStart(ch) || '0' <= ch && ch <= '9'
}

func (p *Parser) parseRawStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseMatchExpression() ast.Expression {
	expression := &ast.MatchExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Subject = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Skip newlines
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// Parse cases
	for p.peekTokenIs(token.CASE) {
		p.nextToken() // consume CASE
		mc := &ast.MatchCase{Token: p.curToken}

		p.nextToken() // move to the value or _

		if p.curTokenIs(token.IDENT) && p.curToken.Literal == "_" {
			mc.IsDefault = true
		} else {
			mc.Value = p.parseExpression(LOWEST)
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		mc.Body = p.parseBlockStatement()
		expression.Cases = append(expression.Cases, mc)

		// Skip newlines between cases
		for p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return expression
}

// parseModeStatement handles `strict on`, `trace off`, `pipefail on { block }`.
func (p *Parser) parseModeStatement() *ast.ModeStatement {
	stmt := &ast.ModeStatement{Token: p.curToken, Mode: strings.ToLower(p.curToken.Literal)}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	switch strings.ToLower(p.curToken.Literal) {
	case "on":
		stmt.Enable = true
	case "off":
		stmt.Enable = false
	default:
		p.errors = append(p.errors,
			fmt.Sprintf("expected 'on' or 'off' after '%s', got %q", stmt.Mode, p.curToken.Literal))
		return nil
	}

	// Optional scoped block: `strict on { ... }`
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken() // move onto {
		stmt.Block = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	if p.curTokenIs(token.EOF) {
		p.errors = append(p.errors, "unexpected EOF")
		return nil
	}

	return block
}

func (p *Parser) parseWithExpression() ast.Expression {
	expression := &ast.WithExpression{Token: p.curToken}
	expression.EnvOverrides = make(map[string]ast.Expression)

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken() // move into arguments

	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		// Expect @NAME = expr
		if p.curToken.Type != token.AT {
			p.errors = append(p.errors, fmt.Sprintf("expected @ in with(), got %s", p.curToken.Type))
			return nil
		}

		if !p.expectPeek(token.IDENT) {
			return nil
		}

		key := p.curToken.Literal

		if !p.expectPeek(token.ASSIGN) {
			return nil
		}
		p.nextToken() // Move past '='

		val := p.parseExpression(LOWEST)
		expression.EnvOverrides[key] = val

		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // skip comma
			p.nextToken() // move to next item
		} else {
			p.nextToken() // move to next token which should be RPAREN or error
		}
	}

	if !p.curTokenIs(token.RPAREN) {
		p.errors = append(p.errors, "expected closing parenthesis for with()")
		return nil
	}

	// Support both block form: with (...) { ... }
	// and one-liner form: with (...) command args
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken() // consume {
		expression.Body = p.parseBlockStatement()
	} else {
		// One-liner: wrap the next statement in a synthetic block
		p.nextToken()
		stmt := p.parseStatement()
		if stmt != nil {
			expression.Body = &ast.BlockStatement{
				Token:      p.curToken,
				Statements: []ast.Statement{stmt},
			}
		}
	}

	return expression
}

// parseProcLiteral parses anonymous function literals: proc(@x, @y) { ... }
func (p *Parser) parseProcLiteral() ast.Expression {
	stmt := &ast.ProcStatement{Token: p.curToken}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: ""}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return &ast.ProcLiteral{Token: stmt.Token, Parameters: stmt.Parameters, Body: stmt.Body}
}

func (p *Parser) parseLoopStatement() *ast.LoopStatement {
	stmt := &ast.LoopStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken() // move to start of expression

	// Check for iterator syntax: @IDENT followed by COLON
	if p.curToken.Type == token.AT && p.peekToken.Type == token.IDENT {
		// Peek further: is it @name : source ?
		// Save state to check
		atToken := p.curToken
		p.nextToken() // consume IDENT
		identToken := p.curToken

		if p.peekTokenIs(token.COLON) {
			// Iterator loop: loop (@x : collection)
			stmt.Iterator = &ast.VarExpression{Token: atToken, Name: identToken.Literal}
			p.nextToken() // consume colon
			p.nextToken() // move to source expression

			stmt.Source = p.parseExpression(LOWEST)
		} else {
			// Not iterator — it's a condition starting with @var
			// We need to reconstruct: we already consumed AT and IDENT
			// Build a VarExpression and continue parsing as condition
			varExpr := &ast.VarExpression{Token: atToken, Name: identToken.Literal}
			// Continue parsing the rest of the condition as an infix from varExpr
			stmt.Condition = p.parseExpressionFrom(varExpr, LOWEST)
		}
	} else {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseExpressionFrom continues Pratt parsing from a pre-parsed left expression.
func (p *Parser) parseExpressionFrom(left ast.Expression, precedence int) ast.Expression {
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return left
		}

		p.nextToken()

		left = infix(left)
	}

	return left
}

func (p *Parser) parseProcStatement() *ast.ProcStatement {
	stmt := &ast.ProcStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionParameters parses proc parameters: (@x, @y, @z)
func (p *Parser) parseFunctionParameters() []*ast.VarExpression {
	params := []*ast.VarExpression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return params
	}

	// First parameter: expect @ident
	if !p.expectPeek(token.AT) {
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	params = append(params, &ast.VarExpression{Token: p.curToken, Name: p.curToken.Literal})

	for p.peekTokenIs(token.COMMA) {
		p.nextToken() // skip comma
		if !p.expectPeek(token.AT) {
			return nil
		}
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		params = append(params, &ast.VarExpression{Token: p.curToken, Name: p.curToken.Literal})
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return params
}

// --- Infix Parsing Functions ---

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

// parseRangeExpression: expr..expr or expr..=expr
func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	expression := &ast.RangeExpression{
		Token:     p.curToken,
		Start:     left,
		Inclusive: p.curToken.Type == token.RANGE_EQ,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.End = p.parseExpression(precedence)

	return expression
}

// parseMethodCallExpression: expr.method() or expr.method(args)
func (p *Parser) parseMethodCallExpression(left ast.Expression) ast.Expression {
	dotToken := p.curToken

	// Accept IDENT or any keyword as a method name (e.g. `.match(...)` where
	// `match` is also a statement keyword). The lexer will have classified it,
	// but at the method-call position any identifier-shaped token is valid.
	if !p.peekTokenIsIdentLike() {
		p.peekError(token.IDENT)
		return left
	}
	p.nextToken()
	methodName := p.curToken.Literal

	var args []ast.Expression
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // consume (
		args = p.parseCallArguments()
	}

	return &ast.MethodCallExpression{
		Token:     dotToken,
		Object:    left,
		Method:    methodName,
		Arguments: args,
	}
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// parseArrayLiteral: [expr, expr, ...]
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseIndexExpression: expr[index]
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseMapLiteral: {key: value, key: value, ...}
func (p *Parser) parseMapLiteral() ast.Expression {
	ml := &ast.MapLiteral{Token: p.curToken}
	ml.Pairs = make(map[ast.Expression]ast.Expression)
	ml.Order = []ast.Expression{}

	// Skip newlines after {
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return ml
	}

	for !p.peekTokenIs(token.RBRACE) && !p.peekTokenIs(token.EOF) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		ml.Pairs[key] = value
		ml.Order = append(ml.Order, key)

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			// Skip newlines after comma
			for p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		} else {
			// Skip newlines before closing brace
			for p.peekTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return ml
}

// parseExpressionList parses a comma-separated list of expressions until end token.
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// --- Helper Functions ---

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// peekTokenIsIdentLike reports whether the peek token looks like an identifier,
// including keywords (so `.match(...)` parses even though `match` is a keyword).
func (p *Parser) peekTokenIsIdentLike() bool {
	if p.peekToken.Type == token.IDENT {
		return true
	}
	lit := p.peekToken.Literal
	if lit == "" {
		return false
	}
	for _, r := range lit {
		if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
