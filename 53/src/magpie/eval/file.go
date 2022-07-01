package eval

import (
	"bufio"
	_ "fmt"
	"io"
	"os"
)

type FileObject struct {
	File    *os.File
	Name    string
	Scanner *bufio.Scanner
}

func (f *FileObject) Inspect() string  { return "<file object: " + f.Name + ">" }
func (f *FileObject) Type() ObjectType { return FILE_OBJ }
func (f *FileObject) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	switch method {
	case "close":
		return f.close(line, args...)
	case "read":
		return f.read(line, args...)
	case "readLine":
		return f.readLine(line, args...)
	case "write":
		return f.write(line, args...)
	case "writeString":
		return f.writeString(line, args...)
	case "writeLine":
		return f.writeLine(line, args...)
	case "name":
		return f.getName(line, args...)
	default:
		return newError(line, ERR_NOMETHOD, method, f.Type())
	}
}

func (f *FileObject) close(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	err := f.File.Close()
	if err != nil {
		return newError(line, "'close' failed. reason: %s", err.Error())
	}
	return TRUE
}

//Note: This method will return three different values:
//   1. nil    - with error message    (ERROR)
//   2. nil    - without error message (EOF)
//   3. string - read string
func (f *FileObject) read(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	readlen, ok := args[0].(*Number)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "read", "*Number", args[0].Type())
	}

	buffer := make([]byte, int(readlen.Value))
	n, err := f.File.Read(buffer)
	if err != io.EOF && err != nil {
		return newError(line, "'read' failed. reason: %s", err.Error())
	}

	if n == 0 && err == io.EOF {
		return NIL
	}
	return NewString(string(buffer))
}

func (f *FileObject) readLine(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}
	if f.Scanner == nil {
		f.Scanner = bufio.NewScanner(f.File)
		f.Scanner.Split(bufio.ScanLines)
	}
	aLine := f.Scanner.Scan()
	if err := f.Scanner.Err(); err != nil {
		return newError(line, "'readline' failed. reason: %s", err.Error())
	}
	if !aLine {
		return NIL
	}
	return NewString(f.Scanner.Text())
}

func (f *FileObject) write(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "write", "*String", args[0].Type())
	}

	n, err := f.File.Write([]byte(content.String))
	if err != nil {
		return newError(line, "'write' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(n))
}

func (f *FileObject) writeString(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "writeString", "*String", args[0].Type())
	}

	ret, err := f.File.WriteString(content.String)
	if err != nil {
		return newError(line, "'writeString' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(ret))
}

func (f *FileObject) writeLine(line string, args ...Object) Object {
	if len(args) != 1 {
		return newError(line, ERR_ARGUMENT, "1", len(args))
	}

	content, ok := args[0].(*String)
	if !ok {
		return newError(line, ERR_PARAMTYPE, "first", "writeLine", "*String", args[0].Type())
	}

	ret, err := f.File.Write([]byte(content.String + "\n"))
	if err != nil {
		return newError(line, "'writeLine' failed. reason: %s", err.Error())
	}

	return NewNumber(float64(ret))
}

func (f *FileObject) getName(line string, args ...Object) Object {
	if len(args) != 0 {
		return newError(line, ERR_ARGUMENT, "0", len(args))
	}

	return NewString(f.File.Name())
}
