package object

import (
	"bytes"
	"fmt"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	MAP_OBJ          = "MAP"
	BREAK_OBJ        = "BREAK"
	CONTINUE_OBJ     = "CONTINUE"
)

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

type Float struct {
	Value float64
}

func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }
func (f *Float) Type() ObjectType { return FLOAT_OBJ }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return s.Value }
func (s *String) Type() ObjectType { return STRING_OBJ }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

type Error struct {
	Message string
}

func (e *Error) Inspect() string  { return "ERROR: " + e.Message }
func (e *Error) Type() ObjectType { return ERROR_OBJ }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

// HashKey is used as a map key for the Map object.
type HashKey struct {
	Type  ObjectType
	Value string // string representation for hashing
}

func HashKeyFromObject(obj Object) (HashKey, bool) {
	switch o := obj.(type) {
	case *String:
		return HashKey{Type: STRING_OBJ, Value: o.Value}, true
	case *Integer:
		return HashKey{Type: INTEGER_OBJ, Value: fmt.Sprintf("%d", o.Value)}, true
	case *Boolean:
		return HashKey{Type: BOOLEAN_OBJ, Value: fmt.Sprintf("%t", o.Value)}, true
	default:
		return HashKey{}, false
	}
}

// MapPair holds a key-value pair in a Map.
type MapPair struct {
	Key   Object
	Value Object
}

// Map represents a hash map / dictionary.
type Map struct {
	Pairs map[HashKey]MapPair
	Order []HashKey // preserves insertion order
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, k := range m.Order {
		pair := m.Pairs[k]
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// FuncParam represents a function parameter with a name.
// Used to avoid circular imports between object and ast packages.
type FuncParam interface {
	ParamName() string
}

// Function represents a user-defined function.
// Parameters is interface{} to avoid circular imports with ast package.
// The evaluator casts it to the appropriate type.
type Function struct {
	Parameters interface{} // []*ast.VarExpression — avoid circular import
	Body       interface{} // *ast.BlockStatement — avoid circular import
	Env        interface{} // *evaluator.Environment — avoid circular import
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	return "proc(...) { ... }"
}
