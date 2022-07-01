# while和do循环支持

在这一节中，我们将加入对于`while`和`do`循环的支持。

相对于`for`循环的处理，`while`和`do`循环简单得多。因为`for`循环有多种类型，而`while`和`do`都只有一种表示类型。我们先来看一下它们的例子：

```c
//while循环
i = 10 
while i > 3 {
	println(i)
	i--
}

a = 10
do {
    if a < 5 { break }
    println(a)
    a--
}
```

比较简单，你应该能够马上抽象出它们的一般表示形式：

```javascript
while <condition> { block }
do { block } //这个和for {block}几乎一样，除了关键字不一样外
```

下面来看一下需要做的更改。

### 词元（Token）的更改

我们需要增加了两个新的关键字：`while`和`do`。

```go
//token.go
const (
	//...
	TOKEN_WHILE    //while
	TOKEN_DO       //do
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_WHILE:
		return "WHILE"
	case TOKEN_DO:
		return "DO"
	}
}

var keywords = map[string]TokenType{
	//...
	"while":    TOKEN_WHILE,
	"do":       TOKEN_DO,
}
```



### 抽象语法树的（AST）更改

根据上面的`while`和`do`循环的一般形式我们很容易就能够得出它们的抽象语法表示，直接看代码：

```go
//ast.go

//while condition { block }
type WhileLoop struct {
	Token     token.Token
	Condition Expression      //条件
	Block     *BlockStatement //块语句
}

func (wl *WhileLoop) Pos() token.Position {
	return wl.Token.Pos
}

func (wl *WhileLoop) End() token.Position {
	return wl.Block.End()
}

func (wl *WhileLoop) expressionNode()      {}
func (wl *WhileLoop) TokenLiteral() string { return wl.Token.Literal }

func (wl *WhileLoop) String() string {
	var out bytes.Buffer

	out.WriteString("while")
	out.WriteString(wl.Condition.String())
	out.WriteString("{")
	out.WriteString(wl.Block.String())
	out.WriteString("}")

	return out.String()
}

//do { block }
type DoLoop struct {
	Token token.Token
	Block *BlockStatement //块语句
}

func (dl *DoLoop) Pos() token.Position {
	return dl.Token.Pos
}

func (dl *DoLoop) End() token.Position {
	return dl.Block.End()
}

func (dl *DoLoop) expressionNode()      {}
func (dl *DoLoop) TokenLiteral() string { return dl.Token.Literal }

func (dl *DoLoop) String() string {
	var out bytes.Buffer

	out.WriteString("do")
	out.WriteString(" { ")
	out.WriteString(dl.Block.String())
	out.WriteString(" }")
	return out.String()
}
```



### 语法解析器（Parser）的更改

首先我们需要对`while`和`do`这两个关键字注册前缀表达式回调函数：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerPrefix(token.TOKEN_WHILE, p.parseWhileLoopExpression)
	p.registerPrefix(token.TOKEN_DO, p.parseDoLoopExpression)

}
```

然后，我们得实现`while`和`do`表达式的解析：

```go
//parser.go
//do { block }
func (p *Parser) parseDoLoopExpression() ast.Expression {
	p.loopDepth++
	loop := &ast.DoLoop{Token: p.curToken}

	p.expectPeek(token.TOKEN_LBRACE)
	loop.Block = p.parseBlockStatement()

	p.loopDepth--
	return loop
}

//while condition { block }
func (p *Parser) parseWhileLoopExpression() ast.Expression {
	p.loopDepth++
	loop := &ast.WhileLoop{Token: p.curToken}

	p.nextToken()
	loop.Condition = p.parseExpressionStatement().Expression

	if p.peekTokenIs(token.TOKEN_RPAREN) {
		p.nextToken()
	}

	if p.peekTokenIs(token.TOKEN_LBRACE) {
		p.nextToken()
		loop.Block = p.parseBlockStatement()
	} else {
		msg := fmt.Sprintf("Syntax Error:%v- for loop must be followed by a '{'", 
							p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	p.loopDepth--
	return loop
}
```



### 解释器(Evaluator)的更改

同样，由于内容比较简单，我们直接来看代码：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	switch node := node.(type) {
	//...
	case *ast.DoLoop:
		return evalDoLoopExpression(node, scope)
	case *ast.WhileLoop:
		return evalWhileLoopExpression(node, scope)
	}

	return nil
}
```

我们加入了两个`case`分支，用来解释`while`和`do`循环。接着来看一下它们的实现：

```go
//eval.go

//do { block }
// 返回值：
//    1. 最后执行的表达式的值
//    2. NIL
//    3. 返回对象(Return Object)
func evalDoLoopExpression(dl *ast.DoLoop, scope *Scope) Object {
	var e Object = NIL
	for {
		e = Eval(dl.Block, scope) //解释块语句
		if e.Type() == ERROR_OBJ {
			return e
		}

		if _, ok := e.(*Break); ok {
			break
		}
		if _, ok := e.(*Continue); ok {
			continue
		}
		if v, ok := e.(*ReturnValue); ok {
			return v
		}
	}

	if e == nil || e.Type() == BREAK_OBJ || e.Type() == CONTINUE_OBJ {
		return NIL
	}

	return e
}

//while condition { block }
// 返回值：
//    1. 最后执行的表达式的值
//    2. NIL
//    3. 返回对象(Return Object)
func evalWhileLoopExpression(wl *ast.WhileLoop, scope *Scope) Object {
	var result Object = NIL
	for {
		condition := Eval(wl.Condition, scope) //条件
		if condition.Type() == ERROR_OBJ {
			return condition
		}

		if !IsTrue(condition) {
			return NIL
		}

		result = Eval(wl.Block, scope)
		if result.Type() == ERROR_OBJ {
			return result
		}

		if _, ok := result.(*Break); ok {
			break
		}
		if _, ok := result.(*Continue); ok {
			continue
		}
		if v, ok := result.(*ReturnValue); ok {
			return v
		}
	}

	if result == nil || result.Type() == BREAK_OBJ || result.Type() == CONTINUE_OBJ {
		return NIL
	}

	return result
}
```

所有的这些都是我们非常熟悉的代码，这里就不再详细解释了。至于22-24和59-61行的`if`判断，上一篇文章也讲的比较清楚了。如果还有读者不清楚的话，请参照前一篇文章的讲解。



## 测试

```go
//main.go

func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		//while
		{`x=3;while x-- > 0 { println(x) } println()`, "nil"},
		{`x=5;while x-- > 0 { println(x); if x == 2 { break } } println()`, "nil"},
		{`x=5;while x-- > 0 {if x==4 {continue} else if x==2 { break } println(x)} println()`, "nil"},

		//do
		{`x = 3; do { x--; println(x) if x == 1 { break } };  println()`, "nil"},

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

func main() {
	TestEval()
}
```



下一节，我们将加入对`多重赋值和多返回值`的支持。
