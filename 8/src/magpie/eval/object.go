package eval

import (
	"fmt"
)

//object
type ObjectType string

const (
	NUMBER_OBJ       = "NUMBER"
	NIL_OBJ          = "NIL_OBJ"
	BOOLEAN_OBJ      = "BOOLEAN"
	ERROR_OBJ        = "ERROR"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
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

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
