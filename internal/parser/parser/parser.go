package parser

import (
	"dush/internal/parser/ast"
	"dush/internal/parser/lexer"
	"dush/internal/parser/token"
	"fmt"
	"strconv"
)

// Precedence constants
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	LOGICAL     // && or ||
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
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
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MODULO:   PRODUCT,
	token.LPAREN:   CALL,
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
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.WITH, p.parseWithExpression)
	p.registerPrefix(token.PROC, p.parseProcLiteral) // TODO: Implement

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
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
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
	case token.PROC:
		return p.parseProcStatement()
	case token.LOOP:
		return p.parseLoopStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

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

func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Look ahead to decide if this is a Command
	// Operators like + - * / && || == != < > are typically NOT parts of command names
	// but - can be start of a flag.
	if p.peekTokenIs(token.ASSIGN) ||
		p.peekTokenIs(token.EQ) || p.peekTokenIs(token.NOT_EQ) ||
		p.peekTokenIs(token.LT) || p.peekTokenIs(token.GT) ||
		p.peekTokenIs(token.APPEND) || p.peekTokenIs(token.PIPE) ||
		p.peekTokenIs(token.AND) || p.peekTokenIs(token.OR) ||
		p.peekTokenIs(token.LPAREN) || p.peekTokenIs(token.SEMICOLON) ||
		p.peekTokenIs(token.EOF) || p.peekTokenIs(token.RPAREN) ||
		p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.COMMA) ||
		p.peekTokenIs(token.PLUS) || p.peekTokenIs(token.ASTERISK) ||
		p.peekTokenIs(token.SLASH) || p.peekTokenIs(token.MODULO) {
		return ident
	}

	// It's a command!
	cmd := &ast.CommandExpression{
		Token: ident.Token,
		Name:  ident.Value,
		Args:  []ast.Expression{},
	}

	// Consume arguments until we hit a terminator token in PEEK
	for !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.EOF) &&
		!p.peekTokenIs(token.AND) && !p.peekTokenIs(token.OR) &&
		!p.peekTokenIs(token.PIPE) && !p.peekTokenIs(token.GT) &&
		!p.peekTokenIs(token.APPEND) && !p.peekTokenIs(token.LT) &&
		!p.peekTokenIs(token.RPAREN) && !p.peekTokenIs(token.RBRACE) &&
		!p.peekTokenIs(token.PLUS) && !p.peekTokenIs(token.ASTERISK) &&
		!p.peekTokenIs(token.SLASH) && !p.peekTokenIs(token.MODULO) {

		p.nextToken() // Move to next token (the argument)

		if p.prefixParseFns[p.curToken.Type] == nil {
			// Treat as string literal
			arg := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
			cmd.Args = append(cmd.Args, arg)
			continue
		}

		arg := p.parseExpression(LOWEST)
		if arg != nil {
			cmd.Args = append(cmd.Args, arg)
		}
		// Note: parseExpression might advance tokens.
		// We rely on it stopping at the end of the expression.
		// Pratt parser guarantees curToken is the last token of expression.
		// So loop check (peek) works for next iteration.
	}

	return cmd
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

func (p *Parser) parseStringLiteral() ast.Expression {
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
		if p.curToken.Type != token.IDENT {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier in with(), got %s", p.curToken.Type))
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

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockStatement()

	return expression
}

// parseProcLiteral parses anonymous function literals: proc(x, y) { ... }
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

	// Check if iterator syntax: IDENT followed by COLON
	if p.curToken.Type == token.IDENT && p.peekToken.Type == token.COLON {
		stmt.Iterator = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		p.nextToken() // consume ident
		p.nextToken() // consume colon

		stmt.Source = p.parseExpression(LOWEST)
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

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
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

// --- Helper Functions ---

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
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
