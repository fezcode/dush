package evaluator

import (
	"dush/internal/evaluator/object"
	"fmt"
	"os"
	"strconv"
	"strings"
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
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
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
	"split": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			str, ok1 := args[0].(*object.String)
			sep, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `split` must be strings")
			}
			parts := strings.Split(str.Value, sep.Value)
			elements := make([]object.Object, len(parts))
			for i, p := range parts {
				elements[i] = &object.String{Value: p}
			}
			return &object.Array{Elements: elements}
		},
	},
	"replace": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			str, ok1 := args[0].(*object.String)
			old, ok2 := args[1].(*object.String)
			newStr, ok3 := args[2].(*object.String)
			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `replace` must be strings")
			}
			return &object.String{Value: strings.ReplaceAll(str.Value, old.Value, newStr.Value)}
		},
	},
	"to_upper": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `to_upper` must be string")
			}
			return &object.String{Value: strings.ToUpper(str.Value)}
		},
	},
	"to_lower": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `to_lower` must be string")
			}
			return &object.String{Value: strings.ToLower(str.Value)}
		},
	},
	"trim": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			str, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `trim` must be string")
			}
			return &object.String{Value: strings.TrimSpace(str.Value)}
		},
	},
	"contains": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			str, ok1 := args[0].(*object.String)
			substr, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `contains` must be strings")
			}
			if strings.Contains(str.Value, substr.Value) {
				return TRUE
			}
			return FALSE
		},
	},
	"join": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `join` must be an array")
			}
			sep, ok := args[1].(*object.String)
			if !ok {
				return newError("second argument to `join` must be a string")
			}
			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				parts[i] = el.Inspect()
			}
			return &object.String{Value: strings.Join(parts, sep.Value)}
		},
	},
	"type": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	"int": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.Float:
				return &object.Integer{Value: int64(arg.Value)}
			case *object.String:
				val, err := strconv.ParseInt(arg.Value, 10, 64)
				if err != nil {
					return newError("cannot convert %q to int", arg.Value)
				}
				return &object.Integer{Value: val}
			case *object.Boolean:
				if arg.Value {
					return &object.Integer{Value: 1}
				}
				return &object.Integer{Value: 0}
			default:
				return newError("cannot convert %s to int", args[0].Type())
			}
		},
	},
	"float": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Float:
				return arg
			case *object.Integer:
				return &object.Float{Value: float64(arg.Value)}
			case *object.String:
				val, err := strconv.ParseFloat(arg.Value, 64)
				if err != nil {
					return newError("cannot convert %q to float", arg.Value)
				}
				return &object.Float{Value: val}
			default:
				return newError("cannot convert %s to float", args[0].Type())
			}
		},
	},
	"exists": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `exists` must be string")
			}
			_, err := os.Stat(path.Value)
			if err == nil {
				return TRUE
			}
			if os.IsNotExist(err) {
				return FALSE
			}
			return FALSE
		},
	},
	"is_dir": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `is_dir` must be string")
			}
			info, err := os.Stat(path.Value)
			if err != nil {
				return FALSE
			}
			if info.IsDir() {
				return TRUE
			}
			return FALSE
		},
	},
}
