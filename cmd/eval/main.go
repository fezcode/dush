package main

import (
	"fmt"
	"dush/internal/evaluator"
	"dush/internal/evaluator/object"
	"dush/internal/parser/lexer"
	"dush/internal/parser/parser"
	"os"
	"strings"
)

func eval(input string) {
	fmt.Printf(">>> %s\n", strings.ReplaceAll(input, "\n", "\n... "))
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Printf("  PARSE ERROR: %s\n", e)
		}
		return
	}
	fmt.Printf("  AST: %s\n", program.String())
	env := evaluator.NewEnvironment()
	env.Stdout = os.Stdout
	env.Stderr = os.Stderr
	result := evaluator.Eval(program, env)
	if result != nil {
		if result.Type() == object.ERROR_OBJ {
			fmt.Printf("  EVAL ERROR: %s\n", result.Inspect())
		} else {
			fmt.Printf("  RESULT: [%s] %s\n", result.Type(), result.Inspect())
		}
	} else {
		fmt.Printf("  RESULT: nil\n")
	}
	fmt.Println()
}

func main() {
	fmt.Println("=== @ VARIABLE BASICS ===")
	eval(`@x = 10`)
	eval(`@x = "hello"`)
	eval("@x = 10\n@x + 5")
	eval("let @x = 10\n@x")
	eval("@name = \"world\"\necho @name")

	fmt.Println("=== CONST ===")
	eval(`const @PI = 3.14`)
	eval("const @PI = 3.14\n@PI = 0")

	fmt.Println("=== PUB ===")
	eval(`pub @KEY = "abc123"`)
	eval("@x = 10\npub @x")

	fmt.Println("=== STRING INTERPOLATION ===")
	eval("@name = \"world\"\necho \"hello @name\"")
	eval("@name = \"world\"\necho 'hello @name'")
	eval("@count = 3\necho \"you have @count items\"")

	fmt.Println("=== METHOD SYNTAX ===")
	eval(`@x = "hello"; @x.upper()`)
	eval(`@x = "hello world"; @x.len()`)
	eval(`@x = "hello"; @x.replace("l", "r")`)
	eval(`@x = "hello world"; @x.split(" ")`)
	eval(`@x = "hello"; @x.slice(0, 3)`)
	eval(`@x = ""; @x.or("default")`)

	fmt.Println("=== BUG 1 FIX: && not eaten by command args ===")
	eval(`echo "a" && echo "b"`)
	eval(`echo "x" || echo "y"`)

	fmt.Println("=== BUG 2 FIX: echo (expr) not a function call ===")
	eval("@x = 5\necho (@x + 1)")

	fmt.Println("=== BUG 3 FIX: save() works ===")
	eval(`let @out = save(echo "captured")`)

	fmt.Println("=== BUG 4 FIX: const parses ===")
	eval(`const @PI = 3.14`)

	fmt.Println("=== COMMAND ARGS: paths, flags ===")
	eval(`echo file.txt`)
	eval(`echo -la`)
	eval(`echo --verbose`)

	fmt.Println("=== SHELL VARIABLES ===")
	eval(`echo @LAST_STATUS`)
	eval(`echo @OS_NAME`)
	eval(`echo @SHELL_PID`)
	eval(`@LAST_STATUS = 99`)

	fmt.Println("=== WITH ===")
	eval(`with (@NODE_ENV = "production") { echo "hi" }`)

	fmt.Println("=== PROC with @ params ===")
	eval("proc add(@a, @b) { return @a + @b }\nadd(1, 2)")

	fmt.Println("=== LOOP with @ iterator ===")
	eval("@sum = 0\nloop (@i : 5) { @sum = @sum + @i }\n@sum")

	fmt.Println("=== BARE IDENTIFIERS ARE LITERALS ===")
	eval(`echo hello world`)
	eval(`echo name`)  // should print "name", not resolve a variable
}
