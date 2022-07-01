package eval

import (
	"fmt"
	"strings"
)

var (
	ERR_ARGUMENT        = "wrong number of arguments. expected=%d, got=%d"
	ERR_NOMETHOD        = "undefined method '%s' for object %s"
	ERR_NOMETHODEX      = "undefined method '%s.%s', Did you mean '%s.%s'?"
	ERR_INDEX           = "index error: '%d' out of range"
	ERR_KEY             = "key error: type %s is not hashable"
	ERR_PREFIXOP        = "unsupported operator for prefix expression:'%s' and type: %s"
	ERR_INFIXOP         = "unsupported operator for infix expression: %s '%s' %s"
	ERR_POSTFIXOP       = "unsupported operator for postfix expression:'%s' and type: %s"
	ERR_UNKNOWNIDENT    = "unknown identifier: '%s' is not defined"
	ERR_DIVIDEBYZERO    = "divide by zero"
	ERR_NOTFUNCTION     = "expect a function, got %s"
	ERR_PARAMTYPE       = "%s argument for '%s' should be type %s. got=%s"
	ERR_NOTITERABLE     = "foreach's operating type must be iterable"
	ERR_IMPORT          = "import error: %s"
	ERR_NAMENOTEXPORTED = "cannot refer to unexported name %s.%s"
	ERR_INVALIDARG      = "invalid argument supplied"
	ERR_NOINDEXABLE     = "index error: type %s is not indexable"
	ERR_NOTREGEXP       = "right type is not a regexp object, got %s"
	ERR_NOCONSTRUCTOR   = "got %d parameters, but the struct has no 'init' method supplied"
	ERR_THROWNOTHANDLED = "throw object '%s' not handled"
	ERR_RANGETYPE       = "range(..) type should be %s type, got %s"
	ERR_MULTIASSIGN     = "the number of names and values are not equal"
	ERR_DECORATOR       = "decorator '%s' is not a function"
	ERR_DECORATED_NAME  = "can not find the name of the decorated function"
	ERR_DECORATOR_FN    = "a decorator must decorate a named function or another decorator"
	ERR_PIPE            = "pipe operator's right hand side is not a function"
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
