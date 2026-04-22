package evaluator

import (
	"dush/internal/evaluator/object"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
)

// ShellVersion can be set by the main package at startup.
var ShellVersion = "dev"

// Variable wraps an object value with metadata flags.
type Variable struct {
	Value    object.Object
	Exported bool // pub: visible to child processes
	Const    bool // const: immutable by user code
}

type Environment struct {
	store     map[string]*Variable
	shellVars map[string]bool // variables the shell runtime is allowed to update
	outer     *Environment
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
}

func NewEnvironment() *Environment {
	s := make(map[string]*Variable)
	sv := make(map[string]bool)
	env := &Environment{
		store:     s,
		shellVars: sv,
		outer:     nil,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}

	// Populate shell variables (read-only for user code)
	env.setShellVar("LAST_STATUS", &object.Integer{Value: 0})
	env.setShellVar("SHELL_PID", &object.Integer{Value: int64(os.Getpid())})

	homeDir, _ := os.UserHomeDir()
	env.setShellVar("HOME_DIR", &object.String{Value: homeDir})

	cwd, _ := os.Getwd()
	env.setShellVar("WORKING_DIR", &object.String{Value: cwd})

	env.setShellVar("OS_NAME", &object.String{Value: runtime.GOOS})

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	env.setShellVar("USER_NAME", &object.String{Value: username})

	env.setShellVar("SHELL_VERSION", &object.String{Value: ShellVersion})
	env.setShellVar("ARGS", &object.Array{Elements: []object.Object{}})
	env.setShellVar("SCRIPT", &object.String{Value: ""})

	return env
}

// SetScriptArgs populates @SCRIPT (path) and @ARGS (user arguments) for a
// script invocation. Safe to call multiple times; later calls overwrite.
func (e *Environment) SetScriptArgs(scriptPath string, args []string) {
	elems := make([]object.Object, 0, len(args))
	for _, a := range args {
		elems = append(elems, &object.String{Value: a})
	}
	e.ShellSet("SCRIPT", &object.String{Value: scriptPath})
	e.ShellSet("ARGS", &object.Array{Elements: elems})
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := &Environment{
		store:     make(map[string]*Variable),
		shellVars: make(map[string]bool),
		outer:     outer,
		Stdin:     outer.Stdin,
		Stdout:    outer.Stdout,
		Stderr:    outer.Stderr,
	}
	return env
}

// setShellVar sets a variable that is managed by the shell runtime.
// These are const from the user's perspective but updatable by the shell.
func (e *Environment) setShellVar(name string, val object.Object) {
	e.store[name] = &Variable{Value: val, Const: true}
	e.shellVars[name] = true
}

// ShellSet updates a shell-managed variable, bypassing const checks.
// Used for LAST_STATUS, WORKING_DIR, etc.
func (e *Environment) ShellSet(name string, val object.Object) {
	// Walk up to find where the shell var is defined
	if v, ok := e.store[name]; ok {
		if e.shellVars[name] {
			v.Value = val
			return
		}
	}
	if e.outer != nil {
		e.outer.ShellSet(name, val)
		return
	}
	// If not found anywhere, set in current scope (shouldn't happen for shell vars)
	e.store[name] = &Variable{Value: val, Const: true}
	e.shellVars[name] = true
}

func (e *Environment) Get(name string) (object.Object, bool) {
	v, ok := e.store[name]
	if ok {
		return v.Value, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	// Fallback to OS environment
	if val, exists := os.LookupEnv(name); exists {
		return &object.String{Value: val}, true
	}
	return nil, false
}

// Set creates or overwrites a variable in the current scope.
// Returns an error string if the variable is const, or empty string on success.
func (e *Environment) Set(name string, val object.Object) string {
	if v, ok := e.store[name]; ok {
		if v.Const {
			return fmt.Sprintf("cannot assign to const variable '%s'", name)
		}
		v.Value = val
		return ""
	}
	e.store[name] = &Variable{Value: val}
	return ""
}

// Update sets a variable in the scope where it was originally defined.
// Returns an error string if the variable is const.
func (e *Environment) Update(name string, val object.Object) string {
	if v, ok := e.store[name]; ok {
		if v.Const {
			return fmt.Sprintf("cannot assign to const variable '%s'", name)
		}
		v.Value = val
		return ""
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	// Variable not found in any scope, create in current
	e.store[name] = &Variable{Value: val}
	return ""
}

// SetConst creates an immutable variable in the current scope.
func (e *Environment) SetConst(name string, val object.Object) string {
	if v, ok := e.store[name]; ok {
		if v.Const {
			return fmt.Sprintf("cannot reassign const variable '%s'", name)
		}
	}
	e.store[name] = &Variable{Value: val, Const: true}
	return ""
}

// Delete removes a variable from the environment.
// Returns an error string if the variable is const/shell-managed, or empty on success.
func (e *Environment) Delete(name string) string {
	if v, ok := e.store[name]; ok {
		if v.Const {
			return fmt.Sprintf("cannot unset const variable '%s'", name)
		}
		if e.shellVars[name] {
			return fmt.Sprintf("cannot unset shell variable '%s'", name)
		}
		delete(e.store, name)
		return ""
	}
	if e.outer != nil {
		return e.outer.Delete(name)
	}
	return ""
}

// SetPub creates an exported variable in the current scope.
func (e *Environment) SetPub(name string, val object.Object, isConst bool) string {
	if v, ok := e.store[name]; ok {
		if v.Const {
			return fmt.Sprintf("cannot reassign const variable '%s'", name)
		}
	}
	e.store[name] = &Variable{Value: val, Exported: true, Const: isConst}
	return ""
}

// MarkPub marks an existing variable as exported.
func (e *Environment) MarkPub(name string) string {
	if v, ok := e.store[name]; ok {
		v.Exported = true
		return ""
	}
	if e.outer != nil {
		return e.outer.MarkPub(name)
	}
	return fmt.Sprintf("undefined variable '%s'", name)
}

// GetExportedVars collects all exported variables for child process env injection.
func (e *Environment) GetExportedVars() map[string]string {
	exports := make(map[string]string)

	// Collect from outer first, so inner values take precedence
	if e.outer != nil {
		outerExports := e.outer.GetExportedVars()
		for k, v := range outerExports {
			exports[k] = v
		}
	}

	for name, v := range e.store {
		if v.Exported {
			exports[name] = objectToStringEnv(v.Value)
		}
	}

	return exports
}

func objectToStringEnv(obj object.Object) string {
	switch obj := obj.(type) {
	case *object.String:
		return obj.Value
	case *object.Integer:
		return strconv.FormatInt(obj.Value, 10)
	case *object.Float:
		return strconv.FormatFloat(obj.Value, 'g', -1, 64)
	case *object.Boolean:
		return strconv.FormatBool(obj.Value)
	default:
		return obj.Inspect()
	}
}
