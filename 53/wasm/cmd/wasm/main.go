package main

import (
	"bytes"
	_ "fmt"
	_ "strings"
	"syscall/js"

	"magpie/eval"
	"magpie/lexer"
	"magpie/parser"
)

func runCode(this js.Value, i []js.Value) interface{} {
	m := make(map[string]interface{})
	var buf bytes.Buffer

	l := lexer.NewLexer(i[0].String())
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			buf.WriteString(msg + "\n")
		}

		m["output"] = buf.String()
		return m
	}

	scope := eval.NewScope(nil, &buf)
	result := eval.Eval(program, scope)
	if (string(result.Type()) == eval.ERROR_OBJ) {
		m["output"] = buf.String() + result.Inspect()
	} else {
		m["output"] = buf.String()
	}

	return m
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("magpie_run_code", js.FuncOf(runCode))
	<-c
}
