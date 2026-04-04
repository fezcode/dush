package evaluator

import (
	"dush/internal/evaluator/object"
	"dush/internal/parser/lexer"
	"dush/internal/parser/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestProcDeclarations(t *testing.T) {
	input := `proc add(@x) { @x + 2 }`

	evaluated := testEval(input)
	if evaluated != nil && evaluated.Type() != object.NULL_OBJ {
		t.Fatalf("object is not NULL. got=%T (%+v)", evaluated, evaluated)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"proc add(@x, @y) { return @x + @y }\nadd(2, 3)", 5},
		{"proc early(@x) { if (@x > 0) { return @x }\nreturn 0 }\nearly(5)", 5},
		{"proc early(@x) { if (@x > 0) { return @x }\nreturn 0 }\nearly(-1)", 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestFloatExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"-2.5", -2.5},
		{"1.5 + 2.5", 4.0},
		{"10.0 - 3.5", 6.5},
		{"2.0 * 3.0", 6.0},
		{"7.0 / 2.0", 3.5},
		{"1 + 2.5", 3.5},
		{"10 * 0.5", 5.0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		result, ok := evaluated.(*object.Float)
		if !ok {
			t.Errorf("object is not Float. got=%T (%+v) for input %q", evaluated, evaluated, tt.input)
			continue
		}
		if result.Value != tt.expected {
			t.Errorf("wrong float value for %q. got=%f, want=%f", tt.input, result.Value, tt.expected)
		}
	}
}

func TestModuloOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"10 % 3", 1},
		{"15 % 5", 0},
		{"7 % 2", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestStringComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`"hello" == "hello"`, true},
		{`"hello" == "world"`, false},
		{`"hello" != "world"`, true},
		{`"hello" != "hello"`, false},
		{`"abc" < "def"`, true},
		{`"def" > "abc"`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestStringBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("hello")`, 5},
		{`to_upper("hello")`, "HELLO"},
		{`replace("hello world", "world", "dush")`, "hello dush"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if ok {
				t.Errorf("got error object: %s", errObj.Message)
				continue
			}
			strObj, ok := evaluated.(*object.String)
			if !ok {
				t.Errorf("object is not String. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if strObj.Value != expected {
				t.Errorf("wrong string value. expected=%q, got=%q", expected, strObj.Value)
			}
		}
	}
}

// === @ Variable System Tests ===

func TestVarAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"@x = 10\n@x", 10},
		{"@x = 5\n@x + 3", 8},
		{"@x = 10\n@x = 20\n@x", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetWithAtSyntax(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let @x = 10\n@x", 10},
		{"let @x = 5\nlet @y = 10\n@x + @y", 15},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestConstImmutability(t *testing.T) {
	input := "const @PI = 3\n@PI = 0"
	evaluated := testEval(input)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error, got %T (%+v)", evaluated, evaluated)
	}
	if errObj.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPubExport(t *testing.T) {
	input := `pub @KEY = "abc123"`
	evaluated := testEval(input)
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		t.Fatalf("unexpected error: %s", evaluated.Inspect())
	}
}

func TestStringInterpolation(t *testing.T) {
	input := "@name = \"world\"\n\"hello @name\""
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		errObj, isErr := evaluated.(*object.Error)
		if isErr {
			t.Fatalf("got error: %s", errObj.Message)
		}
		t.Fatalf("expected String, got %T (%+v)", evaluated, evaluated)
	}
	if strObj.Value != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", strObj.Value)
	}
}

func TestRawStringNoInterpolation(t *testing.T) {
	input := "@name = \"world\"\n'hello @name'"
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T (%+v)", evaluated, evaluated)
	}
	if strObj.Value != "hello @name" {
		t.Errorf("expected %q, got %q", "hello @name", strObj.Value)
	}
}

func TestMethodSyntax(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`@x = "hello"; @x.upper()`, "HELLO"},
		{`@x = "HELLO"; @x.lower()`, "hello"},
		{`@x = "  hello  "; @x.trim()`, "hello"},
		{`@x = "hello"; @x.replace("l", "r")`, "herlo"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			errObj, isErr := evaluated.(*object.Error)
			if isErr {
				t.Errorf("input %q: got error: %s", tt.input, errObj.Message)
				continue
			}
			t.Errorf("input %q: expected String, got %T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
	}
}

func TestMethodLen(t *testing.T) {
	input := `@x = "hello"; @x.len()`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 5)
}

func TestShellVarsReadOnly(t *testing.T) {
	input := `@LAST_STATUS = 99`
	evaluated := testEval(input)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error for shell var write, got %T (%+v)", evaluated, evaluated)
	}
	if errObj.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestProcWithAtParams(t *testing.T) {
	input := "proc add(@a, @b) { return @a + @b }\nadd(1, 2)"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestProcLiteral(t *testing.T) {
	input := "let @add = proc(@x, @y) { @x + @y }\n@add(3, 4)"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 7)
}

func TestLoopWithAtIterator(t *testing.T) {
	input := "@sum = 0\nloop (@i : 5) { @sum = @sum + @i }\n@sum"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10) // 0+1+2+3+4 = 10
}

func TestArrayIteration(t *testing.T) {
	input := "let @arr = split(\"a,b,c\", \",\")\nlet @result = \"\"\nloop (@x : @arr) { @result = @result + @x }\n@result"
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if strObj.Value != "abc" {
		t.Errorf("wrong value. got=%q, want=%q", strObj.Value, "abc")
	}
}

func TestArrayLen(t *testing.T) {
	input := `len(split("a,b,c", ","))`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestTypeBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`type(5)`, "INTEGER"},
		{`type(3.14)`, "FLOAT"},
		{`type("hello")`, "STRING"},
		{`type(true)`, "BOOLEAN"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("object is not String. got=%T (%+v) for input %q", evaluated, evaluated, tt.input)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("wrong value for %q. got=%q, want=%q", tt.input, strObj.Value, tt.expected)
		}
	}
}

func TestFileBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`exists(".")`, true},
		{`is_dir(".")`, true},
		{`exists("non_existent_file_12345")`, false},
		{`is_dir("non_existent_file_12345")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

// === Helpers ===

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
