package evaluator

import (
	"dush/internal/evaluator/object"
	"fmt"
)

var functionBuiltins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"print": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"var": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			// var(x) takes an Identifier? Or string name?
			// In parser, var(x) -> Call(var, [Ident(x)]).
			// Eval(Ident(x)) -> Value of x.
			// So var(x) is redundant in code context?
			// Wait, spec: "echo var(x)".
			// Command args are parsed as expressions.
			// Identifier(x) -> String "x" (fallback).
			// So Eval(Ident(x)) returns "x" (String).
			// Then var("x") is called.
			// So `var` function should take a string and look it up in the environment?
			// BUT `BuiltinFunction` doesn't have access to `env`!
			// This is a problem. Builtins usually are pure.
			// `var` needs Env access.
			return newError("var() cannot be used this way")
		},
	},
}
