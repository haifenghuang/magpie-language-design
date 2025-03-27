package eval

import (
	"fmt"
	"os"
	"unicode/utf8"
)

var fileModeTable = map[string]int{
	"r":   os.O_RDONLY,
	"<":   os.O_RDONLY,
	"w":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
	">":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
	"a":   os.O_APPEND | os.O_CREATE,
	">>":  os.O_APPEND | os.O_CREATE,
	"r+":  os.O_RDWR,
	"+<":  os.O_RDWR,
	"w+":  os.O_RDWR | os.O_CREATE | os.O_TRUNC,
	"+>":  os.O_RDWR | os.O_CREATE | os.O_TRUNC,
	"a+":  os.O_RDWR | os.O_APPEND | os.O_CREATE,
	"+>>": os.O_RDWR | os.O_APPEND | os.O_CREATE,
}

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

func openBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			var fname *String
			var flag int = os.O_RDONLY
			var ok bool
			var perm os.FileMode = os.FileMode(0666)

			tup := NewTuple(true)

			argLen := len(args)
			if argLen < 1 {
				tup.Members[1] = newError(line, ERR_ARGUMENT, "at least one", argLen)
				return tup
			}

			fname, ok = args[0].(*String)
			if !ok {
				tup.Members[1] = newError(line, ERR_PARAMTYPE, "first", "open", "*String", args[0].Type())
				return tup
			}

			if argLen == 2 {
				m, ok := args[1].(*String)
				if !ok {
					tup.Members[1] = newError(line, ERR_PARAMTYPE, "second", "open", "*String", args[1].Type())
					return tup
				}

				flag, ok = fileModeTable[m.String]
				if !ok {
					tup.Members[1] = newError(line, "unknown file mode supplied")
					return tup
				}
			}

			if len(args) == 3 {
				p, ok := args[2].(*Number)
				if !ok {
					tup.Members[1] = newError(line, ERR_PARAMTYPE, "third", "open", "*Integer", args[2].Type())
					return tup
				}

				perm = os.FileMode(int(p.Value))
			}

			f, err := os.OpenFile(fname.String, flag, perm)
			if err != nil {
				tup.Members[1] = newError(line, "'open' failed with error: %s", err.Error())
				return tup
			}

			tup.Members[0] = &FileObject{File: f, Name: "<file object: " + fname.String + ">"}
			return tup
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
		"open":    openBuiltin(),
		"type":    typeBuiltin(),
	}
}

func typeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return newError(line, ERR_ARGUMENT, 1, len(args))
			}

			switch args[0].(type) {
			case *Number:
				return NewString("number")
			case *Nil:
				return NewString("nil")
			case *Boolean:
				return NewString("bool")
			case *Error:
				return NewString("error")
			case *Break:
				return NewString("break")
			case *Continue:
				return NewString("continue")
			case *ReturnValue:
				return NewString("return")
			case *Function:
				return NewString("function")
			case *Builtin:
				return NewString("builtin")
			case *RegEx:
				return NewString("regex")
			case *GoObject:
				return NewString("go")
			case *GoFuncObject:
				return NewString("gofunction")
			case *FileObject:
				return NewString("file")
			case *Os:
				return NewString("os")
			case *Struct:
				return NewString("struct")
			case *Throw:
				return NewString("throw")
			case *String:
				return NewString("string")
			case *Array:
				return NewString("array")
			case *Tuple:
				return NewString("tuple")
			case *Hash:
				return NewString("hash")
			default:
				return newError(line, "argument to `type` not supported, got=%s", args[0].Type())
			}
		},
	}
}
