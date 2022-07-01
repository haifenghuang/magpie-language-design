package eval

import (
	"bytes"
	"fmt"
	"magpie/ast"
	"strings"
)

//object
type ObjectType string

const (
	NUMBER_OBJ       = "NUMBER"
	NIL_OBJ          = "NIL_OBJ"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	ERROR_OBJ        = "ERROR"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
)

var (
	TRUE  = &Boolean{Bool: true}
	FALSE = &Boolean{Bool: false}
	NIL   = &Nil{}
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Number struct {
	Value float64
}

func (n *Number) Inspect() string {
	return fmt.Sprintf("%g", n.Value)
}

func (n *Number) Type() ObjectType { return NUMBER_OBJ }

func NewNumber(f float64) *Number {
	return &Number{Value: f}
}

type Nil struct {
}

func (n *Nil) Inspect() string {
	return "nil"
}
func (n *Nil) Type() ObjectType { return NIL_OBJ }

func NewNil(s string) *Nil {
	return &Nil{}
}

func NewBooleanObj(b bool) *Boolean {
	return &Boolean{Bool: b}
}

type Boolean struct {
	Bool bool
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%v", b.Bool)
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

type String struct {
	String string
}

func (s *String) Inspect() string {
	return s.String
}

func (s *String) Type() ObjectType { return STRING_OBJ }

func NewString(s string) *String {
	return &String{String: s}
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Function struct {
	Literal *ast.FunctionLiteral
	Scope   *Scope
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	return f.Literal.String()
}

type Array struct {
	Members []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	members := []string{}
	for _, e := range a.Members {
		if e.Type() == STRING_OBJ {
			members = append(members, "\""+e.Inspect()+"\"")
		} else {
			members = append(members, e.Inspect())
		}
	}

	out.WriteString("[")
	out.WriteString(strings.Join(members, ", "))
	out.WriteString("]")
	return out.String()
}
