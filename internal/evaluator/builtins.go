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
	"format": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("format requires at least 1 argument")
			}
			formatStr, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to format must be a string")
			}

			// Simple %v style replacement?
			// Or just fmt.Sprintf wrapper?
			// fmt.Sprintf expects []interface{}. We have []object.Object.
			// We need to convert.

			goArgs := make([]interface{}, len(args)-1)
			for i, arg := range args[1:] {
				switch a := arg.(type) {
				case *object.String:
					goArgs[i] = a.Value
				case *object.Integer:
					goArgs[i] = a.Value
				case *object.Boolean:
					goArgs[i] = a.Value
				default:
					goArgs[i] = a.Inspect()
				}
			}

			return &object.String{Value: fmt.Sprintf(formatStr.Value, goArgs...)}
		},
	},
}
