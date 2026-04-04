package parser

import (
	"dush/internal/parser/ast"
	"dush/internal/parser/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	input := `let @x = 5
let @y = 10
let @foobar = 838383`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct {
		expectedName string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]
		testLetStatement(t, stmt, tt.expectedName)
	}
}

func TestConstStatement(t *testing.T) {
	input := `const @PI = 3.14`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ConstStatement)
	if !ok {
		t.Fatalf("expected ConstStatement, got %T", program.Statements[0])
	}

	if stmt.Name.Name != "PI" {
		t.Errorf("expected name 'PI', got %q", stmt.Name.Name)
	}
}

func TestPubStatement(t *testing.T) {
	input := `pub @KEY = "value"`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.PubStatement)
	if !ok {
		t.Fatalf("expected PubStatement, got %T", program.Statements[0])
	}

	if stmt.Name.Name != "KEY" {
		t.Errorf("expected name 'KEY', got %q", stmt.Name.Name)
	}
}

func TestPubConstStatement(t *testing.T) {
	input := `pub const @VER = "1.0"`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.PubStatement)
	if !ok {
		t.Fatalf("expected PubStatement, got %T", program.Statements[0])
	}

	if !stmt.IsConst {
		t.Error("expected IsConst=true")
	}
}

func TestCommandExpressions(t *testing.T) {
	tests := []struct {
		input       string
		commandName string
		numArgs     int
	}{
		{"ls -la", "ls", 1},
		{"git status", "git", 1},
		{`echo "hello world"`, "echo", 1},
		{"echo hello world", "echo", 2},
		{"echo file.txt", "echo", 1},
		{"echo -n hello", "echo", 2},
		{"atlas.ed text.txt", "atlas.ed", 1},
		{"cd .dush", "cd", 1},
		{"node.exe --version", "node.exe", 1},
		{"cd .config", "cd", 1},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Errorf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
			continue
		}

		exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Errorf("input %q: expected ExpressionStatement, got %T", tt.input, program.Statements[0])
			continue
		}

		cmd, ok := exprStmt.Expression.(*ast.CommandExpression)
		if !ok {
			t.Errorf("input %q: expected CommandExpression, got %T", tt.input, exprStmt.Expression)
			continue
		}

		if cmd.Name != tt.commandName {
			t.Errorf("input %q: expected command name %q, got %q", tt.input, tt.commandName, cmd.Name)
		}

		if len(cmd.Args) != tt.numArgs {
			t.Errorf("input %q: expected %d args, got %d", tt.input, tt.numArgs, len(cmd.Args))
		}
	}
}

func TestBug1_ChainOperatorsNotConsumedByCommandArgs(t *testing.T) {
	input := `echo "a" && echo "b"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression, got %T", exprStmt.Expression)
	}

	if infix.Operator != "&&" {
		t.Errorf("expected operator '&&', got %q", infix.Operator)
	}

	leftCmd, ok := infix.Left.(*ast.CommandExpression)
	if !ok {
		t.Fatalf("expected left to be CommandExpression, got %T", infix.Left)
	}
	if leftCmd.Name != "echo" || len(leftCmd.Args) != 1 {
		t.Errorf("expected echo with 1 arg, got %s with %d args", leftCmd.Name, len(leftCmd.Args))
	}

	rightCmd, ok := infix.Right.(*ast.CommandExpression)
	if !ok {
		t.Fatalf("expected right to be CommandExpression, got %T", infix.Right)
	}
	if rightCmd.Name != "echo" || len(rightCmd.Args) != 1 {
		t.Errorf("expected echo with 1 arg, got %s with %d args", rightCmd.Name, len(rightCmd.Args))
	}
}

func TestBug2_GroupedExpressionInCommand(t *testing.T) {
	input := `echo (1 + 2)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	cmd, ok := exprStmt.Expression.(*ast.CommandExpression)
	if !ok {
		t.Fatalf("expected CommandExpression, got %T (AST: %s)", exprStmt.Expression, program.String())
	}

	if cmd.Name != "echo" {
		t.Errorf("expected command name 'echo', got %q", cmd.Name)
	}

	if len(cmd.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(cmd.Args))
	}
}

func TestBug3_SaveNotKeyword(t *testing.T) {
	input := `save("hello")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	_, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression for save(), got %T", exprStmt.Expression)
	}
}

func TestVarExpression(t *testing.T) {
	input := `@x`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	varExpr, ok := exprStmt.Expression.(*ast.VarExpression)
	if !ok {
		t.Fatalf("expected VarExpression, got %T", exprStmt.Expression)
	}

	if varExpr.Name != "x" {
		t.Errorf("expected name 'x', got %q", varExpr.Name)
	}
}

func TestMethodCallExpression(t *testing.T) {
	input := `@x.upper()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	method, ok := exprStmt.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("expected MethodCallExpression, got %T", exprStmt.Expression)
	}

	if method.Method != "upper" {
		t.Errorf("expected method 'upper', got %q", method.Method)
	}

	varExpr, ok := method.Object.(*ast.VarExpression)
	if !ok {
		t.Fatalf("expected VarExpression as receiver, got %T", method.Object)
	}

	if varExpr.Name != "x" {
		t.Errorf("expected receiver 'x', got %q", varExpr.Name)
	}
}

func TestVarInCommandArg(t *testing.T) {
	input := `echo @name`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	cmd, ok := exprStmt.Expression.(*ast.CommandExpression)
	if !ok {
		t.Fatalf("expected CommandExpression, got %T", exprStmt.Expression)
	}

	if len(cmd.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(cmd.Args))
	}

	varExpr, ok := cmd.Args[0].(*ast.VarExpression)
	if !ok {
		t.Fatalf("expected VarExpression arg, got %T", cmd.Args[0])
	}

	if varExpr.Name != "name" {
		t.Errorf("expected 'name', got %q", varExpr.Name)
	}
}

func TestStringInterpolation(t *testing.T) {
	input := `"hello @name"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	interp, ok := exprStmt.Expression.(*ast.InterpolatedStringExpression)
	if !ok {
		t.Fatalf("expected InterpolatedStringExpression, got %T", exprStmt.Expression)
	}

	if len(interp.Parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(interp.Parts))
	}

	lit, ok := interp.Parts[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("part 0: expected StringLiteral, got %T", interp.Parts[0])
	}
	if lit.Value != "hello " {
		t.Errorf("part 0: expected 'hello ', got %q", lit.Value)
	}

	varExpr, ok := interp.Parts[1].(*ast.VarExpression)
	if !ok {
		t.Fatalf("part 1: expected VarExpression, got %T", interp.Parts[1])
	}
	if varExpr.Name != "name" {
		t.Errorf("part 1: expected 'name', got %q", varExpr.Name)
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

	if letStmt.Name.Name != name {
		t.Errorf("letStmt.Name.Name not '%s'. got=%s", name, letStmt.Name.Name)
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
