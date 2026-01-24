package evaluator

import (
	"dush/internal/evaluator/object"
	"io"
	"os"
)

func NewEnvironment() *Environment {
	s := make(map[string]object.Object)
	return &Environment{store: s, outer: nil, Stdout: os.Stdout, Stderr: os.Stderr}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	env.Stdout = outer.Stdout
	env.Stderr = outer.Stderr
	return env
}

type Environment struct {
	store  map[string]object.Object
	outer  *Environment
	Stdout io.Writer
	Stderr io.Writer
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
