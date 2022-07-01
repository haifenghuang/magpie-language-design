# `复合赋值运算符`支持

这一节中，我们将加入`复合赋值运算符（Compound assignment operators）`的支持。先来看一下使用例：

```go
a = 10
a += 10  //等价于  a = a + 10
a -= 10  //等价于  a = a - 10
a *= 10  //等价于  a = a * 10
a /= 10  //等价于  a = a / 10
a %= 10  //等价于  a = a % 10

str = "hello"
str += " world"
```

上面例子中的`+=`、`-=`、`*=`、`/=`、`%/`即复合赋值运算符。使用这几种复合赋值运算符，有时候还是很方便的。至少让我们省去了好多敲击键盘的时间，:smile:

> 对于数字类型，我们支持`+=`、`-=`、`*=`、`/=`、`%/`这五种复合赋值运算符。
>
> 对于字符串类型，我们仅支持`+=`一种运算符。

下面看一下我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`复合赋值运算符`的识别
2. 在语法解析器（Parser）的源码`parser.go`中加入对`复合赋值运算符`的语法解析及运算符优先级。
4. 在解释器（Evaluator）的源码`eval.go`中修改对`复合赋值运算符`的解释。



## 词元（Token）的更改

```go
//token.go
const (
	//...

	TOKEN_PLUS_A     // +=
	TOKEN_MINUS_A    // -=
	TOKEN_ASTERISK_A // *=
	TOKEN_SLASH_A    // /=
	TOKEN_MOD_A      // %=

	//...
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...

	case TOKEN_PLUS_A:
		return "+="
	case TOKEN_MINUS_A:
		return "-="
	case TOKEN_ASTERISK_A:
		return "*="
	case TOKEN_SLASH_A:
		return "/="
	case TOKEN_MOD_A:
		return "%="

	//...
}
```
第5行和15-16行是新增的代码。



## 词法分析器（Lexer）的更改

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '+':
		if l.peek() == '+' {
			tok = token.Token{Type: token.TOKEN_INCREMENT, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_PLUS_A, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_PLUS, l.ch)
		}
	case '-':
		if l.peek() == '-' {
			tok = token.Token{Type: token.TOKEN_DECREMENT, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_MINUS_A, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MINUS, l.ch)
		}
	case '*':
		if l.peek() == '*' {
			tok = token.Token{Type: token.TOKEN_POWER, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '=' {
			tok = token.Token{Type: token.TOKEN_ASTERISK_A, 
                              Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else {
			tok = newToken(token.TOKEN_MULTIPLY, l.ch)
		}
	case '/':
		//...
		if prevToken.Type == token.TOKEN_RPAREN || // (a+c) / b
			prevToken.Type == token.TOKEN_RBRACKET || // a[3] / b
			prevToken.Type == token.TOKEN_IDENTIFIER || // a / b
			prevToken.Type == token.TOKEN_NUMBER { // 3 / b,  3.5 / b
			if l.peek() == '=' {
				tok = token.Token{Type: token.TOKEN_SLASH_A, 
                                  Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.TOKEN_DIVIDE, l.ch)
			}
		}
		//...        

	//...
	}

	//...
}
```

这些都是读者非常熟悉的改动了，因此这里不多做解释了。

## 语法解析器（Parser）的更改

首先我们要给这些``复合赋值运算符`增加中缀回调函数。

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerInfix(token.TOKEN_PLUS_A, p.parseAssignExpression)
	p.registerInfix(token.TOKEN_MINUS_A, p.parseAssignExpression)
	p.registerInfix(token.TOKEN_ASTERISK_A, p.parseAssignExpression)
	p.registerInfix(token.TOKEN_SLASH_A, p.parseAssignExpression)
	p.registerInfix(token.TOKEN_MOD_A, p.parseAssignExpression)
	//...
}
```

其次，我们需要给`复合赋值运算符`（中缀操作符）赋优先级。它们的优先级和`=`的优先级是一样的：

```go
//parser.go
var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_PLUS_A:     ASSIGN,
	token.TOKEN_MINUS_A:    ASSIGN,
	token.TOKEN_ASTERISK_A: ASSIGN,
	token.TOKEN_SLASH_A:    ASSIGN,
	token.TOKEN_MOD_A:      ASSIGN,
	//...
}
```

没有什么太多难理解的。

## 解释器（Evaluator）的更改

对于字符串类型，我们需要更改`evalNumAssignExpression`这个函数：

````go
//eval.go
func evalNumAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	if left.Type() == NUMBER_OBJ && val.Type() == NUMBER_OBJ { //如果左右表达式都是数字类型
		leftVal := left.(*Number).Value
		rightVal := val.(*Number).Value

		var result float64
		switch a.Token.Literal {
		case "+=":
			result = leftVal + rightVal
		case "-=":
			result = leftVal - rightVal
		case "*=":
			result = leftVal * rightVal
		case "/=":
			result = leftVal / rightVal
		case "%=":
			result = math.Mod(leftVal, rightVal)
		}

		ret = NewNumber(result)
		scope.Set(name, ret)//将其设置到scope中
		return
	}
	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}
````

对于字符串类型，我们需要更改`evalStrAssignExpression`函数：

```go
//eval.go
func evalStrAssignExpression(a *ast.AssignExpression, name string, left Object, scope *Scope, val Object) (ret Object) {
	leftVal := left.(*String).String

	switch a.Token.Literal {
	//...
	case "+=":
		if left.Type() == STRING_OBJ && val.Type() == STRING_OBJ {//如果左右表达式都是字符串类型
			leftVal := left.(*String).String
			rightVal := val.(*String).String
			ret = NewString(leftVal+rightVal)
			scope.Set(name, ret) //设置到scope中
			return
		}
	}

	return newError(a.Pos().Sline(), ERR_INFIXOP, left.Type(), a.Token.Literal, val.Type())
}
```

没有太多需要解释的，这里需要提醒的是`evalNumAssignExpression`函数的第22行，和`evalStrAssignExpression`函数的第12行。就是别忘记返回，否则程序就会走到`return newError`的地方。我开始实现的时候，就忘记返回了:smile:。



## 测试

这一节整体来说，读者读起来应该是比较轻松的。下面让我们写个简单的测试程序：

```javascript
a = 5
a += 12

b = 8
b -= 3

c = 12
c *= 2

d = 16
d /= 2

printf("a=%d, b=%d, c=%d, d=%d\n", a, b, c, d)

s = "hello"
s += " world"
println(s)
```



下一节，我们将实现真正意义上的多重赋值。



