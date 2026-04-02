package evaluator

import (
	"dush/internal/evaluator/object"
	"dush/internal/parser/lexer"
	"dush/internal/parser/parser"
	"fmt"
	"os"
)

func EvalSource(fileName string, env *Environment) object.Object {
	content, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(env.Stderr, "source: %v\n", err)
		return nativeBoolToBooleanObject(false)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Fprintf(env.Stderr, "source: errors parsing %s:\n", fileName)
		for _, msg := range p.Errors() {
			fmt.Fprintf(env.Stderr, "\t%s\n", msg)
		}
		return nativeBoolToBooleanObject(false)
	}

	result := Eval(program, env)
	if isError(result) {
		fmt.Fprintf(env.Stderr, "source: execution error in %s: %s\n", fileName, result.Inspect())
		return nativeBoolToBooleanObject(false)
	}

	return nativeBoolToBooleanObject(true)
}
