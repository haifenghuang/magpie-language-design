package eval

import (
	_ "fmt"
	"os"
)

const (
	os_name = "os"
)

type Os struct{}

func NewOsObj() Object {
	ret := &Os{}
	SetGlobalObj(os_name, ret)

	return ret
}

func (o *Os) Inspect() string  { return "<" + os_name + ">" }
func (o *Os) Type() ObjectType { return OS_OBJ }

func (o *Os) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "getenv":
		return o.getenv(line, args...)
	case "setenv":
		return o.setenv(line, args...)
	case "chdir":
		return o.chdir(line, args...)
	case "mkdir":
		return o.mkdir(line, args...)
	case "exit":
		return o.exit(line, args...)
	}
	return newError(line, ERR_NOMETHOD, method, o.Type())
}

func (o *Os) getenv(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "getenv", "*String", args[0].Type())
	}

	ret := os.Getenv(key.String)
	return NewString(ret)
}

func (o *Os) setenv(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	key, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "setenv", "*String", args[0].Type())
	}

	value, ok := args[1].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "second", "setenv", "*String", args[1].Type())
	}

	err := os.Setenv(key.String, value.String)
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) chdir(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	newDir, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "chdir", "*String", args[0].Type())
	}

	err := os.Chdir(newDir.String)
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) mkdir(line string, args ...Object) Object {
	if len(args) != 2 {
		return newError(line, ERR_ARGUMENT, "2", len(args))
	}

	name, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "mkdir", "*String", args[0].Type())
	}

	perm, ok := args[1].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "second", "mkdir", "*Number", args[1].Type())
	}

	err := os.Mkdir(name.String, os.FileMode(int64(perm.Value)))
	if err != nil {
		return FALSE
	}
	return TRUE
}

func (o *Os) exit(line string, args ...Object) Object {
	if len(args) != 0 && len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "0|1", len(args))
	}

	if len(args) == 0 {
		os.Exit(0)
		return NIL
	}

	code, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "exit", "*Number", args[0].Type())
	}

	os.Exit(int(code.Value))

	return NIL
}
