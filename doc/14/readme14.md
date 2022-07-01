# 取反（!）支持

为了能够将来海阔天空，小喜鹊天天锻炼，非常的辛苦！偶尔也需要打个小盹，休息一下。

在这一节中，我们会提供对取反（！）操作的支持（这一节学起来应该比较轻松）。



一样的老套路，来看看对于取反操作，我们需要做的更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型（TOKEN_BANG）
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`取反操作符`的识别
3. 在语法解析器（Parser）的源码`parser.go`中注册`取反操作符`的前缀回调函数。
5. 在解释器（Evaluator）的源码`eval.go`中加入对`取反操作符`的解释。



这一节，我想采取于以往的章节不同的方式，让读者自己来看代码，我只列出代码，而不加以说明。

### 词元（Token）的更改

```go
//token.go
const (
	//...
	TOKEN_BANG      // !

)
//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_BANG:
		return "!"
	//...
	}
}
```



### 词法分析器（Lexer）的更改

```go
//lexer.go

//获取下一个词元（Token）
func (l *Lexer) NextToken() token.Token {
	switch l.ch {
	//...
	case '!':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_NEQ, 
                              Literal: string(l.ch) + string(l.peek())} // !=
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_BANG, l.ch) // !
		}
	}
}

```



### 语法解析器（Parser）的更改

```go
//parser.go
func (p *Parser) registerAction() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	//...
	p.registerPrefix(token.TOKEN_BANG, p.parsePrefixExpression)
}
```

### 解释器（Evaluator）的更改

```go
//eval.go
func evalPrefixExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	switch node.Operator {
	//...
	case "!":
		return evalBangOperatorExpression(node, right, scope)
	//...
	}
}

func evalBangOperatorExpression(node *ast.PrefixExpression, right Object, scope *Scope) Object {
	return nativeBoolToBooleanObject(!IsTrue(right))
}
```

### 测试

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"!-5", "false"},
		{"!!!!-5", "true"},
		{"!true", "false"},
		{"!false", "true"},
		{"!nil", "true"},
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



就是这么简单。有的读者可能会说，你这也太敷衍了事了吧。是的，这不是想让读者轻松点，我自己也轻松一下吗？:smile:



下一节，我们会增加对`数组`的支持。

