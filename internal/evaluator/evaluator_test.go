package evaluator

import (
	"dush/internal/evaluator/object"
	"dush/internal/parser/lexer"
	"dush/internal/parser/parser"
	"fmt"
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

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
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

func TestNestedIfElse(t *testing.T) {
	input := `if (true) { if (false) { 1 } else { 2 } } else { 3 }`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 2)
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

func TestStringExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'raw string'`, "raw string"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("input %q: expected String, got %T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
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
		{`len("")`, 0},
		{`to_upper("hello")`, "HELLO"},
		{`to_lower("HELLO")`, "hello"},
		{`trim("  hello  ")`, "hello"},
		{`replace("hello world", "world", "dush")`, "hello dush"},
		{`contains("hello world", "world")`, true},
		{`contains("hello", "xyz")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case bool:
			testBooleanObject(t, evaluated, expected)
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

func TestConstValue(t *testing.T) {
	input := "const @PI = 3\n@PI"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestPubExport(t *testing.T) {
	input := `pub @KEY = "abc123"`
	evaluated := testEval(input)
	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		t.Fatalf("unexpected error: %s", evaluated.Inspect())
	}
}

func TestPubConstExport(t *testing.T) {
	input := "pub const @VER = \"1.0\"\n@VER"
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T", evaluated)
	}
	if strObj.Value != "1.0" {
		t.Errorf("expected '1.0', got %q", strObj.Value)
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

func TestStringInterpolationMultipleVars(t *testing.T) {
	input := "@first = \"hello\"\n@second = \"world\"\n\"@first @second\""
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
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

func TestMethodContains(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`@x = "hello world"; @x.contains("world")`, true},
		{`@x = "hello world"; @x.contains("xyz")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMethodStartsWith(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`@x = "hello world"; @x.starts_with("hello")`, true},
		{`@x = "hello world"; @x.starts_with("world")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMethodEndsWith(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`@x = "hello world"; @x.ends_with("world")`, true},
		{`@x = "hello world"; @x.ends_with("hello")`, false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMethodSplit(t *testing.T) {
	input := `@x = "a,b,c"; @x.split(",")`
	evaluated := testEval(input)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		strObj := arr.Elements[i].(*object.String)
		if strObj.Value != exp {
			t.Errorf("element %d: expected %q, got %q", i, exp, strObj.Value)
		}
	}
}

func TestMethodReplaceAll(t *testing.T) {
	input := `@x = "aabaa"; @x.replace_all("a", "x")`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "xxbxx" {
		t.Errorf("expected 'xxbxx', got %q", strObj.Value)
	}
}

func TestMethodTrimStart(t *testing.T) {
	input := `@x = "  hello  "; @x.trim_start()`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "hello  " {
		t.Errorf("expected 'hello  ', got %q", strObj.Value)
	}
}

func TestMethodTrimEnd(t *testing.T) {
	input := `@x = "  hello  "; @x.trim_end()`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "  hello" {
		t.Errorf("expected '  hello', got %q", strObj.Value)
	}
}

func TestMethodSlice(t *testing.T) {
	input := `@x = "hello"; @x.slice(1, 4)`
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		errObj, isErr := evaluated.(*object.Error)
		if isErr {
			t.Fatalf("error: %s", errObj.Message)
		}
		t.Fatalf("expected String, got %T", evaluated)
	}
	if strObj.Value != "ell" {
		t.Errorf("expected 'ell', got %q", strObj.Value)
	}
}

func TestMethodOr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`@x = "hello"; @x.or("default")`, "hello"},
		{`@x = ""; @x.or("default")`, "default"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj := evaluated.(*object.String)
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
	}
}

func TestMethodToString(t *testing.T) {
	input := `@x = 42; @x.to_string()`
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("expected String, got %T", evaluated)
	}
	if strObj.Value != "42" {
		t.Errorf("expected '42', got %q", strObj.Value)
	}
}

func TestMethodAbs(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`@x = -5; @x.abs()`, 5},
		{`@x = 5; @x.abs()`, 5},
		{`@x = 0; @x.abs()`, 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestArrayMethodLen(t *testing.T) {
	input := "let @arr = split(\"a,b,c\", \",\")\n@arr.len()"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 3)
}

func TestArrayMethodJoin(t *testing.T) {
	input := `let @arr = split("a,b,c", ","); @arr.join("-")`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "a-b-c" {
		t.Errorf("expected 'a-b-c', got %q", strObj.Value)
	}
}

func TestArrayMethodContains(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"let @arr = split(\"a,b,c\", \",\")\n@arr.contains(\"b\")", true},
		{"let @arr = split(\"a,b,c\", \",\")\n@arr.contains(\"z\")", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestMethodChaining(t *testing.T) {
	input := `@x = "  HELLO  "; @x.trim().lower()`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "hello" {
		t.Errorf("expected 'hello', got %q", strObj.Value)
	}
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

func TestProcRecursion(t *testing.T) {
	input := `proc factorial(@n) {
		if (@n < 2) { return 1 }
		return @n * factorial(@n - 1)
	}
	factorial(5)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 120)
}

func TestProcClosure(t *testing.T) {
	input := `proc makeAdder(@n) {
		return proc(@x) { @x + @n }
	}
	let @addFive = makeAdder(5)
	@addFive(3)`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 8)
}

func TestLoopWithAtIterator(t *testing.T) {
	input := "@sum = 0\nloop (@i : 5) { @sum = @sum + @i }\n@sum"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10) // 0+1+2+3+4 = 10
}

func TestLoopWhileStyle(t *testing.T) {
	input := "@i = 0\n@sum = 0\nloop (@i < 5) { @sum = @sum + @i\n@i = @i + 1 }\n@sum"
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 10)
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

func TestArrayFromSplit(t *testing.T) {
	input := "let @arr = split(\"x,y,z\", \",\")\n@arr.len()"
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

func TestIntConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`int(3.7)`, 3},
		{`int("42")`, 42},
		{`int(true)`, 1},
		{`int(false)`, 0},
		{`int(5)`, 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestFloatConversion(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{`float(3)`, 3.0},
		{`float("3.14")`, 3.14},
		{`float(2.5)`, 2.5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		result := evaluated.(*object.Float)
		if result.Value != tt.expected {
			t.Errorf("input %q: expected %f, got %f", tt.input, tt.expected, result.Value)
		}
	}
}

func TestSplitFunction(t *testing.T) {
	input := `split("hello world", " ")`
	evaluated := testEval(input)
	arr, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("expected Array, got %T", evaluated)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
}

func TestJoinFunction(t *testing.T) {
	input := `join(split("a,b,c", ","), " - ")`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "a - b - c" {
		t.Errorf("expected 'a - b - c', got %q", strObj.Value)
	}
}

func TestFormatFunction(t *testing.T) {
	input := `format("Hello %s, you are %d", "Alice", 30)`
	evaluated := testEval(input)
	strObj, ok := evaluated.(*object.String)
	if !ok {
		errObj, isErr := evaluated.(*object.Error)
		if isErr {
			t.Fatalf("error: %s", errObj.Message)
		}
		t.Fatalf("expected String, got %T", evaluated)
	}
	if strObj.Value != "Hello Alice, you are 30" {
		t.Errorf("expected 'Hello Alice, you are 30', got %q", strObj.Value)
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

// === Match-Case Tests ===

func TestMatchCaseInteger(t *testing.T) {
	input := `@x = 2
match (@x) {
	case 1 { "one" }
	case 2 { "two" }
	case 3 { "three" }
}`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "two" {
		t.Errorf("expected 'two', got %q", strObj.Value)
	}
}

func TestMatchCaseString(t *testing.T) {
	input := `@cmd = "stop"
match (@cmd) {
	case "start" { 1 }
	case "stop" { 2 }
	case "restart" { 3 }
}`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 2)
}

func TestMatchCaseDefault(t *testing.T) {
	input := `@x = 99
match (@x) {
	case 1 { "one" }
	case 2 { "two" }
	case _ { "unknown" }
}`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "unknown" {
		t.Errorf("expected 'unknown', got %q", strObj.Value)
	}
}

func TestMatchCaseNoMatch(t *testing.T) {
	input := `@x = 99
match (@x) {
	case 1 { "one" }
	case 2 { "two" }
}`
	evaluated := testEval(input)
	testNullObject(t, evaluated)
}

func TestMatchCaseFirstMatchWins(t *testing.T) {
	input := `@x = 1
match (@x) {
	case 1 { "first" }
	case 1 { "second" }
}`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "first" {
		t.Errorf("expected 'first', got %q", strObj.Value)
	}
}

func TestMatchCaseBoolean(t *testing.T) {
	input := `@x = true
match (@x) {
	case false { "no" }
	case true { "yes" }
}`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "yes" {
		t.Errorf("expected 'yes', got %q", strObj.Value)
	}
}

func TestMatchCaseWithExpression(t *testing.T) {
	input := `@x = 2 + 3
match (@x) {
	case 4 { "four" }
	case 5 { "five" }
	case 6 { "six" }
}`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "five" {
		t.Errorf("expected 'five', got %q", strObj.Value)
	}
}

// === Logical Operators ===

func TestAndOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestOrOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestAndShortCircuit(t *testing.T) {
	// If && short-circuits, the second expr (which would error) never evaluates
	input := "false && @undefined_var_xyz"
	evaluated := testEval(input)
	// Should return false, not error
	testBooleanObject(t, evaluated, false)
}

func TestOrShortCircuit(t *testing.T) {
	// If || short-circuits, the second expr never evaluates
	input := "true || @undefined_var_xyz"
	evaluated := testEval(input)
	testBooleanObject(t, evaluated, true)
}

// === Error Handling ===

func TestErrorUndefinedVariable(t *testing.T) {
	input := "@nonexistent"
	evaluated := testEval(input)
	errObj, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error, got %T (%+v)", evaluated, evaluated)
	}
	if errObj.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestStringConcatWithInt(t *testing.T) {
	// dush coerces int + string to string concatenation
	input := `"count: " + "5"`
	evaluated := testEval(input)
	strObj := evaluated.(*object.String)
	if strObj.Value != "count: 5" {
		t.Errorf("expected 'count: 5', got %q", strObj.Value)
	}
}

func TestErrorDivisionByZero(t *testing.T) {
	input := `10 / 0`
	evaluated := testEval(input)
	_, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error for division by zero, got %T (%+v)", evaluated, evaluated)
	}
}

func TestErrorWrongArgCount(t *testing.T) {
	input := `len("hello", "world")`
	evaluated := testEval(input)
	_, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("expected Error for wrong arg count, got %T (%+v)", evaluated, evaluated)
	}
}

// === Scope Tests ===

func TestClosureScope(t *testing.T) {
	input := `@x = 10
proc getX() { return @x }
@x = 20
getX()`
	evaluated := testEval(input)
	// Closures capture the environment, @x was updated before call
	testIntegerObject(t, evaluated, 20)
}

func TestProcLocalScope(t *testing.T) {
	input := `@x = 10
proc shadow() {
	let @x = 99
	return @x
}
shadow()`
	evaluated := testEval(input)
	testIntegerObject(t, evaluated, 99)
}

// === Job Manager Unit Tests ===

func TestJobManagerAddAndList(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job1 := jm.Add("sleep 10", nil)
	job2 := jm.Add("ping localhost", nil)

	if job1.ID != 1 {
		t.Errorf("expected job1 ID=1, got %d", job1.ID)
	}
	if job2.ID != 2 {
		t.Errorf("expected job2 ID=2, got %d", job2.ID)
	}

	jobs := jm.List()
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobManagerGet(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job := jm.Add("test cmd", nil)
	got := jm.Get(job.ID)
	if got == nil {
		t.Fatal("expected to get job back")
	}
	if got.Command != "test cmd" {
		t.Errorf("expected 'test cmd', got %q", got.Command)
	}

	notFound := jm.Get(999)
	if notFound != nil {
		t.Error("expected nil for non-existent job")
	}
}

func TestJobManagerRemove(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job := jm.Add("test", nil)
	if !jm.Remove(job.ID) {
		t.Error("expected Remove to return true")
	}
	if jm.Remove(job.ID) {
		t.Error("expected Remove to return false for already-removed job")
	}
	if len(jm.List()) != 0 {
		t.Error("expected empty list after removal")
	}
}

func TestJobManagerMarkDone(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job := jm.Add("test", nil)
	if job.Status != JobRunning {
		t.Errorf("expected JobRunning, got %v", job.Status)
	}

	jm.MarkDone(job.ID, nil)
	if job.Status != JobDone {
		t.Errorf("expected JobDone, got %v", job.Status)
	}

	// Done channel should be closed
	select {
	case <-job.Done:
		// good
	default:
		t.Error("expected Done channel to be closed")
	}
}

func TestJobManagerMarkFailed(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job := jm.Add("test", nil)
	jm.MarkDone(job.ID, fmt.Errorf("something broke"))

	if job.Status != JobFailed {
		t.Errorf("expected JobFailed, got %v", job.Status)
	}
	if job.Error == nil {
		t.Error("expected non-nil error")
	}
}

func TestJobManagerCleanup(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	jm.Add("running", nil)
	job2 := jm.Add("done", nil)
	job3 := jm.Add("failed", nil)

	jm.MarkDone(job2.ID, nil)
	jm.MarkDone(job3.ID, fmt.Errorf("err"))

	removed := jm.Cleanup()
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	remaining := jm.List()
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining, got %d", len(remaining))
	}
	if remaining[0].Status != JobRunning {
		t.Error("remaining job should be running")
	}
}

func TestJobManagerKillNoProcess(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	job := jm.Add("test", nil)
	err := jm.Kill(job.ID)
	if err == nil {
		t.Error("expected error when killing job with no process")
	}
}

func TestJobManagerKillNonExistent(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	err := jm.Kill(999)
	if err == nil {
		t.Error("expected error for non-existent job")
	}
}

func TestJobStatusString(t *testing.T) {
	tests := []struct {
		status   JobStatus
		expected string
	}{
		{JobRunning, "running"},
		{JobDone, "done"},
		{JobFailed, "failed"},
		{JobStatus(99), "unknown"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("status %d: expected %q, got %q", tt.status, tt.expected, tt.status.String())
		}
	}
}

func TestJobListOrdering(t *testing.T) {
	jm := &JobManager{
		jobs:   make(map[int]*Job),
		nextID: 1,
	}

	jm.Add("first", nil)
	jm.Add("second", nil)
	jm.Add("third", nil)

	jobs := jm.List()
	if jobs[0].Command != "first" || jobs[1].Command != "second" || jobs[2].Command != "third" {
		t.Error("jobs should be ordered by ID")
	}
}

// === String escapes, @ARGS/@SCRIPT, .chomp(), quote-safe @{...} ===

func TestStringEscapes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"a\nb"`, "a\nb"},
		{`"a\tb"`, "a\tb"},
		{`"a\rb"`, "a\rb"},
		{`"a\0b"`, "a\x00b"},
		{`"a\\b"`, `a\b`},
		{`"a\"b"`, `a"b`},
		{`"a\'b"`, "a'b"},
		{`"\@name"`, "@name"},
		{`"\x41\x42"`, "AB"},
		{`"\u{1F600}"`, "😀"},
		{`"\u{41}"`, "A"},
		// Unknown escape is kept literal (forgiving).
		{`"\q"`, `\q`},
		// Raw single-quoted strings are untouched.
		{`'a\nb'`, `a\nb`},
		{`'\@name'`, `\@name`},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("input %q: expected String, got %T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
	}
}

func TestStringChomp(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\n".chomp()`, "hello"},
		{`"hello\r\n".chomp()`, "hello"},
		{`"hello".chomp()`, "hello"},
		{`"hello\n\n".chomp()`, "hello\n"}, // only one trailing newline is stripped
		{`"".chomp()`, ""},
		{`"\n".chomp()`, ""},
		{`"\r\n".chomp()`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("input %q: expected String, got %T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
	}
}

func TestQuoteSafeInterpolation(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Inner " inside @{...} is literal, no escaping needed.
		{`@items = ["a", "b"]; "list: @{@items.join(", ")}"`, "list: a, b"},
		// Mixed literal text + interpolation + method chain.
		{`@n = "world"; "hello @{@n.upper()}!"`, "hello WORLD!"},
		// Arithmetic expression inside @{}.
		{`"sum: @{1 + 2 + 3}"`, "sum: 6"},
		// \@ suppresses interpolation; @{} still interpolates after.
		{`@x = 7; "\@x is not @x, but @{@x} is"`, "@x is not 7, but 7 is"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		strObj, ok := evaluated.(*object.String)
		if !ok {
			t.Errorf("input %q: expected String, got %T (%+v)", tt.input, evaluated, evaluated)
			continue
		}
		if strObj.Value != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, strObj.Value)
		}
	}
}

func TestScriptArgs(t *testing.T) {
	env := NewEnvironment()
	env.SetScriptArgs("/tmp/foo.dush", []string{"one", "two", "three"})

	evalWith := func(input string) object.Object {
		l := lexer.New(input)
		p := parser.New(l)
		return Eval(p.ParseProgram(), env)
	}

	// @SCRIPT
	s, ok := evalWith(`@SCRIPT`).(*object.String)
	if !ok {
		t.Fatalf("@SCRIPT: expected String, got %T", evalWith(`@SCRIPT`))
	}
	if s.Value != "/tmp/foo.dush" {
		t.Errorf("@SCRIPT = %q, want %q", s.Value, "/tmp/foo.dush")
	}

	// @ARGS.len()
	n, ok := evalWith(`@ARGS.len()`).(*object.Integer)
	if !ok {
		t.Fatalf("@ARGS.len(): expected Integer, got %T", evalWith(`@ARGS.len()`))
	}
	if n.Value != 3 {
		t.Errorf("@ARGS.len() = %d, want 3", n.Value)
	}

	// Indexing
	first, ok := evalWith(`@ARGS[0]`).(*object.String)
	if !ok {
		t.Fatalf("@ARGS[0]: expected String, got %T", evalWith(`@ARGS[0]`))
	}
	if first.Value != "one" {
		t.Errorf("@ARGS[0] = %q, want %q", first.Value, "one")
	}

	// Join
	j, ok := evalWith(`@ARGS.join(" | ")`).(*object.String)
	if !ok {
		t.Fatalf("@ARGS.join: expected String, got %T", evalWith(`@ARGS.join(" | ")`))
	}
	if j.Value != "one | two | three" {
		t.Errorf("@ARGS.join = %q", j.Value)
	}
}

func TestScriptArgsDefaultEmpty(t *testing.T) {
	// A fresh env (no SetScriptArgs call) should have empty @ARGS and empty @SCRIPT.
	env := NewEnvironment()
	evalWith := func(input string) object.Object {
		l := lexer.New(input)
		p := parser.New(l)
		return Eval(p.ParseProgram(), env)
	}

	n, ok := evalWith(`@ARGS.len()`).(*object.Integer)
	if !ok || n.Value != 0 {
		t.Errorf("default @ARGS.len() expected 0, got %+v", evalWith(`@ARGS.len()`))
	}
	s, ok := evalWith(`@SCRIPT`).(*object.String)
	if !ok || s.Value != "" {
		t.Errorf("default @SCRIPT expected \"\", got %+v", evalWith(`@SCRIPT`))
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
