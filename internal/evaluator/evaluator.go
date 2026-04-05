package evaluator

import (
	"bufio"
	"bytes"
	"context"
	"dush/internal/builtins"
	"dush/internal/evaluator/object"
	"dush/internal/parser/ast"
	"errors"
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
		if errMsg := env.Set(node.Name.Name, val); errMsg != "" {
			return newError("%s", errMsg)
		}
	case *ast.ConstStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if errMsg := env.SetConst(node.Name.Name, val); errMsg != "" {
			return newError("%s", errMsg)
		}
	case *ast.PubStatement:
		if node.Value != nil {
			val := Eval(node.Value, env)
			if isError(val) {
				return val
			}
			if errMsg := env.SetPub(node.Name.Name, val, node.IsConst); errMsg != "" {
				return newError("%s", errMsg)
			}
		} else {
			if errMsg := env.MarkPub(node.Name.Name); errMsg != "" {
				return newError("%s", errMsg)
			}
		}
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.InterpolatedStringExpression:
		return evalInterpolatedString(node, env)
	case *ast.VarExpression:
		return evalVarExpression(node, env)
	case *ast.MethodCallExpression:
		return evalMethodCall(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.CommandExpression:
		return evalCommandExpression(node, env)
	case *ast.BackgroundExpression:
		return evalBackgroundExpression(node, env)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.MatchExpression:
		return evalMatchExpression(node, env)
	case *ast.InfixExpression:
		// Assignment special handling: @x = value
		if node.Operator == "=" {
			if varExpr, ok := node.Left.(*ast.VarExpression); ok {
				val := Eval(node.Right, env)
				if isError(val) {
					return val
				}
				if errMsg := env.Update(varExpr.Name, val); errMsg != "" {
					return newError("%s", errMsg)
				}
				return val
			}
			return newError("left side of assignment must be a variable (@name)")
		}

		if node.Operator == "&&" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if isTruthy(left) {
				return Eval(node.Right, env)
			}
			return left
		}
		if node.Operator == "||" {
			left := Eval(node.Left, env)
			if isError(left) {
				return left
			}
			if isTruthy(left) {
				return left
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

func evalInterpolatedString(node *ast.InterpolatedStringExpression, env *Environment) object.Object {
	var buf bytes.Buffer
	for _, part := range node.Parts {
		val := Eval(part, env)
		if isError(val) {
			return val
		}
		buf.WriteString(objectToString(val))
	}
	return &object.String{Value: buf.String()}
}

func evalVarExpression(node *ast.VarExpression, env *Environment) object.Object {
	val, ok := env.Get(node.Name)
	if !ok {
		return newError("undefined variable '@%s'", node.Name)
	}
	return val
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
	switch node.(type) {
	case *ast.CommandExpression:
		return true
	case *ast.InfixExpression:
		return true
	}
	if ident, ok := node.(*ast.Identifier); ok {
		if _, ok := functionBuiltins[ident.Value]; ok {
			return false
		}
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
		body := function.Body.(*ast.BlockStatement)
		evaluated := Eval(body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return function.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *Environment {
	env := NewEnclosedEnvironment(fn.Env.(*Environment))
	params := fn.Parameters.([]*ast.VarExpression)
	for i, param := range params {
		env.Set(param.Name, args[i])
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
		source := Eval(node.Source, env)
		if isError(source) {
			return source
		}

		iterName := node.Iterator.Name

		switch src := source.(type) {
		case *object.String:
			for _, ch := range src.Value {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(iterName, &object.String{Value: string(ch)})
				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		case *object.Array:
			for _, elem := range src.Elements {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(iterName, elem)
				val := Eval(node.Body, loopEnv)
				if val != nil && (val.Type() == object.RETURN_VALUE_OBJ || val.Type() == object.ERROR_OBJ) {
					return val
				}
			}
		case *object.Integer:
			for i := int64(0); i < src.Value; i++ {
				loopEnv := NewEnclosedEnvironment(env)
				loopEnv.Set(iterName, &object.Integer{Value: i})
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
		newEnv.SetPub(k, val, false)
	}

	return Eval(node.Body, newEnv)
}

// evalIdentifier: In the @ system, bare identifiers resolve to:
// 1. Function builtins (len, split, etc.)
// 2. User-defined procs stored in env (proc add(...) { ... })
// 3. Shell commands (fallback)
func evalIdentifier(node *ast.Identifier, env *Environment) object.Object {
	// 1. Function builtins
	if builtin, ok := functionBuiltins[node.Value]; ok {
		return builtin
	}

	// 2. User-defined procs/functions in environment
	if val, ok := env.Get(node.Value); ok {
		if val.Type() == object.FUNCTION_OBJ {
			return val
		}
	}

	// 3. Shell command
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
		var argStr string

		switch arg := argExpr.(type) {
		case *ast.VarExpression:
			val, ok := env.Get(arg.Name)
			if !ok {
				return newError("undefined variable '@%s'", arg.Name)
			}
			argStr = objectToString(val)
		case *ast.MethodCallExpression:
			val := evalMethodCall(arg, env)
			if isError(val) {
				return val
			}
			argStr = objectToString(val)
		case *ast.InterpolatedStringExpression:
			val := evalInterpolatedString(arg, env)
			if isError(val) {
				return val
			}
			argStr = objectToString(val)
		case *ast.StringLiteral:
			argStr = arg.Value
		default:
			val := Eval(argExpr, env)
			if isError(val) {
				return val
			}
			argStr = objectToString(val)
		}

		if strings.ContainsAny(argStr, "*?") {
			matches, err := filepath.Glob(argStr)
			if err == nil && len(matches) > 0 {
				args = append(args, matches...)
				continue
			}
		}

		if strings.HasPrefix(argStr, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				argStr = filepath.Join(home, strings.TrimPrefix(argStr, "~"))
			}
		}

		args = append(args, argStr)
	}

	if node.Name == "source" || node.Name == "." {
		if len(args) != 1 {
			fmt.Fprintf(env.Stderr, "%s: filename argument required\n", node.Name)
			env.ShellSet("LAST_STATUS", &object.Integer{Value: 1})
			return nativeBoolToBooleanObject(false)
		}
		return EvalSource(args[0], env)
	}

	if node.Name == "read" {
		return evalReadCommand(args, env)
	}

	if node.Name == "unset" {
		return evalUnsetCommand(args, env)
	}

	if cmd, ok := builtins.GetCommand(node.Name); ok {
		err := cmd.Execute(context.Background(), args, env.Stdout, env.Stderr)
		var exitCode int64 = 0
		if err != nil {
			fmt.Fprintf(env.Stderr, "%s: %v\n", node.Name, err)
			exitCode = 1
		}
		env.ShellSet("LAST_STATUS", &object.Integer{Value: exitCode})
		return nativeBoolToBooleanObject(exitCode == 0)
	}

	cmd := exec.Command(node.Name, args...)
	// As a shell, we intentionally allow running executables from the current directory
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	cmd.Stdin = env.Stdin
	cmd.Stdout = env.Stdout
	cmd.Stderr = env.Stderr

	exports := env.GetExportedVars()
	if len(exports) > 0 {
		cmd.Env = os.Environ()
		for k, v := range exports {
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

	env.ShellSet("LAST_STATUS", &object.Integer{Value: exitCode})
	return nativeBoolToBooleanObject(exitCode == 0)
}

func evalBackgroundExpression(node *ast.BackgroundExpression, env *Environment) object.Object {
	// Build the command description from the AST
	cmdDesc := node.Expression.String()

	// For command expressions, we can start the process without waiting
	if cmdNode, ok := node.Expression.(*ast.CommandExpression); ok {
		return evalBackgroundCommand(cmdNode, cmdDesc, env)
	}

	// For other expressions (pipes, chains), run in a goroutine
	bgEnv := NewEnclosedEnvironment(env)
	bgEnv.Stdout = env.Stdout
	bgEnv.Stderr = env.Stderr
	bgEnv.Stdin = env.Stdin

	// Create a job with no exec.Cmd (expression-based)
	job := Jobs.Add(cmdDesc, nil)
	fmt.Fprintf(env.Stderr, "[%d] started\n", job.ID)

	go func() {
		Eval(node.Expression, bgEnv)
		Jobs.MarkDone(job.ID, nil)
		fmt.Fprintf(env.Stderr, "[%d] done\t%s\n", job.ID, cmdDesc)
	}()

	env.ShellSet("LAST_STATUS", &object.Integer{Value: 0})
	return &object.Integer{Value: int64(job.ID)}
}

func evalBackgroundCommand(node *ast.CommandExpression, cmdDesc string, env *Environment) object.Object {
	// Build args the same way as evalCommandExpression
	var args []string
	for _, argExpr := range node.Args {
		var argStr string
		switch arg := argExpr.(type) {
		case *ast.VarExpression:
			val, ok := env.Get(arg.Name)
			if !ok {
				return newError("undefined variable '@%s'", arg.Name)
			}
			argStr = objectToString(val)
		case *ast.StringLiteral:
			argStr = arg.Value
		default:
			val := Eval(argExpr, env)
			if isError(val) {
				return val
			}
			argStr = objectToString(val)
		}

		if strings.ContainsAny(argStr, "*?") {
			matches, err := filepath.Glob(argStr)
			if err == nil && len(matches) > 0 {
				args = append(args, matches...)
				continue
			}
		}

		if strings.HasPrefix(argStr, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				argStr = filepath.Join(home, strings.TrimPrefix(argStr, "~"))
			}
		}

		args = append(args, argStr)
	}

	// Check if the command is a builtin — run it in a goroutine instead of exec
	if builtinCmd, ok := builtins.GetCommand(node.Name); ok {
		bgEnv := NewEnclosedEnvironment(env)
		bgEnv.Stdout = env.Stdout
		bgEnv.Stderr = env.Stderr
		bgEnv.Stdin = env.Stdin

		job := Jobs.Add(cmdDesc, nil)
		fmt.Fprintf(env.Stderr, "[%d] started\n", job.ID)

		go func() {
			err := builtinCmd.Execute(context.Background(), args, bgEnv.Stdout, bgEnv.Stderr)
			Jobs.MarkDone(job.ID, err)
		}()

		env.ShellSet("LAST_STATUS", &object.Integer{Value: 0})
		return &object.Integer{Value: int64(job.ID)}
	}

	cmd := exec.Command(node.Name, args...)
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	cmd.Stdout = env.Stdout
	cmd.Stderr = env.Stderr

	exports := env.GetExportedVars()
	if len(exports) > 0 {
		cmd.Env = os.Environ()
		for k, v := range exports {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	if err := cmd.Start(); err != nil {
		return newError("failed to start background job: %s", err)
	}

	job := Jobs.Add(cmdDesc, cmd)
	fmt.Fprintf(env.Stderr, "[%d] %d\n", job.ID, cmd.Process.Pid)

	go func() {
		err := cmd.Wait()
		Jobs.MarkDone(job.ID, err)
	}()

	env.ShellSet("LAST_STATUS", &object.Integer{Value: 0})
	return &object.Integer{Value: int64(job.ID)}
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

func evalMatchExpression(me *ast.MatchExpression, env *Environment) object.Object {
	subject := Eval(me.Subject, env)
	if isError(subject) {
		return subject
	}

	for _, mc := range me.Cases {
		if mc.IsDefault {
			return Eval(mc.Body, env)
		}

		caseVal := Eval(mc.Value, env)
		if isError(caseVal) {
			return caseVal
		}

		if objectsEqual(subject, caseVal) {
			return Eval(mc.Body, env)
		}
	}

	return NULL
}

func objectsEqual(a, b object.Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *object.Integer:
		return av.Value == b.(*object.Integer).Value
	case *object.Float:
		return av.Value == b.(*object.Float).Value
	case *object.String:
		return av.Value == b.(*object.String).Value
	case *object.Boolean:
		return av.Value == b.(*object.Boolean).Value
	default:
		return a == b
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

// evalReadCommand reads a line from stdin and assigns words to @ variables.
// Usage: read @var1 @var2 ...
// If fewer variables than words, the last variable gets the remaining words.
// If no variables given, stores the line in @REPLY.
func evalReadCommand(args []string, env *Environment) object.Object {
	reader := bufio.NewReader(env.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		env.ShellSet("LAST_STATUS", &object.Integer{Value: 1})
		return nativeBoolToBooleanObject(false)
	}
	line = strings.TrimRight(line, "\r\n")

	// Strip @ prefix from variable names
	varNames := make([]string, 0, len(args))
	for _, a := range args {
		varNames = append(varNames, strings.TrimPrefix(a, "@"))
	}

	if len(varNames) == 0 {
		env.Set("REPLY", &object.String{Value: line})
		env.ShellSet("LAST_STATUS", &object.Integer{Value: 0})
		return nativeBoolToBooleanObject(true)
	}

	words := strings.Fields(line)

	for i, name := range varNames {
		if i == len(varNames)-1 {
			// Last variable gets the rest
			if i < len(words) {
				env.Set(name, &object.String{Value: strings.Join(words[i:], " ")})
			} else {
				env.Set(name, &object.String{Value: ""})
			}
		} else if i < len(words) {
			env.Set(name, &object.String{Value: words[i]})
		} else {
			env.Set(name, &object.String{Value: ""})
		}
	}

	env.ShellSet("LAST_STATUS", &object.Integer{Value: 0})
	return nativeBoolToBooleanObject(true)
}

// evalUnsetCommand removes variables from the environment.
// Usage: unset @var1 @var2 ...
func evalUnsetCommand(args []string, env *Environment) object.Object {
	if len(args) == 0 {
		fmt.Fprintln(env.Stderr, "unset: usage: unset @name...")
		env.ShellSet("LAST_STATUS", &object.Integer{Value: 1})
		return nativeBoolToBooleanObject(false)
	}

	exitCode := int64(0)
	for _, a := range args {
		name := strings.TrimPrefix(a, "@")
		if errMsg := env.Delete(name); errMsg != "" {
			fmt.Fprintf(env.Stderr, "unset: %s\n", errMsg)
			exitCode = 1
		}
	}

	env.ShellSet("LAST_STATUS", &object.Integer{Value: exitCode})
	return nativeBoolToBooleanObject(exitCode == 0)
}
