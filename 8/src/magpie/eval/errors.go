package eval

import (
	"fmt"
	"strings"
)

var (
	ERR_PREFIXOP     = "unsupported operator for prefix expression:'%s' and type: %s"
	ERR_INFIXOP      = "unsupported operator for infix expression: %s '%s' %s"
	ERR_UNKNOWNIDENT = "unknown identifier: '%s' is not defined"
	ERR_DIVIDEBYZERO = "divide by zero"
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
