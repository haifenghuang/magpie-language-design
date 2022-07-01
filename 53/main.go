package main

import (
	"fmt"
	"github.com/maja42/ember"
	"magpie/eval"
	"magpie/lexer"
	"magpie/parser"
	"os"
	"runtime"
)

/*
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
*/

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
		{"let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; x", "16"},
		{"let x = 2 + (3 * 4) / ( 6 - 3 ) + 10; y", "error"},
		{`{ let x = 10 { x } }`, "10"},
		{"let x = \"hello world\"; x", "hello world"},
		{`let x = "hello " + "world"; x`, "hello world"},
		{"let add = fn(x,y) {x+y}; add(1,2)", "3"},
		{"let add = fn(x,y) {x+y}; let sub = fn(x,y) {x-y}; add(sub(5,3), sub(4,2))", "4"},
		{"len(\"Hello World\")", "11"},
		{"println(10, \"Hello\")", "nil"},
		{"print(10, \"Hello\")", "nil"},
		{"let x = 2++; x", "2"},
		{"let x = 3--; x", "3"},

		{"let x = 12; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 10; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "3"},
		{"let x = 3; let result = if x > 10 {2} else if x > 5 {3} else {4}; result", "4"},
		{"let x = 8; let result = if x >= 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x <= 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x == 8 {2} else if x > 5 {3} else {4}; result", "2"},
		{"let x = 8; let result = if x != 8 {2} else if x > 5 {3} else {4}; result", "3"},
		{`let x = "hello"; let result = if len(x) == 5 { x }; result`, "hello"},
		{"let str = \"Hello \"+\"World!\"; str", "Hello World!"},
		{"let str = \"Hello \"+\"World!\"; str.upper()", "HELLO WORLD!"},

		{"!-5", "false"},
		{"!!!!-5", "true"},
		{"!true", "false"},
		{"!false", "true"},
		{"!nil", "true"},

		{"let arr = [1, 10.5, \"Hello\", true]; arr[0]", "1"},
		{"let arr = [1, 10.5, \"Hello\", true]; arr[1]", "10.5"},
		{"let arr = [1, 10.5, \"Hello\", true]; arr[2]", "Hello"},
		{"let arr = [1, 10.5, \"Hello\", true]; arr[3]", "true"},
		{"let arr = [1, 10.5, \"Hello\", true]; len(arr)", "4"},
		{"let arr = [1, 10.5, \"Hello\", true]; arr.push(\"world\"); arr[4]", "world"},

		//tuple
		{"let tup = (1, 10.5, \"Hello\", true); tup[0]", "1"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[1]", "10.5"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[2]", "Hello"},
		{"let tup = (1, 10.5, \"Hello\", true); tup[3]", "true"},
		{"let tup = (1, 10.5, \"Hello\", true); len(tup)", "4"},
		{"let tup = (1,); len(tup)", "1"},
		{"let tup = (); len(tup)", "0"},
		{`let tup = (1,); tup[0]=10`, "error"},

		//hash
		{`let myHash = {"name": "huanghaifeng", "height": 165}; println(myHash["name"], myHash["height"])`, "nil"},

		//assign
		{`a = "hello world"; a`, "hello world"},
		{`a = "hello world"; a[2]="w"; a`, "hewlo world"},
		{`arr=[1, "hello", true]; arr[0] = "good"; arr[0]`, "good"},
		{`tup=(1, "hello", true); tup[0] = "good"; tup[0]`, "error"},
		{`myHash={}; myHash["name"]="huanghaifeng"; myHash["name"]`, "huanghaifeng"},

		//function statement & function literal
		{`fn add(x,y) {return x+ y} add(2,3)`, "5"},
		{`let sum = fn(x,y) {return x+ y}; sum(2,3)`, "5"},
		{`let sum = fn(x,y) {return x+ y}(2,3); sum`, "5"},

		//for loop
		{`arr = [1, true, "Hello"]; for item in arr { println(item) } println()`, "nil"},
		{`arr = [1, true, "Hello"]; for idx, item in arr { if idx == 2 { break } println(item) } println()`, "nil"},
		{`hash = {"name": "huanghaifeng", "height": 165}; for k, v in hash { print(k, "=", v) println() } println()`, "nil"},
		{`str = "Hello"; for c in str { println(c) } println()`, "nil"},
		{`tup = (1, true, "Hello"); for item in tup { println(item) } println()`, "nil"},
		{`tup = (1, true, "Hello"); for idx, item in tup { if idx == 2 { break } println(item) } println()`, "nil"},

		//while
		{`x = 3; while x-- > 0 { println(x) } println()`, "nil"},
		{`x = 5; while x-- > 0 { println(x); if x == 2 { break } } println()`, "nil"},
		{`x = 5; while x-- > 0 { if x == 4 { continue } else if x == 2 { break } println(x) } println()`, "nil"},

		//do
		{`x = 3; do { x--; println(x) if x == 1 { break } };  println()`, "nil"},

		//multiple assignment & return-values
		{`a, b, c = 1, true, "hello"; println(a) println(b) println(c)`, "nil"},
		{`a, b, c = 2, false, ["x", "y", "z"]; println(a) println(b) println(c[1])`, "nil"},
		{`fn math(x, y) { return x+y, x-y }  add, sub = math(5,3) println(add) println(sub)`, "nil"},
		{`fn xxx(x, y) { return x+y, x-y, x * y }  a, _, c = xxx(5,3) println(a) println(c)`, "nil"},

		//printf
		{`a, b, c, d = 1, true, "hello", 12.343678; printf("a=%g, b=%t, c=%s, d=%.2f\n", a, b, c,d)`, "nil"},
		{`printf("2**3=%g, 2.34.floor=%.0f\n", 2.pow(3), 2.34.floor())`, "nil"},

		// &&, ||
		{`if 10 == 10 && 10 > 5 { printf("10 == 10 && 10 > 5\n")}`, "nil"},
		{`if 10 == 10 && 10 > 12 { printf("10 == 10 && 10 > 12\n") } else { println("10 not larger than 12") }`, "nil"},
		{`if 10 == 10 || 10 > 12 { printf("10 == 10 || 10 > 12\n")}`, "nil"},
		{`if 10 == 11 || 10 > 12 { printf("10 == 11 || 10 > 12\n") } else { println(" 10 not equal 11 and 10 not larger than 12") }`, "nil"},

		//import
		{`import examples.sub_package.calc; println(Add(2,3))`, "nil"},
		{`import examples.sub_package.calc; println(_add(2,3))`, "error"},

		//regexp
		{`name = "Huang HaiFeng"; if name =~ /huang/i { println("Hello Huang") }`, "nil"},
		{`name = "Huang HaiFeng"; if ( name !~ /xxx/ ) { println( "Hello xxx" ) }`, "nil"},
		{`name = "Huang HaiFeng"; if name =~ /Huang/ { println("Hello Huang") }`, "nil"},
		{`match = /\d+\t/.match("abc 123	mnj"); if (match) { println("matched") }`, "nil"},
		{`arr = / /.split("ba na za"); if (len(arr) > 0) { println("/ /.split('ba na za')[1]=", arr[1]) } else { println("Not splitted") }`, "nil"},

		//go object
		{`fmt.Printf("Hello %s!\n", "go function"); println()`, "nil"},
		{`s = fmt.Sprintf("Hello %s!", "World"); println(s)`, "nil"},
		{`fmt.Println(runtime.GOOS); println()`, "nil"},
		{`fmt.Println(runtime.GOARCH); println()`, "nil"},

		//anonymous functions(lambdas)
		{`let add = fn (x, factor) { x + factor(x) } result = add(5, (x) => x * 2) println(result)`, "nil"},
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

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("%s", evaluated.Inspect())
			} else {
				fmt.Printf("%s = %s\n", tt.input, tt.expected)
			}
		}
	}
}

/*
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
*/

// Register go package methods/types
// Here we demonstrate the use of import go language's methods.
func RegisterGoGlobals() (err error) {
	err = eval.RegisterGoFunctions("fmt", map[string]interface{}{
		"Println":  fmt.Println,
		"Print":    fmt.Print,
		"Printf":   fmt.Printf,
		"Sprintf":  fmt.Sprintf,
		"Sprintln": fmt.Sprintln,
	})
	if err != nil {
		return
	}

	err = eval.RegisterGoVars("runtime", map[string]interface{}{
		"GOOS":   runtime.GOOS,
		"GOARCH": runtime.GOARCH,
	})

	return
}

func runProgram(filename string) {
	l, err := lexer.NewFileLexer(filename)
	if err != nil {
		fmt.Printf("error reading %s\n", filename)
		os.Exit(1)
	}

	p := parser.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	scope := eval.NewScope(nil, os.Stdout)

	result := eval.Eval(program, scope)
	if result.Type() == eval.ERROR_OBJ {
		fmt.Println(result.Inspect())
	}
}

func runWithEmbedFile() bool {
	attachments, err := ember.Open()
	if err != nil {
		return false
	}
	defer attachments.Close()

	contents := attachments.List()
	if len(contents) == 0 {
		return false
	}

	name := "main"
	foundMain := false
	for _, content := range contents {
		if content == name {
			foundMain = true
			break
		}
	}
	if !foundMain {
		return false
	}

	buf, err := attachments.GetResource(name)
	if err != nil {
		fmt.Printf("error reading embedded file.\n", err)
		return false
	}

	str := string(buf)
	l := lexer.NewLexer(str)
	p := parser.NewParser(l)
	p.Attachments = attachments
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	scope := eval.NewScope(nil, os.Stdout)

	result := eval.Eval(program, scope)
	if result.Type() == eval.ERROR_OBJ {
		fmt.Println(result.Inspect())
		os.Exit(1)
	}

	return true
}

func main() {
	if runWithEmbedFile() {
		return
	}

	args := os.Args[1:]

	err := RegisterGoGlobals()
	if err != nil {
		fmt.Printf("RegisterGoGlobals failed: %s\n", err)
		os.Exit(1)
	}

	if len(args) == 1 {
		runProgram(args[0])
	} else {
		TestEval()
	}
}
