package evaluator

import (
	"bytes"
	"context"
	"dush/internal/builtins"
	"dush/internal/evaluator/object"
	"dush/internal/parser/ast"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
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
				env.Update(leftIdent.Value, val)
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

		// Shell Pipes and Redirections
		if isShellOperation(node.Left, env) && (node.Operator == "|" || node.Operator == ">" || node.Operator == ">>" || node.Operator == "<") {
			return evalShellOperation(node, env)
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
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.ProcStatement:
		funcObj := &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
		env.Set(node.Name.Value, funcObj)
		return nil
	case *ast.ProcLiteral:
		return &object.Function{
			Parameters: node.Parameters,
			Body:       node.Body,
			Env:        env,
		}
	case *ast.LoopStatement:
		return evalLoopStatement(node, env)
	case *ast.WithExpression:
		return evalWithExpression(node, env)
	case *ast.CallExpression:
		// Special handling for output capture 'save'
		if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "save" {
			var buf bytes.Buffer
			oldStdout := env.Stdout
			env.Stdout = &buf

			args := evalExpressions(node.Arguments, env)

			env.Stdout = oldStdout

			if len(args) == 1 && isError(args[0]) {
				return args[0]
			}
			
			// Return stdout captured as String
			return &object.String{Value: buf.String()}
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

func isShellOperation(node ast.Expression, env *Environment) bool {
	switch n := node.(type) {
	case *ast.CommandExpression:
		return true
	case *ast.Identifier:
		// If the identifier is a variable in scope, it's not a shell command
		if _, ok := env.Get(n.Value); ok {
			return false
		}
		return true
	case *ast.InfixExpression: // for chaining multiple pipes/redirections it might be infix
		return true
	}
	return false
}

func getFileName(node ast.Expression, env *Environment) string {
	val := Eval(node, env)
	if isError(val) {
		return ""
	}
	return objectToString(val)
}

func evalShellOperation(node *ast.InfixExpression, env *Environment) object.Object {
	switch node.Operator {
	case "|":
		pr, pw := io.Pipe()

		leftEnv := NewEnclosedEnvironment(env)
		leftEnv.Stdout = pw

		rightEnv := NewEnclosedEnvironment(env)
		rightEnv.Stdin = pr

		// Run the right side in a goroutine because it might block reading from Stdin
		// Wait, actually, standard pipeline: 
		// Left writes to pw, Right reads from pr.
		// `cmd.Run()` blocks. If we run Left first, it'll block writing to pw if buffer is full.
		// If we run Right first, it'll block reading from pr until Left writes.
		// We must run them concurrently and wait for both. 
		// But in dush, `Eval` is synchronous. 

		// We can span a goroutine for the left side
		go func() {
			Eval(node.Left, leftEnv)
			pw.Close()
		}()

		res := Eval(node.Right, rightEnv)
		pr.Close()
		return res

	case ">", ">>":
		fileName := getFileName(node.Right, env)
		if fileName == "" {
			return newError("invalid file name for redirection")
		}

		flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if node.Operator == ">>" {
			flags = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		}

		f, err := os.OpenFile(fileName, flags, 0644)
		if err != nil {
			return newError("could not open file %s: %v", fileName, err)
		}
		defer f.Close()

		newEnv := NewEnclosedEnvironment(env)
		newEnv.Stdout = f

		return Eval(node.Left, newEnv)

	case "<":
		fileName := getFileName(node.Right, env)
		if fileName == "" {
			return newError("invalid file name for redirection")
		}

		f, err := os.OpenFile(fileName, os.O_RDONLY, 0)
		if err != nil {
			return newError("could not open file %s: %v", fileName, err)
		}
		defer f.Close()

		newEnv := NewEnclosedEnvironment(env)
		newEnv.Stdin = f

		return Eval(node.Left, newEnv)
	}

	return newError("unsupported shell operation: %s", node.Operator)
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

		switch src := source.(type) {
		case *object.String:
			for _, ch := range src.Value {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(node.Iterator.Value, &object.String{Value: string(ch)})

				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		case *object.Array:
			for _, elem := range src.Elements {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(node.Iterator.Value, elem)

				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		case *object.Integer:
			for i := int64(0); i < src.Value; i++ {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(node.Iterator.Value, &object.Integer{Value: i})

				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		default:
			return newError("iteration not supported on %s", source.Type())
		}
	}
	return NULL
}

func evalWithExpression(node *ast.WithExpression, env *Environment) object.Object {
	newEnv := NewEnclosedEnvironment(env)
	
	for k, vExpr := range node.EnvOverrides {
		val := Eval(vExpr, env)
		if isError(val) {
			return val
		}
		newEnv.EnvOverrides[k] = objectToString(val)
	}

	return Eval(node.Body, newEnv)
}

func evalIdentifier(node *ast.Identifier, env *Environment) object.Object {
	// 1. Try Environment (Variables, Functions)
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	// 2. Try Builtins (len, format)
	if builtin, ok := functionBuiltins[node.Value]; ok {
		return builtin
	}

	// 3. Try Command Execution (ls, git, etc.)
	// If it's a known command, execute it instead of returning string.
	// Check built-in commands
	isCmd := false
	if _, ok := builtins.GetCommand(node.Value); ok {
		isCmd = true
	} else {
		// Check system path
		_, err := exec.LookPath(node.Value)
		if err == nil {
			isCmd = true
		}
	}

	if isCmd {
		cmdNode := &ast.CommandExpression{
			Token: node.Token,
			Name:  node.Value,
			Args:  []ast.Expression{},
		}
		return evalCommandExpression(cmdNode, env)
	}

	// 4. Fallback: Treat as String Literal (Bare word)
	return &object.String{Value: node.Value}
}
func evalCommandExpression(node *ast.CommandExpression, env *Environment) object.Object {
	var args []string
	for _, argExpr := range node.Args {
		var argStr string

		// Handle Identifiers: Try Env, fallback to Literal Name
		if ident, ok := argExpr.(*ast.Identifier); ok {
			if obj, ok := env.Get(ident.Value); ok {
				argStr = objectToString(obj)
			} else {
				argStr = ident.Value
			}
		} else if prefix, ok := argExpr.(*ast.PrefixExpression); ok && prefix.Operator == "-" {
			// Handle Flags (Prefix Expressions like -m)
			if ident, ok := prefix.Right.(*ast.Identifier); ok {
				argStr = "-" + ident.Value
			} else if integer, ok := prefix.Right.(*ast.IntegerLiteral); ok {
				argStr = fmt.Sprintf("-%d", integer.Value)
			} else {
				val := Eval(argExpr, env)
				if isError(val) { return val }
				argStr = objectToString(val)
			}
		} else {
			// Evaluate normally
			val := Eval(argExpr, env)
			if isError(val) {
				return val
			}
			argStr = objectToString(val)
		}

		// Apply Globbing if it contains wildcards
		if strings.ContainsAny(argStr, "*?") {
			matches, err := filepath.Glob(argStr)
			if err == nil && len(matches) > 0 {
				args = append(args, matches...)
				continue
			}
		}

		args = append(args, argStr)
	}

	// Check Registered Builtins (ls, cd, echo, etc.)
	if cmd, ok := builtins.GetCommand(node.Name); ok {
		// Use a background context for now since Eval doesn't propagate it
		// Ideally, we'd refactor Eval to take Context
		err := cmd.Execute(context.Background(), args, env.Stdout, env.Stderr)
		var exitCode int64 = 0
		if err != nil {
			// Builtins log their own errors to errOut usually, but if not:
			fmt.Fprintf(env.Stderr, "%s: %v\n", node.Name, err)
			exitCode = 1
		}
		env.Set("LAST_STATUS", &object.Integer{Value: exitCode})
		return nativeBoolToBooleanObject(exitCode == 0)
	}

	// Fallback to System Command
	cmd := exec.Command(node.Name, args...)
	cmd.Stdin = env.Stdin
	cmd.Stdout = env.Stdout
	cmd.Stderr = env.Stderr

	// Inject the custom env overrides
	overrides := env.GetAllOverrides()
	if len(overrides) > 0 {
		cmd.Env = os.Environ()
		for k, v := range overrides {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

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
	case *object.Float:
		return fmt.Sprintf("%g", obj.Value)
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
	case left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ:
		// Mixed float/int or float/float arithmetic
		if isNumeric(left) && isNumeric(right) {
			return evalFloatInfixExpression(operator, toFloat(left), toFloat(right))
		}
		if operator == "==" {
			return nativeBoolToBooleanObject(left == right)
		}
		if operator == "!=" {
			return nativeBoolToBooleanObject(left != right)
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "%":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal % rightVal}
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

func isNumeric(obj object.Object) bool {
	return obj.Type() == object.INTEGER_OBJ || obj.Type() == object.FLOAT_OBJ
}

func toFloat(obj object.Object) float64 {
	switch o := obj.(type) {
	case *object.Integer:
		return float64(o.Value)
	case *object.Float:
		return o.Value
	default:
		return 0
	}
}

func evalFloatInfixExpression(operator string, left, right float64) object.Object {
	switch operator {
	case "+":
		return &object.Float{Value: left + right}
	case "-":
		return &object.Float{Value: left - right}
	case "*":
		return &object.Float{Value: left * right}
	case "/":
		if right == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: left / right}
	case "<":
		return nativeBoolToBooleanObject(left < right)
	case ">":
		return nativeBoolToBooleanObject(left > right)
	case "==":
		return nativeBoolToBooleanObject(left == right)
	case "!=":
		return nativeBoolToBooleanObject(left != right)
	default:
		return newError("unknown operator: FLOAT %s FLOAT", operator)
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "/", "\\", ":":
		// Support path joining and drive letters
		return &object.String{Value: leftVal + operator + rightVal}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
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
	switch r := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -r.Value}
	case *object.Float:
		return &object.Float{Value: -r.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
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
