package parser

import (
	"dush/internal/parser/ast"
	"dush/internal/parser/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5
let y = 10
let foobar = 838383
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func TestCommandExpressions(t *testing.T) {
	input := `
ls -la
git status
echo "hello world"
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	expectedCmds := []struct {
		name string
		args int
	}{
		{"ls", 1},
		{"git", 1},
		{"echo", 1},
	}

	for i, tt := range expectedCmds {
		stmt, ok := program.Statements[i].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("stmt is not ast.ExpressionStatement. got=%T", program.Statements[i])
		}

		cmd, ok := stmt.Expression.(*ast.CommandExpression)
		if !ok {
			t.Fatalf("exp is not ast.CommandExpression. got=%T", stmt.Expression)
		}

		if cmd.Name != tt.name {
			t.Errorf("cmd.Name not %s. got=%s", tt.name, cmd.Name)
		}

		if len(cmd.Args) != tt.args {
			t.Errorf("len(cmd.Args) not %d. got=%d", tt.args, len(cmd.Args))
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s", name, letStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
