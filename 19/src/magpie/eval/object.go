package eval

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"magpie/ast"
	"math"
	"strconv"
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
	HASH_OBJ         = "HASH"
)

var (
	TRUE  = &Boolean{Bool: true}
	FALSE = &Boolean{Bool: false}
	NIL   = &Nil{}
)

type Object interface {
	Type() ObjectType
	Inspect() string
	CallMethod(line string, scope *Scope, method string, args ...Object) Object
}

type Hashable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Number struct {
	Value float64
}

func (n *Number) Inspect() string {
	return fmt.Sprintf("%g", n.Value)
}

func (n *Number) Type() ObjectType { return NUMBER_OBJ }

func (n *Number) HashKey() HashKey {
	return HashKey{Type: n.Type(), Value: uint64(n.Value)}
}

func (n *Number) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "ceil":
		return n.ceil(line, args...)
	case "floor":
		return n.floor(line, args...)
	case "trunc":
		return n.trunc(line, args...)
	case "sqrt":
		return n.sqrt(line, args...)
	case "pow":
		return n.pow(line, args...)
	case "round":
		return n.round(line, args...)
	case "str":
		return n.str(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, n.Type())
}

func (n *Number) ceil(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewNumber(math.Ceil(n.Value))
}

func (n *Number) floor(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewNumber(math.Floor(n.Value))

}

func (n *Number) trunc(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewNumber(math.Trunc(n.Value))
}

func (n *Number) sqrt(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewNumber(math.Sqrt(n.Value))
}

func (n *Number) pow(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	temp := args[0].(*Number).Value
	return NewNumber(math.Pow(n.Value, temp))
}

func (n *Number) round(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	precision := int64(args[0].(*Number).Value)

	format := fmt.Sprintf("%%.%df", precision)    //'%.xf', x is the precision, e.g. %.2f
	resultStr := fmt.Sprintf(format, n.Value)     //convert to string
	ret, err := strconv.ParseFloat(resultStr, 64) //convert string back to float
	if err != nil {
		return NewNumber(math.NaN())
	}
	return NewNumber(ret)
}

func (n *Number) str(line string, args ...Object) Object {
	argLen := len(args)
	if argLen != 0 {
		return newError(line, ERR_ARGUMENT, "0", argLen)
	}

	return NewString(fmt.Sprintf("%g", n.Value))
}

func NewNumber(f float64) *Number {
	return &Number{Value: f}
}

type Nil struct {
}

func (n *Nil) Inspect() string {
	return "nil"
}
func (n *Nil) Type() ObjectType { return NIL_OBJ }

func (n *Nil) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, n.Type())
}

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

func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Bool {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

func (b *Boolean) CallMethod(line string, scope *Scope, method string, args ...Object) Object {

	switch method {
	case "toYesNo":
		return b.toYesNo(line, args...)
	case "toTrueFalse":
		return b.toTrueFalse(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, b.Type())
}

func (b *Boolean) toYesNo(line string, args ...Object) Object {
	if b.Bool {
		return NewString("yes")
	}
	return NewString("no")
}

func (b *Boolean) toTrueFalse(line string, args ...Object) Object {
	if b.Bool {
		return NewString("true")
	}
	return NewString("false")
}

type String struct {
	String string
}

func (s *String) Inspect() string {
	return s.String
}

func (s *String) Type() ObjectType { return STRING_OBJ }

func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.String))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

func (s *String) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "lower":
		return s.lower(line, args...)
	case "upper":
		return s.upper(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, s.Type())
}

func (s *String) lower(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if s.String == "" {
		return s
	}

	str := strings.ToLower(s.String)
	return NewString(str)
}

func (s *String) upper(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if s.String == "" {
		return s
	}

	ret := strings.ToUpper(s.String)
	return NewString(ret)
}

func NewString(s string) *String {
	return &String{String: s}
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, rv.Type())
}

type Function struct {
	Literal *ast.FunctionLiteral
	Scope   *Scope
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	return f.Literal.String()
}
func (f *Function) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, f.Type())
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

func (a *Array) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "len":
		return a.len(line, args...)
	case "push":
		return a.push(line, args...)
	case "pop":
		return a.pop(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, a.Type())
}

func (a *Array) len(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	return NewNumber(float64(len(a.Members)))
}

func (a *Array) pop(line string, args ...Object) Object {
	last := len(a.Members) - 1
	if len(args) == 0 {
		if last < 0 {
			return newError(line, ERR_INDEX, last)
		}
		popped := a.Members[last]
		a.Members = a.Members[:last]
		return popped
	}
	idx := int64(args[0].(*Number).Value)
	if idx < 0 {
		idx = idx + int64(last+1)
	}
	if idx < 0 || idx > int64(last) {
		return newError(line, ERR_INDEX, idx)
	}
	popped := a.Members[idx]
	a.Members = append(a.Members[:idx], a.Members[idx+1:]...)
	return popped
}

func (a *Array) push(line string, args ...Object) Object {
	l := len(args)
	if l != 1 {
		return newError(line, ERR_ARGUMENT, "1", l)
	}
	a.Members = append(a.Members, args[0])
	return a
}

type HashPair struct {
	Key   Object
	Value Object
}

func NewHash() *Hash {
	return &Hash{Pairs: make(map[HashKey]HashPair)}
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, pair := range h.Pairs {
		var key, val string
		if pair.Key.Type() == STRING_OBJ {
			key = "\"" + pair.Key.Inspect() + "\""
		} else {
			key = pair.Key.Inspect()
		}

		if pair.Value.Type() == STRING_OBJ {
			val = "\"" + pair.Value.Inspect() + "\""
		} else {
			val = pair.Value.Inspect()
		}

		pairs = append(pairs, fmt.Sprintf("%s:%s", key, val))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

func (h *Hash) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "keys":
		return h.keys(line, args...)
	case "values":
		return h.values(line, args...)
	case "pop", "delete", "remove":
		return h.pop(line, args...)
	case "push", "set":
		return h.push(line, args...)
	}

	return newError(line, ERR_NOMETHOD, method, h.Type())
}

func (h *Hash) keys(line string, args ...Object) Object {
	keys := &Array{}
	for _, pair := range h.Pairs {
		keys.Members = append(keys.Members, pair.Key)
	}

	return keys
}

func (h *Hash) values(line string, args ...Object) Object {
	values := &Array{}
	for _, pair := range h.Pairs {
		values.Members = append(values.Members, pair.Value)
	}

	return values
}

func (h *Hash) pop(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}
	hashable, ok := args[0].(Hashable)
	if !ok {
		return newError(line, ERR_KEY, args[0].Type())
	}
	if hashPair, ok := h.Pairs[hashable.HashKey()]; ok {
		delete(h.Pairs, hashable.HashKey())
		return hashPair.Value
	}

	return NIL
}

func (h *Hash) push(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}
	if hashable, ok := args[0].(Hashable); ok {
		h.Pairs[hashable.HashKey()] = HashPair{Key: args[0], Value: args[1]}
	} else {
		return newError(line, ERR_KEY, args[0].Type())
	}

	return h
}
