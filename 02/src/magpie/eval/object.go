package eval

import (
	"fmt"
)

//object
type ObjectType string

const (
	NUMBER_OBJ = "NUMBER"
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
