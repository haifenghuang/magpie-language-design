package main

import (
	"fmt"
	"magpie/eval"
	"magpie/lexer"
	"magpie/parser"
	"magpie/token"
	"os"
)

func TestLexer() {
	input := " let x = 2 + (3 * 4) / ( 5 - 3 ) + 10 - a ** 2;"
	fmt.Printf("Input = %s\n", input)

	l := lexer.NewLexer(input)
	for {
		tok := l.NextToken()
		fmt.Printf("Type: %s, Literal = %s\n", tok.Type, tok.Literal)
		if tok.Type == token.TOKEN_EOF {
			break
		}
	}
}

func TestParser() {
	input := " 2 + (3 * 4) / ( 5 - 3 ) + 10 + 2 ** 2 ** 3 + xyz"
	expected := "((((2 + ((3 * 4) / (5 - 3))) + 10) + (2 ** (2 ** 3))) + xyz)"
	l := lexer.NewLexer(input)
	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	if program.String() != expected {
		fmt.Printf("Syntax error: expected %s, got %s\n", expected, program.String())
		os.Exit(1)
	}

	fmt.Printf("input  = %s\n", input)
	fmt.Printf("output = %s\n", program.String())
}

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"-1 - 2.333", "-3.333"},
		{"1 + 2", "3"},
		{"2 + (3 * 4) / ( 6 - 3 ) + 10", "16"},
		{"2 + 3 * 4 / 6 - 3  + 10", "11"},
		{"(5 + 2) * (4 - 2) + 6", "20"},
		{"5 + 2 * 4 - 2 + 6", "17"},
		{"5 + 2.1 * 4 - 2 + 6.2", "17.6"},
		{"2 + 2 ** 2 ** 3", "258"},
		{"10", "10"},
		{"nil", "nil"},
		{"true", "true"},
		{"false", "false"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, err := range p.Errors() {
				fmt.Println(err)
			}
			break
		}

		evaluated := eval.Eval(program)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("%s", evaluated.Inspect())
			} else {
				fmt.Printf("%s = %s\n", tt.input, tt.expected)
			}
		}
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 1 {
		if args[0] == "--lexer" {
			TestLexer()
		} else if args[0] == "--parser" {
			TestParser()
		}
		os.Exit(0)
	}
	TestEval()
}
