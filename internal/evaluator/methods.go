package evaluator

import (
	"dush/internal/evaluator/object"
	"dush/internal/parser/ast"
	"fmt"
	"math"
	"strings"
)

func evalMethodCall(node *ast.MethodCallExpression, env *Environment) object.Object {
	obj := Eval(node.Object, env)
	if isError(obj) {
		return obj
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	switch o := obj.(type) {
	case *object.String:
		return evalStringMethod(o, node.Method, args)
	case *object.Array:
		return evalArrayMethod(o, node.Method, args)
	case *object.Integer:
		return evalIntegerMethod(o, node.Method, args)
	case *object.Float:
		return evalFloatMethod(o, node.Method, args)
	case *object.Map:
		return evalMapMethod(o, node.Method, args)
	default:
		return newError("type %s has no method '%s'", obj.Type(), node.Method)
	}
}

// --- String Methods ---

func evalStringMethod(s *object.String, method string, args []object.Object) object.Object {
	switch method {
	case "upper":
		return &object.String{Value: strings.ToUpper(s.Value)}
	case "lower":
		return &object.String{Value: strings.ToLower(s.Value)}
	case "len":
		return &object.Integer{Value: int64(len(s.Value))}
	case "trim":
		return &object.String{Value: strings.TrimSpace(s.Value)}
	case "trim_start":
		if len(args) == 1 {
			if cutset, ok := args[0].(*object.String); ok {
				return &object.String{Value: strings.TrimPrefix(s.Value, cutset.Value)}
			}
		}
		return &object.String{Value: strings.TrimLeft(s.Value, " \t\n\r")}
	case "trim_end":
		if len(args) == 1 {
			if cutset, ok := args[0].(*object.String); ok {
				return &object.String{Value: strings.TrimSuffix(s.Value, cutset.Value)}
			}
		}
		return &object.String{Value: strings.TrimRight(s.Value, " \t\n\r")}
	case "chomp":
		// Strip a single trailing \n or \r\n — handy for save() output.
		v := s.Value
		switch {
		case strings.HasSuffix(v, "\r\n"):
			v = v[:len(v)-2]
		case strings.HasSuffix(v, "\n"):
			v = v[:len(v)-1]
		}
		return &object.String{Value: v}
	case "contains":
		if len(args) != 1 {
			return newError("contains() takes 1 argument, got %d", len(args))
		}
		substr, ok := args[0].(*object.String)
		if !ok {
			return newError("contains() argument must be a string")
		}
		return nativeBoolToBooleanObject(strings.Contains(s.Value, substr.Value))
	case "starts_with":
		if len(args) != 1 {
			return newError("starts_with() takes 1 argument, got %d", len(args))
		}
		prefix, ok := args[0].(*object.String)
		if !ok {
			return newError("starts_with() argument must be a string")
		}
		return nativeBoolToBooleanObject(strings.HasPrefix(s.Value, prefix.Value))
	case "ends_with":
		if len(args) != 1 {
			return newError("ends_with() takes 1 argument, got %d", len(args))
		}
		suffix, ok := args[0].(*object.String)
		if !ok {
			return newError("ends_with() argument must be a string")
		}
		return nativeBoolToBooleanObject(strings.HasSuffix(s.Value, suffix.Value))
	case "replace":
		if len(args) != 2 {
			return newError("replace() takes 2 arguments, got %d", len(args))
		}
		old, ok1 := args[0].(*object.String)
		new, ok2 := args[1].(*object.String)
		if !ok1 || !ok2 {
			return newError("replace() arguments must be strings")
		}
		return &object.String{Value: strings.Replace(s.Value, old.Value, new.Value, 1)}
	case "replace_all":
		if len(args) != 2 {
			return newError("replace_all() takes 2 arguments, got %d", len(args))
		}
		old, ok1 := args[0].(*object.String)
		new, ok2 := args[1].(*object.String)
		if !ok1 || !ok2 {
			return newError("replace_all() arguments must be strings")
		}
		return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
	case "split":
		if len(args) != 1 {
			return newError("split() takes 1 argument, got %d", len(args))
		}
		sep, ok := args[0].(*object.String)
		if !ok {
			return newError("split() argument must be a string")
		}
		parts := strings.Split(s.Value, sep.Value)
		elements := make([]object.Object, len(parts))
		for i, p := range parts {
			elements[i] = &object.String{Value: p}
		}
		return &object.Array{Elements: elements}
	case "slice":
		if len(args) < 1 || len(args) > 2 {
			return newError("slice() takes 1-2 arguments, got %d", len(args))
		}
		start, ok := args[0].(*object.Integer)
		if !ok {
			return newError("slice() start must be an integer")
		}
		s64 := start.Value
		if s64 < 0 {
			s64 = int64(len(s.Value)) + s64
		}
		if s64 < 0 {
			s64 = 0
		}
		if s64 > int64(len(s.Value)) {
			s64 = int64(len(s.Value))
		}

		var e64 int64
		if len(args) == 2 {
			end, ok := args[1].(*object.Integer)
			if !ok {
				return newError("slice() end must be an integer")
			}
			e64 = end.Value
			if e64 < 0 {
				e64 = int64(len(s.Value)) + e64
			}
		} else {
			e64 = int64(len(s.Value))
		}
		if e64 < s64 {
			e64 = s64
		}
		if e64 > int64(len(s.Value)) {
			e64 = int64(len(s.Value))
		}
		return &object.String{Value: s.Value[s64:e64]}
	case "or":
		if len(args) != 1 {
			return newError("or() takes 1 argument, got %d", len(args))
		}
		if s.Value == "" {
			return args[0]
		}
		return s
	case "to_string":
		return s
	default:
		return newError("STRING has no method '%s'", method)
	}
}

// --- Array Methods ---

func evalArrayMethod(a *object.Array, method string, args []object.Object) object.Object {
	switch method {
	case "len":
		return &object.Integer{Value: int64(len(a.Elements))}
	case "join":
		if len(args) != 1 {
			return newError("join() takes 1 argument, got %d", len(args))
		}
		sep, ok := args[0].(*object.String)
		if !ok {
			return newError("join() argument must be a string")
		}
		parts := make([]string, len(a.Elements))
		for i, e := range a.Elements {
			parts[i] = objectToString(e)
		}
		return &object.String{Value: strings.Join(parts, sep.Value)}
	case "contains":
		if len(args) != 1 {
			return newError("contains() takes 1 argument, got %d", len(args))
		}
		target := objectToString(args[0])
		for _, e := range a.Elements {
			if objectToString(e) == target {
				return nativeBoolToBooleanObject(true)
			}
		}
		return nativeBoolToBooleanObject(false)
	case "push":
		if len(args) == 0 {
			return newError("push() requires at least 1 argument")
		}
		a.Elements = append(a.Elements, args...)
		return a
	case "pop":
		if len(a.Elements) == 0 {
			return NULL
		}
		last := a.Elements[len(a.Elements)-1]
		a.Elements = a.Elements[:len(a.Elements)-1]
		return last
	case "first":
		if len(a.Elements) == 0 {
			return NULL
		}
		return a.Elements[0]
	case "last":
		if len(a.Elements) == 0 {
			return NULL
		}
		return a.Elements[len(a.Elements)-1]
	case "slice":
		if len(args) < 1 || len(args) > 2 {
			return newError("slice() takes 1-2 arguments, got %d", len(args))
		}
		start, ok := args[0].(*object.Integer)
		if !ok {
			return newError("slice() start must be an integer")
		}
		s := start.Value
		length := int64(len(a.Elements))
		if s < 0 {
			s = length + s
		}
		if s < 0 {
			s = 0
		}
		if s > length {
			s = length
		}
		var e int64
		if len(args) == 2 {
			end, ok := args[1].(*object.Integer)
			if !ok {
				return newError("slice() end must be an integer")
			}
			e = end.Value
			if e < 0 {
				e = length + e
			}
		} else {
			e = length
		}
		if e < s {
			e = s
		}
		if e > length {
			e = length
		}
		newElements := make([]object.Object, e-s)
		copy(newElements, a.Elements[s:e])
		return &object.Array{Elements: newElements}
	case "reverse":
		newElements := make([]object.Object, len(a.Elements))
		for i, el := range a.Elements {
			newElements[len(a.Elements)-1-i] = el
		}
		return &object.Array{Elements: newElements}
	case "map":
		if len(args) != 1 {
			return newError("map() takes 1 argument (a function), got %d", len(args))
		}
		fn, ok := args[0].(*object.Function)
		if !ok {
			return newError("map() argument must be a function")
		}
		result := make([]object.Object, len(a.Elements))
		for i, el := range a.Elements {
			env := extendFunctionEnvSingle(fn, el)
			body := fn.Body.(*ast.BlockStatement)
			val := Eval(body, env)
			val = unwrapReturnValue(val)
			if isError(val) {
				return val
			}
			result[i] = val
		}
		return &object.Array{Elements: result}
	case "filter":
		if len(args) != 1 {
			return newError("filter() takes 1 argument (a function), got %d", len(args))
		}
		fn, ok := args[0].(*object.Function)
		if !ok {
			return newError("filter() argument must be a function")
		}
		result := []object.Object{}
		for _, el := range a.Elements {
			env := extendFunctionEnvSingle(fn, el)
			body := fn.Body.(*ast.BlockStatement)
			val := Eval(body, env)
			val = unwrapReturnValue(val)
			if isError(val) {
				return val
			}
			if isTruthy(val) {
				result = append(result, el)
			}
		}
		return &object.Array{Elements: result}
	default:
		return newError("ARRAY has no method '%s'", method)
	}
}

// extendFunctionEnvSingle creates a new env for a single-arg function call.
func extendFunctionEnvSingle(fn *object.Function, arg object.Object) *Environment {
	env := NewEnclosedEnvironment(fn.Env.(*Environment))
	params := fn.Parameters.([]*ast.VarExpression)
	if len(params) > 0 {
		env.Set(params[0].Name, arg)
	}
	return env
}

// --- Map Methods ---

func evalMapMethod(m *object.Map, method string, args []object.Object) object.Object {
	switch method {
	case "len":
		return &object.Integer{Value: int64(len(m.Pairs))}
	case "keys":
		keys := make([]object.Object, 0, len(m.Order))
		for _, k := range m.Order {
			keys = append(keys, m.Pairs[k].Key)
		}
		return &object.Array{Elements: keys}
	case "values":
		vals := make([]object.Object, 0, len(m.Order))
		for _, k := range m.Order {
			vals = append(vals, m.Pairs[k].Value)
		}
		return &object.Array{Elements: vals}
	case "has":
		if len(args) != 1 {
			return newError("has() takes 1 argument, got %d", len(args))
		}
		hashKey, ok := object.HashKeyFromObject(args[0])
		if !ok {
			return newError("unusable as map key: %s", args[0].Type())
		}
		_, exists := m.Pairs[hashKey]
		return nativeBoolToBooleanObject(exists)
	case "delete":
		if len(args) != 1 {
			return newError("delete() takes 1 argument, got %d", len(args))
		}
		hashKey, ok := object.HashKeyFromObject(args[0])
		if !ok {
			return newError("unusable as map key: %s", args[0].Type())
		}
		if _, exists := m.Pairs[hashKey]; exists {
			delete(m.Pairs, hashKey)
			// Remove from order
			for i, k := range m.Order {
				if k == hashKey {
					m.Order = append(m.Order[:i], m.Order[i+1:]...)
					break
				}
			}
			return TRUE
		}
		return FALSE
	case "merge":
		if len(args) != 1 {
			return newError("merge() takes 1 argument (a map), got %d", len(args))
		}
		other, ok := args[0].(*object.Map)
		if !ok {
			return newError("merge() argument must be a map")
		}
		// Create a new map with merged pairs
		newPairs := make(map[object.HashKey]object.MapPair)
		newOrder := make([]object.HashKey, len(m.Order))
		copy(newOrder, m.Order)
		for k, v := range m.Pairs {
			newPairs[k] = v
		}
		for _, k := range other.Order {
			if _, exists := newPairs[k]; !exists {
				newOrder = append(newOrder, k)
			}
			newPairs[k] = other.Pairs[k]
		}
		return &object.Map{Pairs: newPairs, Order: newOrder}
	default:
		return newError("MAP has no method '%s'", method)
	}
}

// --- Integer Methods ---

func evalIntegerMethod(i *object.Integer, method string, args []object.Object) object.Object {
	switch method {
	case "abs":
		v := i.Value
		if v < 0 {
			v = -v
		}
		return &object.Integer{Value: v}
	case "to_string":
		return &object.String{Value: fmt.Sprintf("%d", i.Value)}
	default:
		return newError("INTEGER has no method '%s'", method)
	}
}

// --- Float Methods ---

func evalFloatMethod(f *object.Float, method string, args []object.Object) object.Object {
	switch method {
	case "abs":
		return &object.Float{Value: math.Abs(f.Value)}
	case "to_string":
		return &object.String{Value: fmt.Sprintf("%g", f.Value)}
	default:
		return newError("FLOAT has no method '%s'", method)
	}
}
