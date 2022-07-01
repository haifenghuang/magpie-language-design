package eval

import (
	"fmt"
	"strings"
)

var (
	ERR_ARGUMENT     = "wrong number of arguments. expected=%d, got=%d"
	ERR_NOMETHOD     = "undefined method '%s' for object %s"
	ERR_INDEX        = "index error: '%d' out of range"
	ERR_KEY          = "key error: type %s is not hashable"
	ERR_PREFIXOP     = "unsupported operator for prefix expression:'%s' and type: %s"
	ERR_INFIXOP      = "unsupported operator for infix expression: %s '%s' %s"
	ERR_POSTFIXOP    = "unsupported operator for postfix expression:'%s' and type: %s"
	ERR_UNKNOWNIDENT = "unknown identifier: '%s' is not defined"
	ERR_DIVIDEBYZERO = "divide by zero"
	ERR_NOTFUNCTION  = "not a function: %s"
	ERR_PARAMTYPE    = "%s argument for '%s' should be type %s. got=%s"
	ERR_NOTITERABLE  = "foreach's operating type must be iterable"
	ERR_NOINDEXABLE  = "index error: type %s is not indexable"
)

func newError(line string, format string, args ...interface{}) *Error {
	msg := "Runtime Error at " + strings.TrimLeft(line, " \t") + "\n\t" + fmt.Sprintf(format, args...) + "\n"
	return &Error{Message: msg}
}

type Error struct {
	Message string
}

func (e *Error) Inspect() string  { return e.Message }
func (e *Error) Type() ObjectType { return ERROR_OBJ }

func isError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR_OBJ
	}
	return false
}

func (e *Error) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, "%s", e.Message)
}
