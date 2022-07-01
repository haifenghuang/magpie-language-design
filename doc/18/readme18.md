# `后缀表达式(++、--)`支持

在这一节中，我们将给语言提供后缀表达式`++和--`支持。这一节讲解的内容，对后面我们将要讲解的循环，可以说是一个预热。

我们看一下形式：

```perl
var++
var--
```

很显然，我们这一节需要增加两个新的词元。

让我们来看一下需要对代码做的更改：

1. 在词元（Token）的源码`token.go`中加入新的词元（Token）类型
1. 在词法分析器（Lexer）的源码`lexer.go`中加入对新的词元的分析。
2. 在抽象语法树（AST）的源码`ast.go`中加入`后缀表达式`对应的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中加入对`后缀表达式`的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中加入对`后缀表达式`的解释。



## 词元（Token）更改

```go
//token.go
const (
    //...
	TOKEN_INCREMENT // ++
	TOKEN_DECREMENT // --
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_INCREMENT:
		return "++"
	case TOKEN_DECREMENT:
		return "--"
	//...
	}
}
```



## 词法分析器（Lexer）的更改

```go
func (l *Lexer) NextToken() token.Token {
	//...
	switch l.ch {
	case '+':
		if l.peek() == '+' { //++
			tok = token.Token{Type: token.TOKEN_INCREMENT, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_PLUS, l.ch)
		}
	case '-':
		if l.peek() == '-' { //--
			tok = token.Token{Type: token.TOKEN_DECREMENT, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MINUS, l.ch)
		}
	//...
	}

	//...
}
```

第5行和第12行的`if`分支分别对后缀的`++、--`进行了处理。



## 抽象语法树（AST）的更改

我们再回顾一下，后缀表达式的形式：

```perl
var++
var--
```

仔细观察的话，虽然我们这里说的是后缀，但是我们可以将其看作一个中缀表达式进行处理：

```perl
<left-expression> ++ <其它表达式>
<left-expression> -- <其它表达式>
```

只不过对于后缀表达式，我们不关心这里的所谓`<其它表达式>`，对于这个`<其它表达式>`，程序中已经有了相关的处理。例如：

```perl
10++ + 12 
```

这里的表达式经过解析器来解析后的结果如下：

```go
((l0++) + 12)
```

下面来看一下后缀表达式的抽象语法表示：

```go
//ast.go
//后缀表达式: ++,--
type PostfixExpression struct {
	Token    token.Token
	Left     Expression //左表达式
	Operator string //操作符：++、--
}

func (pe *PostfixExpression) Pos() token.Position {
	return pe.Token.Pos
}

func (pe *PostfixExpression) End() token.Position {
	ret := pe.Left.End()
	ret.Col = ret.Col + len(pe.Operator)
	return ret
}

func (pe *PostfixExpression) expressionNode() {}

func (pe *PostfixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PostfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Left.String())
	out.WriteString(pe.Operator)
	out.WriteString(")")

	return out.String()
}
```



## 语法解析器（Parser）的更改

刚才也提到，对于`++、--`这两个后缀操作符，实际上我们可以将其看作中缀操作符。因此我们注册的是中缀表达式回调函数。

我们需要做三处更改：

1. 对新增加的词元类型（Token type）注册中缀表达式回调函数
2. 新增解析`后缀表达式`的函数
2. 对后缀操作符`++`和`--`赋予优先级

我们先来看看优先级的更改。请看一个例子：

```perl
10++ + 12
10++ * 12
-10++ + 12
```

我们期望解析器解析后的结果如下：

```perl
(   ( 10++ ) + 12 )
(   ( 10++ ) * 12 )
((- ( 10++ )) + 12 )
```

就是说我们希望这几个例子中的`++`、`--`最先被处理，因此它的优先级应该比`前缀操作符（!、+、-）`的优先级高：

```go
//parser.go
const (
	//...
	PREFIX      //!true, -10, +10
	INCREMENT   //++, --
	//...
)

var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_INCREMENT: INCREMENT,
	token.TOKEN_DECREMENT: INCREMENT,
}
```

接下来，我们看一下前两点的更改：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	//下面实际上注册的是中缀表达式回调函数
	p.registerInfix(token.TOKEN_INCREMENT, p.parsePostfixExpression)
	p.registerInfix(token.TOKEN_DECREMENT, p.parsePostfixExpression)
}

func (p *Parser) parsePostfixExpression(left ast.Expression) ast.Expression {
	return &ast.PostfixExpression{Token: p.curToken, Left: left, Operator: p.curToken.Literal}
}
```

第5-6行，我们给`TOKEN_INCREMENT`和`TOKEN_DECREMENT`注册了中缀表达式回调函数。

第9行`parsePostfixExpression`就更简单了，直接生成`PostfixExpression`结构返回。



## 解释器（Evaluator）的更改

我们需要在解释器（Evaluator）的`Eval`函数的`switch`分支中加入对`后缀表达式`的处理：

```go
//eval.go
func Eval(node ast.Node) (val Object) {
	switch node := node.(type) {
	//...

	case *ast.PostfixExpression:
		left := Eval(node.Left, scope) //处理左表达式,即处理`left++`中的`left`
		if left.Type() == ERROR_OBJ {
			return left
		}
		return evalPostfixExpression(node, left, scope)
	//...
	}
	return nil
}

func evalPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch node.Operator {
	case "++":
		return evalIncrementPostfixExpression(node, left, scope)
	case "--":
		return evalDecrementPostfixExpression(node, left, scope)
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}
    
//后缀++： left++
func evalIncrementPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch left.Type() {
	case NUMBER_OBJ:
		leftObj := left.(*Number)
		returnVal := NewNumber(leftObj.Value) //取出左表达式的值
		scope.Set(node.Left.String(), NewNumber(leftObj.Value+1)) //将值加1后，再写回Scope中
		return returnVal
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}

//后缀--: left--
func evalDecrementPostfixExpression(node *ast.PostfixExpression, left Object, scope *Scope) Object {
	switch left.Type() {
	case NUMBER_OBJ:
		leftObj := left.(*Number)
		returnVal := NewNumber(leftObj.Value) //取出左表达式的值
		scope.Set(node.Left.String(), NewNumber(leftObj.Value-1)) //将值加1后，再写回Scope中
		return returnVal
	default:
		return newError(node.Pos().Sline(), ERR_POSTFIXOP, node.Operator, left.Type())
	}
}
```

内容虽然有点多，但是没有太复杂的地方。



`evalIncrementPostfixExpression`和`evalDecrementPostfixExpression`函数中，我们增加了一个错误常量`ERR_POSTFIXOP`，这个是在`errors.go`中定义的：

```go
//errors.go
var (
	//...
	ERR_POSTFIXOP    = "unsupported operator for postfix expression:'%s' and type: %s"
)
```



## 测试

```go
//main.go
func TestEval() {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 2++ + 5; x", "7"},
		{"let x = 2++ * 5; x", "10"},
		{"let x = -2-- * 5; x", "-10"},
		{"let i = 0; let arr =[1,2,3]; arr[i++]", "1"},
	}

	for _, tt := range tests {
		l := lexer.NewLexer(tt.input)
		p := parser.NewParser(l)
		program := p.ParseProgram()
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
	TestEval()
}
```



下一节，我们将讲解对`赋值表达式(=)`的支持。
