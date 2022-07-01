package eval

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"magpie/ast"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
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
	TUPLE_OBJ        = "TUPLE"
	HASH_OBJ         = "HASH"
	BREAK_OBJ        = "BREAK"
	CONTINUE_OBJ     = "CONTINUE"
	FALLTHROUGH_OBJ  = "FALLTHROUGH"
	REGEX_OBJ        = "REGEX"
	GO_OBJ           = "GO_OBJ"
	GFO_OBJ          = "GFO_OBJ"
	FILE_OBJ         = "FILE"
	OS_OBJ           = "OS_OBJ"
	STRUCT_OBJ       = "STRUCT"
)

var (
	TRUE        = &Boolean{Bool: true}
	FALSE       = &Boolean{Bool: false}
	BREAK       = &Break{}
	CONTINUE    = &Continue{}
	FALLTHROUGH = &Fallthrough{}
	NIL         = &Nil{}
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

func (s *String) iter() bool { return true }
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
	Value  Object   // for old campatibility
	Values []Object // return multiple values
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string {
	//return rv.Value.Inspect()

	var out bytes.Buffer
	values := []string{}
	for _, v := range rv.Values {
		values = append(values, v.Inspect())
	}

	out.WriteString(strings.Join(values, ", "))

	return out.String()
}
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

func (a *Array) iter() bool       { return true }
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

func (h *Hash) iter() bool       { return true }
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

func NewTuple(isMulti bool) *Tuple {
	//we assume tuple has at least two members
	return &Tuple{IsMulti: isMulti, Members: []Object{NIL, NIL}}
}

type Tuple struct {
	// Used in function return values.
	// if a function returns multiple values, they will wrap the results into a tuple,
	// the flag will be set to true
	IsMulti bool
	Members []Object
}

func (t *Tuple) iter() bool { return true }

func (t *Tuple) Inspect() string {
	var out bytes.Buffer
	members := []string{}
	for _, m := range t.Members {
		if m.Type() == STRING_OBJ {
			members = append(members, "\""+m.Inspect()+"\"")
		} else {
			members = append(members, m.Inspect())
		}
	}
	out.WriteString("(")
	out.WriteString(strings.Join(members, ", "))
	if (len(t.Members)) == 1 {
		out.WriteString(",")
	}
	out.WriteString(")")

	return out.String()
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }

func (t *Tuple) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "get":
		return t.get(line, args...)
	case "empty":
		return t.empty(line, args...)
	case "len":
		return t.len(line, args...)

	}
	return newError(line, ERR_NOMETHOD, method, t.Type())
}

func (t *Tuple) len(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	return NewNumber(float64(len(t.Members)))
}

func (t *Tuple) get(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	idxObj, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "get", "*Number", args[0].Type())
	}

	val := int64(idxObj.Value)
	if val < 0 || val >= int64(len(t.Members)) {
		return newError(line, ERR_INDEX, val)
	}
	return t.Members[val]
}

func (t *Tuple) empty(line string, args ...Object) Object {
	l := len(args)
	if l != 0 {
		return newError(line, ERR_ARGUMENT, "0", l)
	}

	if len(t.Members) == 0 {
		return TRUE
	}
	return FALSE
}

func (t *Tuple) HashKey() HashKey {
	// https://en.wikipedia.org/wiki/Jenkins_hash_function
	var hash uint64 = 0
	for _, v := range t.Members {
		hashable, ok := v.(Hashable)
		if !ok {
			errStr := fmt.Sprintf(ERR_KEY, v.Type())
			panic(errStr)
		}

		h := hashable.HashKey()

		hash += h.Value
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15

	return HashKey{Type: t.Type(), Value: hash}
}

type Break struct{}

func (b *Break) Inspect() string  { return "break" }
func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, b.Type())
}

type Continue struct{}

func (c *Continue) Inspect() string  { return "continue" }
func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, c.Type())
}

type Fallthrough struct{}

func (f *Fallthrough) Inspect() string  { return "fallthrough" }
func (f *Fallthrough) Type() ObjectType { return FALLTHROUGH_OBJ }
func (f *Fallthrough) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, f.Type())
}

//Whether the Object is iterable (HASH, ARRAY, STRING, TUPLE, Some of the GoObject)
type Iterable interface {
	iter() bool
}

//This `Formatter` struct is mainly used to encapsulate golang
//`fmt` package's `Formatter` interface.
type Formatter struct {
	Obj Object
}

const (
	availFlags = "-+# 0"
)

func (ft *Formatter) Format(s fmt.State, verb rune) {
	//重新组装format
	format := make([]byte, 0, 128)
	format = append(format, '%')
	var f byte
	for i := 0; i < len(availFlags); i++ {
		if f = availFlags[i]; s.Flag(int(f)) {
			format = append(format, f)
		}
	}
	var width, prec int
	var ok bool
	if width, ok = s.Width(); ok {
		format = strconv.AppendInt(format, int64(width), 10)
	}
	if prec, ok = s.Precision(); ok {
		format = append(format, '.')
		format = strconv.AppendInt(format, int64(prec), 10)
	}
	if verb > utf8.RuneSelf {
		format = append(format, string(verb)...)
	} else {
		//Here we use '%_' to print the object's type
		if verb == '_' {
			format = append(format, byte('T'))
		} else if verb == 'd' { //%d
			//如果代码中使用到%d的形式，由于我们的数字对象中
			//存储的是浮点类型，所以会出现类似下面的内容：
			//  x=%!d(float64=12)
			//因此，这里我们把%d转换成%g
			format = append(format, byte('g'))
		} else {
			format = append(format, byte(verb))
		}
	}

	formatStr := string(format)
	if formatStr == "%T" {
		t := reflect.TypeOf(ft.Obj)
		strArr := strings.Split(t.String(), ".") //t.String() = "*eval.xxx"
		fmt.Fprintf(s, "%s", strArr[1])          //NEED CHECK for "index out of bounds?"
		return
	}

	switch obj := ft.Obj.(type) {
	case *Boolean:
		fmt.Fprintf(s, formatStr, obj.Bool)
	case *Number:
		fmt.Fprintf(s, formatStr, obj.Value)
	case *String:
		fmt.Fprintf(s, formatStr, obj.String)
	default:
		fmt.Fprintf(s, formatStr, obj.Inspect())
	}
}

type RegEx struct {
	RegExp *regexp.Regexp
	Value  string
}

func (re *RegEx) Inspect() string  { return re.Value }
func (re *RegEx) Type() ObjectType { return REGEX_OBJ }

func (re *RegEx) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "match":
		return re.match(line, args...)
	case "replace":
		return re.replace(line, args...)
	case "split":
		return re.split(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, re.Type())
}

func (re *RegEx) match(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "match", "*String", args[0].Type())
	}

	str := args[0].(*String)
	matched := re.RegExp.MatchString(str.String)
	if matched {
		return TRUE
	}
	return FALSE
}

func (re *RegEx) replace(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "replace", "*String", args[0].Type())
	}

	if args[1].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "second", "replace", "*String", args[1].Type())
	}

	str := args[0].(*String)
	repl := args[1].(*String)
	result := re.RegExp.ReplaceAllString(str.String, repl.String)
	return NewString(result)
}

func (re *RegEx) split(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	if args[0].Type() != STRING_OBJ {
		return newError(line, ERR_PARAMTYPE, "first", "split", "*String", args[0].Type())
	}

	str := args[0].(*String)
	splitResult := re.RegExp.Split(str.String, -1)

	a := &Array{}
	for i := 0; i < len(splitResult); i++ {
		a.Members = append(a.Members, NewString(splitResult[i]))
	}
	return a
}

type Struct struct {
	Scope *Scope //struct's scope
}

func (s *Struct) Inspect() string {
	var out bytes.Buffer
	out.WriteString("( ")
	for k, v := range s.Scope.store {
		out.WriteString(k)
		out.WriteString("->")
		out.WriteString(v.Inspect())
		out.WriteString(" ")
	}
	out.WriteString(" )")

	return out.String()
}

func (s *Struct) Type() ObjectType { return STRUCT_OBJ }
func (s *Struct) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	var fn2 Object
	var fn *Function
	var ok bool

	if fn2, ok = s.Scope.Get(method); !ok {
		return newError(line, ERR_NOMETHOD, method, s.Type())
	}

	fn = fn2.(*Function)
	extendedScope := extendFunctionScope(fn, args)
	extendedScope.Set("self", s)
	obj := Eval(fn.Literal.Body, extendedScope)
	return unwrapReturnValue(obj)
}

func initGlobalObj() {
	//Predefine `stdin`, `stdout`, `stderr`
	SetGlobalObj("stdin", &FileObject{File: os.Stdin})
	SetGlobalObj("stdout", &FileObject{File: os.Stdout})
	SetGlobalObj("stderr", &FileObject{File: os.Stderr})
}

func init() {
	initGlobalObj()

	NewOsObj()
}
