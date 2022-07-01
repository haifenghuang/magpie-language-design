package eval

import (
	"fmt"
	"unicode/utf8"
)

type BuiltinFunc func(line string, scope *Scope, args ...Object) Object

type Builtin struct {
	Fn BuiltinFunc
}

func (b *Builtin) Inspect() string  { return "<builtin function>" }
func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

var builtins map[string]*Builtin

func printBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			resultStr := ""
			for _, arg := range args {
				resultStr = resultStr + arg.Inspect()
			}
			fmt.Print(resultStr)
			return NIL
		},
	}
}

func printlnBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				fmt.Println()
			}

			resultStr := ""
			for _, arg := range args {
				resultStr = resultStr + arg.Inspect() + "\n"
			}
			fmt.Print(resultStr)
			return NIL
		},
	}
}

func lenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return newError(line, ERR_ARGUMENT, 1, len(args))
			}

			switch arg := args[0].(type) {
			case *String:
				n := utf8.RuneCountInString(arg.String)
				return NewNumber(float64(n))
			default:
				return newError(line, "argument to `len` not supported, got %s", args[0].Type())
			}
		},
	}
}

func init() {
	builtins = map[string]*Builtin{
		"print":   printBuiltin(),
		"println": printlnBuiltin(),
		"say":     printlnBuiltin(),
		"len":     lenBuiltin(),
	}
}
