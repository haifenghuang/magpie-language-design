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

func (b *Builtin) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, b.Type())
}

var builtins map[string]*Builtin

func printBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			resultStr := ""
			for _, arg := range args {
				resultStr = resultStr + arg.Inspect()
			}
			fmt.Fprint(scope.Writer, resultStr)
			return NIL
		},
	}
}

func printlnBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				fmt.Fprintln(scope.Writer)
			}

			resultStr := ""
			for _, arg := range args {
				resultStr = resultStr + arg.Inspect()
			}
			fmt.Fprintln(scope.Writer, resultStr)
			return NIL
		},
	}
}

func printfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) < 1 {
				return newError(line, ERR_ARGUMENT, ">0", len(args))
			}

			formatObj, ok := args[0].(*String)
			if !ok {
				return newError(line, ERR_PARAMTYPE, "first", "printf", "*String", args[0].Type())
			}

			subArgs := args[1:]
			wrapped := make([]interface{}, len(subArgs))
			for i, v := range subArgs {
				wrapped[i] = &Formatter{Obj: v}
			}

			formatStr := formatObj.String
			fmt.Fprintf(scope.Writer, formatStr, wrapped...)

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
			case *Array:
				return NewNumber(float64(len(arg.Members)))
			case *Tuple:
				return NewNumber(float64(len(arg.Members)))
			case *Hash:
				return NewNumber(float64(len(arg.Pairs)))
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
		"printf":  printfBuiltin(),
		"say":     printlnBuiltin(),
		"len":     lenBuiltin(),
	}
}
