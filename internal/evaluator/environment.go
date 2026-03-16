package evaluator

import (
	"dush/internal/evaluator/object"
	"io"
	"os"
)

func NewEnvironment() *Environment {
	s := make(map[string]object.Object)
	e := make(map[string]string)
	return &Environment{store: s, EnvOverrides: e, outer: nil, Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.Stdin = outer.Stdin
	env.Stdout = outer.Stdout
	env.Stderr = outer.Stderr
	return env
}

type Environment struct {
	store        map[string]object.Object
	EnvOverrides map[string]string // String variables tailored for OS environment injection
	outer        *Environment
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
}

func (e *Environment) Get(name string) (object.Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val object.Object) object.Object {
	e.store[name] = val
	return val
}

// Update sets a variable in the scope where it was originally defined.
// If the variable doesn't exist in any scope, it sets it in the current scope.
func (e *Environment) Update(name string, val object.Object) object.Object {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return val
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	// Variable not found in any scope, set in current
	e.store[name] = val
	return val
}

func (e *Environment) GetAllOverrides() map[string]string {
	overrides := make(map[string]string)
	
	// Collect from outer first, so inner overrides take precedence
	if e.outer != nil {
		outerOverrides := e.outer.GetAllOverrides()
		for k, v := range outerOverrides {
			overrides[k] = v
		}
	}
	
	for k, v := range e.EnvOverrides {
		overrides[k] = v
	}
	
	return overrides
}
