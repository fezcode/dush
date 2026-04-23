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

func TestLetWithStringValue(t *testing.T) {
	input := `let @name = "alice"`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", program.Statements[0])
	}
	if stmt.Name.Name != "name" {
		t.Errorf("expected name 'name', got %q", stmt.Name.Name)
	}
	strLit, ok := stmt.Value.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral value, got %T", stmt.Value)
	}
	if strLit.Value != "alice" {
		t.Errorf("expected 'alice', got %q", strLit.Value)
	}
}

func TestLetWithBoolValue(t *testing.T) {
	input := `let @flag = true`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.LetStatement)
	boolLit, ok := stmt.Value.(*ast.BooleanLiteral)
	if !ok {
		t.Fatalf("expected BooleanLiteral, got %T", stmt.Value)
	}
	if !boolLit.Value {
		t.Error("expected true")
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
		{"mkdir -p dir/subdir", "mkdir", 2},
		{"git commit -m message", "git", 3},
		{"echo 'raw string arg'", "echo", 1},
		{`cd "Primer (2004) [WEBRip]"/`, "cd", 1},
		{`cd "dir with spaces"/sub`, "cd", 1},
		{`echo pre"mid"post`, "echo", 1},
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

// TestAdjacentArgPieces verifies that string/word/raw-string/var pieces with
// no space between them are merged into a single command argument. This is the
// "cd: too many arguments" bug — tab-completion produces `cd "dir"/` and the
// trailing slash was landing in a separate arg.
func TestAdjacentArgPieces(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		cmdName string
		// expectedArgs: the string value each arg stringifies to (StringLiteral.Value
		// or InterpolatedStringExpression.String() minus wrapping quotes).
		expectedArgs []string
	}{
		{
			name:         "quoted dir with trailing slash (reported bug)",
			input:        `cd "Primer (2004) [WEBRip] [1080p] [YTS.AM]"/`,
			cmdName:      "cd",
			expectedArgs: []string{`Primer (2004) [WEBRip] [1080p] [YTS.AM]/`},
		},
		{
			name:         "quoted dir followed by subpath",
			input:        `cd "dir with spaces"/sub/path`,
			cmdName:      "cd",
			expectedArgs: []string{`dir with spaces/sub/path`},
		},
		{
			name:         "word then string then word",
			input:        `echo pre"mid"post`,
			cmdName:      "echo",
			expectedArgs: []string{`premidpost`},
		},
		{
			name:         "word followed by string",
			input:        `echo foo"bar"`,
			cmdName:      "echo",
			expectedArgs: []string{`foobar`},
		},
		{
			name:         "two adjacent strings",
			input:        `echo "a""b"`,
			cmdName:      "echo",
			expectedArgs: []string{`ab`},
		},
		{
			name:         "raw string with trailing slash",
			input:        `cd 'some dir'/`,
			cmdName:      "cd",
			expectedArgs: []string{`some dir/`},
		},
		{
			name:         "raw string then word",
			input:        `cd 'some dir'/subdir`,
			cmdName:      "cd",
			expectedArgs: []string{`some dir/subdir`},
		},
		{
			name:         "double and raw string adjacent",
			input:        `echo "x"'y'z`,
			cmdName:      "echo",
			expectedArgs: []string{`xyz`},
		},
		{
			name:         "space separates into distinct args",
			input:        `cd "dir one" "dir two"`,
			cmdName:      "cd",
			expectedArgs: []string{`dir one`, `dir two`},
		},
		{
			name:         "mixed: adjacent joins, spaces split",
			input:        `cp "src dir"/file.txt "dst dir"/out`,
			cmdName:      "cp",
			expectedArgs: []string{`src dir/file.txt`, `dst dir/out`},
		},
		{
			name:         "plain word args untouched",
			input:        `cd folder/subfolder`,
			cmdName:      "cd",
			expectedArgs: []string{`folder/subfolder`},
		},
		{
			name:         "trailing slash on bare word (regression guard)",
			input:        `cd folder/`,
			cmdName:      "cd",
			expectedArgs: []string{`folder/`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			exprStmt, ok := program.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			cmd, ok := exprStmt.Expression.(*ast.CommandExpression)
			if !ok {
				t.Fatalf("expected CommandExpression, got %T", exprStmt.Expression)
			}

			if cmd.Name != tt.cmdName {
				t.Errorf("command name: expected %q, got %q", tt.cmdName, cmd.Name)
			}

			if len(cmd.Args) != len(tt.expectedArgs) {
				t.Fatalf("expected %d args, got %d (%v)", len(tt.expectedArgs), len(cmd.Args), argStrings(cmd.Args))
			}

			for i, want := range tt.expectedArgs {
				got := argPlainString(cmd.Args[i])
				if got != want {
					t.Errorf("arg %d: expected %q, got %q (type %T)", i, want, got, cmd.Args[i])
				}
			}
		})
	}
}

// TestAdjacentArgWithVariable ensures variable interpolation still works when
// joined with adjacent literal pieces (e.g. @base"/file.txt").
func TestAdjacentArgWithVariable(t *testing.T) {
	input := `echo @dir/file.txt`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	cmd, ok := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.CommandExpression)
	if !ok {
		t.Fatalf("expected CommandExpression")
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(cmd.Args))
	}
	ise, ok := cmd.Args[0].(*ast.InterpolatedStringExpression)
	if !ok {
		t.Fatalf("expected InterpolatedStringExpression, got %T", cmd.Args[0])
	}
	if len(ise.Parts) < 2 {
		t.Fatalf("expected at least 2 parts (var + literal), got %d", len(ise.Parts))
	}
	if _, ok := ise.Parts[0].(*ast.VarExpression); !ok {
		t.Errorf("expected first part to be VarExpression, got %T", ise.Parts[0])
	}
}

// argPlainString extracts the flattened string value from a command arg,
// unwrapping StringLiteral and rendering InterpolatedStringExpression parts
// as concatenated literal text (for literal-only test cases).
func argPlainString(e ast.Expression) string {
	switch v := e.(type) {
	case *ast.StringLiteral:
		return v.Value
	case *ast.InterpolatedStringExpression:
		var out string
		for _, p := range v.Parts {
			if sl, ok := p.(*ast.StringLiteral); ok {
				out += sl.Value
			} else {
				out += "<" + p.String() + ">"
			}
		}
		return out
	default:
		return e.String()
	}
}

func argStrings(args []ast.Expression) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = argPlainString(a)
	}
	return out
}

func TestDottedCommandNames(t *testing.T) {
	tests := []struct {
		input string
		name  string
		args  int
	}{
		{"atlas.ed", "atlas.ed", 0},
		{"atlas.ed file.txt", "atlas.ed", 1},
		{"node.exe script.js", "node.exe", 1},
		{"a.b.c arg1 arg2", "a.b.c", 2},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		cmd, ok := exprStmt.Expression.(*ast.CommandExpression)
		if !ok {
			t.Errorf("input %q: expected CommandExpression, got %T", tt.input, exprStmt.Expression)
			continue
		}
		if cmd.Name != tt.name {
			t.Errorf("input %q: expected name %q, got %q", tt.input, tt.name, cmd.Name)
		}
		if len(cmd.Args) != tt.args {
			t.Errorf("input %q: expected %d args, got %d", tt.input, tt.args, len(cmd.Args))
		}
	}
}

func TestDotPrefixedArgs(t *testing.T) {
	tests := []struct {
		input   string
		cmdName string
		arg0    string
	}{
		{"cd .dush", "cd", ".dush"},
		{"cd .config", "cd", ".config"},
		{"ls .hidden", "ls", ".hidden"},
		{"cat .gitignore", "cat", ".gitignore"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		cmd := exprStmt.Expression.(*ast.CommandExpression)

		if cmd.Name != tt.cmdName {
			t.Errorf("input %q: expected cmd %q, got %q", tt.input, tt.cmdName, cmd.Name)
		}
		if len(cmd.Args) != 1 {
			t.Errorf("input %q: expected 1 arg, got %d", tt.input, len(cmd.Args))
			continue
		}
		argStr, ok := cmd.Args[0].(*ast.StringLiteral)
		if !ok {
			t.Errorf("input %q: expected StringLiteral arg, got %T", tt.input, cmd.Args[0])
			continue
		}
		if argStr.Value != tt.arg0 {
			t.Errorf("input %q: expected arg %q, got %q", tt.input, tt.arg0, argStr.Value)
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

func TestOrChainOperator(t *testing.T) {
	input := `cmd1 arg1 || cmd2 arg2`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression, got %T", exprStmt.Expression)
	}

	if infix.Operator != "||" {
		t.Errorf("expected '||', got %q", infix.Operator)
	}

	leftCmd := infix.Left.(*ast.CommandExpression)
	if leftCmd.Name != "cmd1" || len(leftCmd.Args) != 1 {
		t.Errorf("left: expected cmd1 with 1 arg, got %s with %d", leftCmd.Name, len(leftCmd.Args))
	}

	rightCmd := infix.Right.(*ast.CommandExpression)
	if rightCmd.Name != "cmd2" || len(rightCmd.Args) != 1 {
		t.Errorf("right: expected cmd2 with 1 arg, got %s with %d", rightCmd.Name, len(rightCmd.Args))
	}
}

func TestPipeOperator(t *testing.T) {
	input := `ls -la | grep foo`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression for pipe, got %T", exprStmt.Expression)
	}

	if infix.Operator != "|" {
		t.Errorf("expected '|', got %q", infix.Operator)
	}

	leftCmd := infix.Left.(*ast.CommandExpression)
	if leftCmd.Name != "ls" {
		t.Errorf("left command: expected 'ls', got %q", leftCmd.Name)
	}

	rightCmd := infix.Right.(*ast.CommandExpression)
	if rightCmd.Name != "grep" {
		t.Errorf("right command: expected 'grep', got %q", rightCmd.Name)
	}
}

func TestRedirectOperator(t *testing.T) {
	input := `echo hello > output.txt`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression for redirect, got %T", exprStmt.Expression)
	}

	if infix.Operator != ">" {
		t.Errorf("expected '>', got %q", infix.Operator)
	}
}

func TestAppendOperator(t *testing.T) {
	input := `echo hello >> log.txt`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression for append, got %T", exprStmt.Expression)
	}

	if infix.Operator != ">>" {
		t.Errorf("expected '>>', got %q", infix.Operator)
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

func TestMethodChaining(t *testing.T) {
	input := `@x.trim().upper()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	outerMethod, ok := exprStmt.Expression.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("expected MethodCallExpression, got %T", exprStmt.Expression)
	}

	if outerMethod.Method != "upper" {
		t.Errorf("outer method: expected 'upper', got %q", outerMethod.Method)
	}

	innerMethod, ok := outerMethod.Object.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("expected inner MethodCallExpression, got %T", outerMethod.Object)
	}

	if innerMethod.Method != "trim" {
		t.Errorf("inner method: expected 'trim', got %q", innerMethod.Method)
	}
}

func TestMethodWithArgs(t *testing.T) {
	input := `@x.replace("a", "b")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	method := exprStmt.Expression.(*ast.MethodCallExpression)

	if method.Method != "replace" {
		t.Errorf("expected 'replace', got %q", method.Method)
	}
	if len(method.Arguments) != 2 {
		t.Errorf("expected 2 args, got %d", len(method.Arguments))
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

func TestBackgroundExpression(t *testing.T) {
	tests := []struct {
		input       string
		commandName string
		numArgs     int
	}{
		{"sleep 10 &", "sleep", 1},
		{"ping localhost &", "ping", 1},
		{"long.running.task &", "long.running.task", 0},
		{`echo "hello" &`, "echo", 1},
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

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		bg, ok := exprStmt.Expression.(*ast.BackgroundExpression)
		if !ok {
			t.Errorf("input %q: expected BackgroundExpression, got %T", tt.input, exprStmt.Expression)
			continue
		}

		cmd, ok := bg.Expression.(*ast.CommandExpression)
		if !ok {
			t.Errorf("input %q: inner expression should be CommandExpression, got %T", tt.input, bg.Expression)
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

func TestBackgroundWithPipe(t *testing.T) {
	input := `ls | grep foo &`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	bg, ok := exprStmt.Expression.(*ast.BackgroundExpression)
	if !ok {
		t.Fatalf("expected BackgroundExpression, got %T", exprStmt.Expression)
	}

	infix, ok := bg.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression inside bg, got %T", bg.Expression)
	}
	if infix.Operator != "|" {
		t.Errorf("expected pipe operator, got %q", infix.Operator)
	}
}

func TestBackgroundNoArgs(t *testing.T) {
	input := `mycommand &`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	bg, ok := exprStmt.Expression.(*ast.BackgroundExpression)
	if !ok {
		t.Fatalf("expected BackgroundExpression, got %T", exprStmt.Expression)
	}

	cmd := bg.Expression.(*ast.CommandExpression)
	if cmd.Name != "mycommand" {
		t.Errorf("expected 'mycommand', got %q", cmd.Name)
	}
	if len(cmd.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(cmd.Args))
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

func TestStringNoInterpolation(t *testing.T) {
	input := `"just a plain string"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	strLit, ok := exprStmt.Expression.(*ast.StringLiteral)
	if !ok {
		// Could also be InterpolatedString with one part
		interp, ok2 := exprStmt.Expression.(*ast.InterpolatedStringExpression)
		if ok2 && len(interp.Parts) == 1 {
			return // acceptable
		}
		t.Fatalf("expected StringLiteral, got %T", exprStmt.Expression)
	}
	if strLit.Value != "just a plain string" {
		t.Errorf("expected 'just a plain string', got %q", strLit.Value)
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (@x > 5) { @x }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExpr, ok := exprStmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("expected IfExpression, got %T", exprStmt.Expression)
	}
	if ifExpr.Alternative != nil {
		t.Error("expected no alternative")
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (@x > 5) { @x } else { 0 }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	ifExpr := exprStmt.Expression.(*ast.IfExpression)
	if ifExpr.Alternative == nil {
		t.Fatal("expected alternative block")
	}
}

func TestMatchExpression(t *testing.T) {
	input := `match (@x) {
		case 1 { "one" }
		case 2 { "two" }
		case _ { "other" }
	}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	matchExpr, ok := exprStmt.Expression.(*ast.MatchExpression)
	if !ok {
		t.Fatalf("expected MatchExpression, got %T", exprStmt.Expression)
	}

	if len(matchExpr.Cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(matchExpr.Cases))
	}

	if matchExpr.Cases[0].IsDefault {
		t.Error("first case should not be default")
	}
	if matchExpr.Cases[1].IsDefault {
		t.Error("second case should not be default")
	}
	if !matchExpr.Cases[2].IsDefault {
		t.Error("third case should be default")
	}
}

func TestMatchWithStrings(t *testing.T) {
	input := `match (@cmd) {
		case "start" { 1 }
		case "stop" { 2 }
	}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	matchExpr := exprStmt.Expression.(*ast.MatchExpression)

	if len(matchExpr.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(matchExpr.Cases))
	}
}

func TestLoopWhileStyle(t *testing.T) {
	input := `loop (@i < 10) { @i }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	loop, ok := program.Statements[0].(*ast.LoopStatement)
	if !ok {
		t.Fatalf("expected LoopStatement, got %T", program.Statements[0])
	}

	if loop.Iterator != nil {
		t.Errorf("while-style loop should have nil iterator, got %+v", loop.Iterator)
	}
}

func TestLoopForEachStyle(t *testing.T) {
	input := `loop (@item : @items) { @item }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	loop := program.Statements[0].(*ast.LoopStatement)

	if loop.Iterator == nil || loop.Iterator.Name != "item" {
		t.Errorf("expected iterator 'item', got %+v", loop.Iterator)
	}
}

func TestLoopRangeStyle(t *testing.T) {
	input := `loop (@i : 5) { @i }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	loop := program.Statements[0].(*ast.LoopStatement)

	if loop.Iterator == nil || loop.Iterator.Name != "i" {
		t.Errorf("expected iterator 'i', got %+v", loop.Iterator)
	}
}

func TestProcDeclaration(t *testing.T) {
	input := `proc greet(@name) { "hello" }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	_, ok := program.Statements[0].(*ast.ProcStatement)
	if !ok {
		t.Fatalf("expected ProcStatement, got %T", program.Statements[0])
	}
}

func TestProcLiteral(t *testing.T) {
	input := `let @fn = proc(@x) { @x + 1 }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.LetStatement)
	_, ok := stmt.Value.(*ast.ProcLiteral)
	if !ok {
		t.Fatalf("expected ProcLiteral, got %T", stmt.Value)
	}
}

func TestFunctionCall(t *testing.T) {
	input := `add(1, 2)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	call, ok := exprStmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expression)
	}
	if len(call.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(call.Arguments))
	}
}

func TestFunctionCallNoArgs(t *testing.T) {
	input := `doSomething()`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	call := exprStmt.Expression.(*ast.CallExpression)
	if len(call.Arguments) != 0 {
		t.Errorf("expected 0 arguments, got %d", len(call.Arguments))
	}
}

func TestWithExpression(t *testing.T) {
	input := `with (@PATH="/usr/bin") { echo hello }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	withExpr, ok := exprStmt.Expression.(*ast.WithExpression)
	if !ok {
		t.Fatalf("expected WithExpression, got %T", exprStmt.Expression)
	}

	if len(withExpr.EnvOverrides) != 1 {
		t.Errorf("expected 1 env override, got %d", len(withExpr.EnvOverrides))
	}
	if _, ok := withExpr.EnvOverrides["PATH"]; !ok {
		t.Error("expected PATH override")
	}
}

func TestInfixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"@x + @y", "+"},
		{"@x - @y", "-"},
		{"@x * @y", "*"},
		{"@x / @y", "/"},
		{"@x % @y", "%"},
		{"@x == @y", "=="},
		{"@x != @y", "!="},
		{"@x < @y", "<"},
		{"@x > @y", ">"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		infix, ok := exprStmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Errorf("input %q: expected InfixExpression, got %T", tt.input, exprStmt.Expression)
			continue
		}
		if infix.Operator != tt.operator {
			t.Errorf("input %q: expected operator %q, got %q", tt.input, tt.operator, infix.Operator)
		}
	}
}

func TestPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"!true", "!"},
		{"-5", "-"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		exprStmt := program.Statements[0].(*ast.ExpressionStatement)
		prefix, ok := exprStmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Errorf("input %q: expected PrefixExpression, got %T", tt.input, exprStmt.Expression)
			continue
		}
		if prefix.Operator != tt.operator {
			t.Errorf("expected %q, got %q", tt.operator, prefix.Operator)
		}
	}
}

func TestOperatorPrecedence(t *testing.T) {
	// Verify multiplication binds tighter than addition
	input := "1 + 2 * 3"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix := exprStmt.Expression.(*ast.InfixExpression)
	// Top level should be +, with right side being *
	if infix.Operator != "+" {
		t.Errorf("expected top operator '+', got %q", infix.Operator)
	}
	rightInfix, ok := infix.Right.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected right side to be InfixExpression, got %T", infix.Right)
	}
	if rightInfix.Operator != "*" {
		t.Errorf("expected right operator '*', got %q", rightInfix.Operator)
	}
}

func TestAssignmentExpression(t *testing.T) {
	input := `@x = 42`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	exprStmt := program.Statements[0].(*ast.ExpressionStatement)
	infix, ok := exprStmt.Expression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("expected InfixExpression (assignment), got %T", exprStmt.Expression)
	}
	if infix.Operator != "=" {
		t.Errorf("expected '=', got %q", infix.Operator)
	}
}

func TestMultipleStatements(t *testing.T) {
	input := "let @x = 1\nlet @y = 2\n@x + @y"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(program.Statements))
	}
}

func TestSemicolonSeparatedStatements(t *testing.T) {
	input := `@x = 1; @y = 2`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestReturnStatement(t *testing.T) {
	input := `return 42`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", program.Statements[0])
	}
	if stmt.TokenLiteral() != "return" {
		t.Errorf("expected 'return', got %q", stmt.TokenLiteral())
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
