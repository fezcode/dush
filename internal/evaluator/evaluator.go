package evaluator

import (
	"context"
	"dush/internal/builtins"
	"dush/internal/evaluator/object"
	"dush/internal/parser/ast"
	"fmt"
	"os"
	"os/exec"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CommandExpression:
		return evalCommandExpression(node, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.InfixExpression:
		// Assignment special handling
		if node.Operator == "=" {
			if leftIdent, ok := node.Left.(*ast.Identifier); ok {
				val := Eval(node.Right, env)
				if isError(val) {
					return val
				}
				env.Set(leftIdent.Value, val)
				return val
			}
			return newError("left side of assignment must be an identifier")
		}

		if node.Operator == "&&" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if isTruthy(left) {
				return Eval(node.Right, env)
			}
			return left // False
		}
		if node.Operator == "||" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if isTruthy(left) {
				return left // True
			}
			return Eval(node.Right, env)
		}

		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.ProcStatement:
		funcObj := &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
		env.Set(node.Name.Value, funcObj)
		return nil
	case *ast.LoopStatement:
		return evalLoopStatement(node, env)
	case *ast.CallExpression:
		// Special form: var(name)
		if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "var" {
			if len(node.Arguments) != 1 {
				return newError("var takes 1 argument")
			}
			arg := node.Arguments[0]
			var name string
			if id, ok := arg.(*ast.Identifier); ok {
				name = id.Value
			} else if str, ok := arg.(*ast.StringLiteral); ok {
				name = str.Value
			} else {
				val := Eval(arg, env)
				if isError(val) {
					return val
				}
				if s, ok := val.(*object.String); ok {
					name = s.Value
				} else {
					return newError("var argument must be identifier or string")
				}
			}

			if val, ok := env.Get(name); ok {
				return val
			}
			// Check env vars?
			if val := os.Getenv(name); val != "" {
				return &object.String{Value: val}
			}
			return newError("variable not found: %s", name)
		}

		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	}
	return nil
}

func evalExpressions(exps []ast.Expression, env *Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch function := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return function.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *Environment {
	env := NewEnclosedEnvironment(fn.Env.(*Environment))
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalProgram(program *ast.Program, env *Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalLoopStatement(node *ast.LoopStatement, env *Environment) object.Object {
	if node.Condition != nil {
		// Conditional Loop (while)
		for {
			cond := Eval(node.Condition, env)
			if isError(cond) {
				return cond
			}
			if !isTruthy(cond) {
				break
			}

			val := Eval(node.Body, env)
			if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
				return val
			}
		}
	} else {
		// Iterator Loop
		source := Eval(node.Source, env)
		if isError(source) {
			return source
		}

		if str, ok := source.(*object.String); ok {
			for _, ch := range str.Value {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(node.Iterator.Value, &object.String{Value: string(ch)})

				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		} else {
			return newError("iteration not supported on %s", source.Type())
		}
	}
	return NULL
}

func evalIdentifier(node *ast.Identifier, env *Environment) object.Object {
	// Strict Mode: Bare identifiers are NOT variables (integers, strings, etc).
	// But they CAN be Functions (procedures).

	if val, ok := env.Get(node.Value); ok {
		if val.Type() == object.FUNCTION_OBJ {
			return val
		}
		// If it's data, we ignore it here. Must use var().
	}

	if builtin, ok := functionBuiltins[node.Value]; ok {
		return builtin
	}

	// Fallback: Treat as command (0 args)
	cmdNode := &ast.CommandExpression{
		Token: node.Token,
		Name:  node.Value,
		Args:  []ast.Expression{},
	}
	return evalCommandExpression(cmdNode, env)
}

func evalCommandExpression(node *ast.CommandExpression, env *Environment) object.Object {
	var args []string
	for _, argExpr := range node.Args {
		// Handle Identifiers: Try Env, fallback to Literal Name
		if ident, ok := argExpr.(*ast.Identifier); ok {
			if obj, ok := env.Get(ident.Value); ok {
				args = append(args, objectToString(obj))
			} else {
				args = append(args, ident.Value)
			}
			continue
		}

		// Handle Flags (Prefix Expressions like -m)
		if prefix, ok := argExpr.(*ast.PrefixExpression); ok {
			if prefix.Operator == "-" {
				if ident, ok := prefix.Right.(*ast.Identifier); ok {
					args = append(args, "-"+ident.Value)
					continue
				}
				if integer, ok := prefix.Right.(*ast.IntegerLiteral); ok {
					args = append(args, fmt.Sprintf("-%d", integer.Value))
					continue
				}
			}
		}

		val := Eval(argExpr, env)
		if isError(val) {
			return val
		}
		args = append(args, objectToString(val))
	}

	// Check Registered Builtins (ls, cd, echo, etc.)
	if cmd, ok := builtins.GetCommand(node.Name); ok {
		// Use a background context for now since Eval doesn't propagate it
		// Ideally, we'd refactor Eval to take Context
		err := cmd.Execute(context.Background(), args, env.Stdout, env.Stderr)
		var exitCode int64 = 0
		if err != nil {
			// Builtins log their own errors to errOut usually, but if not:
			// fmt.Fprintf(env.Stderr, "%s: %v\n", node.Name, err)
			exitCode = 1
		}
		env.Set("LAST_STATUS", &object.Integer{Value: exitCode})
		return nativeBoolToBooleanObject(exitCode == 0)
	}

	// Fallback to System Command
	cmd := exec.Command(node.Name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = env.Stdout
	cmd.Stderr = env.Stderr

	err := cmd.Run()
	var exitCode int64 = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = int64(exitErr.ExitCode())
		} else {
			exitCode = 1
			fmt.Fprintf(env.Stderr, "command execution failed: %s\n", err)
		}
	}

	env.Set("LAST_STATUS", &object.Integer{Value: exitCode})
	return nativeBoolToBooleanObject(exitCode == 0)
}

func objectToString(obj object.Object) string {
	switch obj := obj.(type) {
	case *object.String:
		return obj.Value
	case *object.Integer:
		return fmt.Sprintf("%d", obj.Value)
	case *object.Boolean:
		return fmt.Sprintf("%t", obj.Value)
	default:
		return obj.Inspect()
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "+" && (left.Type() == object.STRING_OBJ || right.Type() == object.STRING_OBJ):
		return &object.String{Value: objectToString(left) + objectToString(right)}
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIfExpression(ie *ast.IfExpression, env *Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}
