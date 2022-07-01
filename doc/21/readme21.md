# `命名函数(Named function)`支持

在这一节中，我们要加入对于`命名函数`的支持。先来看一下`命名函数`的例子：

```go
fn add(x, y) { //命名函数声明
    return x + y
}
let sum = add(2,3) //函数调用
```

我们再来看一下之前我们学习的`函数字面量表达式`的声明：

```go
fn (x,y) { //函数字面量表达式的声明
	return x + y
}(2,3) //直接函数调用
```

仔细观察的话就会发现，`命名函数`其实就是`函数字面量表达式`中的`fn`关键字的后面多了一个`标识符`（这个标识符就是函数名）。

下面看一下我们需要做的更改：

1. 因为没有引入任何新的关键字或者操作符，所以词元（Token）源码`token.go`无需更改
2. 同样词法分析器（Lexer）源码`lexer.go`也无需更改
3. 在抽象语法树（AST）的源码`ast.go`中修改`函数字面量`的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中修改对`函数字面量`的语法解析。
5. 在解释器（Evaluator）的源码`eval.go`中修改对`函数字面量`的解释。



## 抽象语法树（AST）的更改

从本节开头处的描述中，我们知道`命名函数`实际上就是`函数字面量表达式`的变形，只不过多了一个函数名而已。所以这里我们只需要给`函数字面量表达式`加入一个名字字段即可：

```go
//ast.go
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Name       string      // 函数的名字
	Parameters []*Identifier
	Body       *BlockStatement
}

//函数字面量的字符串表示
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(" ")
		out.WriteString(fl.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {")
	out.WriteString(fl.Body.String())
	out.WriteString("}")

	return out.String()
}
```

我们在`FunctionLiteral`结构中新加了一个`Name`字段（代码第4行）。同时在`函数字面量的字符串表示`方法中，加入了相关的逻辑（代码19-22行）。



## 语法解析器（Parser）的更改

我们只需要更改`parseFunctionLiteral`这个函数，使其能够识别函数名即可：

```go
//parser.go
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if p.peekTokenIs(token.TOKEN_IDENTIFIER) { //判断'fn'后面是否是一个标识符
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(token.TOKEN_LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.TOKEN_LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}
```

第5-8行是新增的代码。



## 解释器（Evaluator）的更改

同样的，我们需要对`evalFunctionLiteral`函数做少量的变更：

```go
//eval.go
func evalFunctionLiteral(fl *ast.FunctionLiteral, scope *Scope) Object {
	fn := &Function{Literal: fl, Scope: scope}
	if fl.Name != "" {
		scope.Set(fl.Name, fn)
	}
	return fn
}
```

代码4-6行是新增的。我们判断函数是否有名字，如果有的话，将其函数名作为key，函数对象（这里是`fn`）作为value放入scope中。



## 测试

下面我们写一个简单的程序测试一下：
```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{`fn add(x,y) {return x+ y} add(2,3)`, "5"},
		{`fn add(x,y) { return fn(x) { x } (x) + fn(x) {x}(y) } add(2,3)`, "5"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()

		scope := eval.NewScope(nil, os.Stdout)
		evaluated := eval.Eval(program, scope)
		if evaluated != nil {
			if evaluated.Inspect() != tt.expected {
				fmt.Printf("%s\n", evaluated.Inspect())
			} else {
				fmt.Printf("%s = %s\n", tt.input, tt.expected)
			}
		}
	}
}

func main() {
	TestEval()
}
```

这里需要对第二个测试的语句进行一下说明，我们把它用多行表示：

```go
fn add(x,y) {
  return 
    fn(x) { x } (x) //直接将add函数的第一个参数'x'传递给'函数字面量表达式'(直接执行)
           + 
    fn(x) { x } (y) //直接将add函数的第二个参数'y'传递给'函数字面量表达式'(直接执行)。 这里的'x'是形参，使用啥名字都可以
}
add(2,3)
```



下一节，我们会让解释器在浏览器中运行（通过使用`go`对wasm支持）。

