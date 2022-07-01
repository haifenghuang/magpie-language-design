# `匿名函数`支持

这一节中，我们将给`magpie语言`增加匿名函数（anonymous function）的支持。如果读者学过`c#`的话，这个实际上就类似`c#`的`Lambda表达式`。先来看一下使用例：

```javascript
# anonymous functions(lambdas)
let add = fn (x, factor) {
  x + factor(x)
}
result = add(5, (x) => x * 2)
//result = add(5, x => x * 2) //只有一个参数的时候，'x'两边的括号可以省略
println(result)  # result: 15
```

第5行`add`函数的第二个参数就是一个`Lambda表达式`。从上面的例子中我们可以看出`Lambda表达式`的左边是括号括起来的参数（只有一个参数的情况下，括号可以省略），右边是一个表达式，其实也可以是花括弧括起来的语句块。我们来看一下其一般形式：

```c#
(input-parameters) => expression
(input-parameters) => { block }
```

从这个一般形式中我们可以得出结论：`=>`是个中缀操作符。

如果我们把这个`Lambda表达式`和`函数字面量(function literal)`进行一下对比：

```
fn (input-parameters) { block }
```

可以看到，它和`Lambda表达式`非常的相似（只不过lambda表达式没有了`fn`关键字，同时在参数和块语句间多了`=>`操作符），实际上我们的语法解析器就是将其当作`函数字面量(function literal)`来处理的。这样做的好处就是我们不用修改解释器（Evaluator）的代码。

现在让我们看一下需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型（`=>`）
2. 在词法分析器（Lexer）的源码`lexer.go`中加入对`=>`的解析
3. 在语法解析器（Parser）的源码`parser.go`中加入对`=>`的语法解析

## 词元（Token）的更改

因为都是读者再熟悉不过的内容，不做解释，直接看代码：

```go
//token.go
const (
	//...
	TOKEN_FATARROW // =>
)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_FATARROW:
		return "=>"
	}
}
```



## 词法分析器（Lexer）的更改

我们的词法分析器需要能够识别`=>`。

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '=':
		if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_EQ, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '~' {
			tok = token.Token{Type: token.TOKEN_MATCH, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '>' {
			tok = token.Token{Type: token.TOKEN_FATARROW, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_ASSIGN, l.ch)
		}
		//...
	}

}
```

16-19行是新增的代码。



## 语法解析器（Parser）的更改

对于`=>`操作符，我们需要给它注册中缀表达式回调函数：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerInfix(token.TOKEN_FATARROW, p.parseFatArrow)
}

func (p *Parser) parseFatArrow(left ast.Expression) ast.Expression {
}
```

`parseFatArrow`函数我们之后会讲到。讲解之前，让我们看一下对于`() => 5`这种`Lambda表达式`来说，我们需要如何处理。如果读者还有印象的话，`parseGroupedExpression`函数是用来处理`(`的，因此我们需要更改此函数来增加对这种形式的判断：

```go
//parser.go
type Parser struct {
	//...
	savedToken token.Token //主要使用在解析匿名函数中
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	savedToken := p.curToken
	p.savedToken = p.curToken //将括号这个词元保存起来
	p.nextToken()

	// NOTE: 如果之前的词元是类型是TOKEN_LPAREN, 当前的词元类型是
    //       TOKEN_RPAREN, 则表示是一个空的括弧，即`()`
	if savedToken.Type == token.TOKEN_LPAREN && p.curTokenIs(token.TOKEN_RPAREN) {
		if p.peekTokenIs(token.TOKEN_FATARROW) { //例如 '() => 5'
			p.nextToken() //skip current token
			ret := p.parseFatArrow(nil)
			return ret
		}

		//empty tuple, e.g. 'x = ()'
		return &ast.TupleLiteral{Token: savedToken, Members: []ast.Expression{}}
	}

	//...
}
```

15-19行是新增加的代码，我们判断空括号`()`后面如果跟着一个`=>`则就认为是一个`Lambda表达式`。就会调用`parseFatArrow`函数来进行解析。需要注意的是，`parseFatArrow`函数接受的参数是一个`left`表达式，而这里我们传入的是`nil`（因为括号中没有参数）。我们在`Parser`结构中定义了一个`savedToken`变量（代码第4行），用来记住`(`的位置。

接下来让我们看一下`parseFatArrow`函数的实现：

```go
//parser.go

//(x, y) => x + y + 5      左边的表达式是一个'*TupleLiteral'
//(x) => x + 5             左边的表达式是一个'*Identifier'
// x  => x + 5             左边的表达式是一个'*Identifier'
//()  => 5 + 5             左边的表达式为'nil'
func (p *Parser) parseFatArrow(left ast.Expression) ast.Expression {
	var pos token.Position
	if left != nil {
		pos = left.Pos()
	} else { //如果left参数为nil的话
		pos = p.savedToken.Pos //使用`parseGroupedExpression`函数中保存的`savedToken`的位置
	}

    //构造一个函数字面量表达式（function literal）
	tok := token.Token{Pos: pos, Type: token.TOKEN_FUNCTION, Literal: "fn"}
	fn := &ast.FunctionLiteral{Token: tok}
    
	switch exprType := left.(type) { //判断左边的表达式的类型
	case nil:
		//没有参数
	case *ast.Identifier:
		//单个参数
		fn.Parameters = append(fn.Parameters, exprType)
	case *ast.TupleLiteral:
		// 多个参数
		for _, v := range exprType.Members {
			switch param := v.(type) {
			case *ast.Identifier:
				fn.Parameters = append(fn.Parameters, param)
			default: //不是标识符，则报错
				msg := fmt.Sprintf("Syntax Error:%v- Arrow function expects a list of identifiers as arguments", param.Pos())
				p.errors = append(p.errors, msg)
				p.errorLines = append(p.errorLines, param.Pos().Sline())
				return nil
			}
		}
	default:
		msg := fmt.Sprintf("Syntax Error:%v- Arrow function expects identifiers as arguments", exprType.Pos())
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, exprType.Pos().Sline())
		return nil
	}

	p.nextToken()
	if p.curTokenIs(token.TOKEN_LBRACE) { //块语句，使用'parseBlockStatement'来解析
		fn.Body = p.parseBlockStatement()
    } else { //非块语句,例如(x) => x + 2或者(x) => return x + 2, 则我们手动构造一个块语句
		/* 注意：这里如果我们使用`parseExpressionStatement`,则下面的例子会报错：
		      (x) => return x  //error: no prefix parse functions for 'RETURN' found

		   `parseExpressionStatement`可以解析`(x) => x + 2`这种形式，因为这里的`x + 2`是一个
		   'ExpressionStatement',而对于`(x) => return x`这种形式，这里的`return x`是一个
		   `ReturnStatement`。所以我们需要使用`parseStatement`函数来处理这两种形式。
		*/
		fn.Body = &ast.BlockStatement{
			Statements: []ast.Statement{
				p.parseStatement(),
			},
		}
	}
	return fn
}
```

代码看起来比较多，但是实际上不是特别复杂。唯一需要注意的是对于`(x) => x + 2`和`(x) => return x+2`这两种形式，为了和`(x) => { body }`一样统一处理，我们手动构造了一个只包含一个语句的`BlockStatement`。

最后，别忘记`=>`是一个中缀符号，我们需要给其赋优先级：

```go
//parser.go
var precedences = map[token.TokenType]int{
	token.TOKEN_ASSIGN:   ASSIGN,
	token.TOKEN_FATARROW: ASSIGN,
	//...
}
```

这里，我们将`=>`赋的优先级和赋值操作符`=`的优先级一样。

> 这里对`=>`优先级的设置是参照的`c#`语言中，对`Lambda表达式`的优先级设置。

## 测试

```javascript
# function.mp
# anonymous functions(lambdas)
let add = fn (x, factor) {
  x + factor(x)
}
result = add(5, (x) => x * 2)
//或者result = add(5, x => x * 2)
println(result)  # result: 15

```



下一节，我们会增加对链式比较操作符（`a < b <=c`）的支持。



