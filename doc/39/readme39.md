# `链式比较操作符`支持

如果读者熟悉python语言的话，可能会知道python语言可以支持`链式比较操作符（Comparison operator Chaining）`，像下面这样：

```python
if a < b < c :
    pass

#等价于
if a < b and b < c :
   pass
```

在这一节中，让我们来加入这个支持。从上面这个例子中，我们可以看出，这个`链式比较操作符`的形式如下：

```xml
#例：   a                <                      b                  < =                c
<left-expression> <compare-operator> <right-expression> <next-compare-operator> <next-expression>
```

有了上面的介绍，让我们来看看需要做哪些更改：

1. 在抽象语法树(AST)的源码`ast.go`中，加入对这种表达式的抽象语法表示。
2. 在语法解析器（Parser）的源码`parser.go`中，加入对这种形式的表达式的解析。
3. 在解释器（Evaluator）的源码`eval.go`中加入对`链式比较操作符`的解释。



## 抽象语法树（AST）的更改

对于这种形式的表达式，如何使用抽象语法树（AST: Abstract Syntax Tree）来表示呢？如果读者还记得的话，让我们来看一下中缀表达式的抽象语法表示：

```go
// <left-expression> operator <right-expression>
type InfixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
	Left     Expression
}
```

而`链式比较操作符`比这个中缀表达式多了一个操作符和一个表达式。为了简单起见，我们可以扩展这个`InfixExpression`：

```go
//ast.go
//<left-expression> operator <right-expression> <next-operator> <next-expression>
type InfixExpression struct {
	Token        token.Token
	Operator     string
	Right        Expression
	Left         Expression
    
	HasNext      bool      // Right表达式后，是否紧跟着一个比较操作符
	NextOperator string    // 下一个操作符
	Next         Expression //下一个表达式
}
```

第9-11行是新增的代码。`NextOperator`和`Next`两个字段只有在`HasNext`字段为true的时候才有意义。

有了这个修改后的`InfixExpression`结构，我们还需要修改这个中缀表达式的字符串表示：

```go
//ast.go
//中缀表达式的字符串表示
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())

	if ie.HasNext {
		out.WriteString(" " + ie.NextOperator + " ")
		out.WriteString(ie.Next.String())
	}

	out.WriteString(")")

	return out.String()
}
```

第11-14行是新加入的代码。



## 语法解析器（Parser）的更改

因为我们修改了中缀表达式的抽象语法表示，所以我们还需要在语法解析层面，更改相应的逻辑：

```go
//parser.go
//<left-expression> operator <right-expression> <next-operator> <next-expression>
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	//.....

	if p.isCompareOperator() { //如果下一个词元的类型是比较操作符
		p.nextToken()
		expression.HasNext = true //将HasNext字段设置为true
		expression.NextOperator = p.curToken.Literal //得到当前的操作符

		p.nextToken()
		expression.Next = p.parseExpression(precedence) //解析<next-expression>
	}

	 //继续判断下一个词元的类型。目前我们最多只支持两级比较操作符级联
	if p.isCompareOperator() {
		msg := fmt.Sprintf("Syntax Error:%v- too much comare operator", p.peekToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.peekToken.Pos.Sline())
		return nil
	}

	return expression
}

//判断下一个词元类型是否是比较操作符
func (p *Parser) isCompareOperator() bool {
	return p.peekTokenIs(token.TOKEN_LT) || p.peekTokenIs(token.TOKEN_LE) || // <, <=
		p.peekTokenIs(token.TOKEN_GT) || p.peekTokenIs(token.TOKEN_GE) ||    // >, >=
		p.peekTokenIs(token.TOKEN_EQ) || p.peekTokenIs(token.TOKEN_NEQ)      // ==, !=
}
```

12-27行是新增的代码。我们在原来解析中缀表达式代码的基础上，加入了判断下一个操作符的逻辑，如果`<right-expression>`的下一个操作符是比较操作符，则取得下一个比较操作符的值`NextOperator`（代码第15行)，并解析`<next-expression>`，将其结果赋值给`Next`字段(代码第18行)。`isCompareOperator()`函数用来判断下一个词元类型是否为比较操作符，如果是则返回true。



## 解释器（Evaluator）的更改

在进行解释器的修改之前，让我们再来温习一下`链式比较操作符`的形式：

```xml
#例：   a                <              b               < =                  c
<left-expression> <compare-op> <right-expression> <next-compare-op> <next-expression>
```

对于`a < b`这种中缀表达式的解释，我们返回的是TRUE（条件成立）或者FALSE（条件不成立）。而对于`a < b < c`这种形式，我们需要先判断`a < b`，如果条件不成立的话（FALSE），那么我们无需解释`b < c`，直接返回FALSE即可。

如果条件成立的话（TRUE），我们需要接着解释`b < c`，如果条件还是成立的话则返回TRUE，否则返回FALSE。



有了上面的介绍，我们来看一下`evalStringInfixExpression`和`evalNumberInfixExpression`函数的更改：

```go
//eval.go
func evalStringInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*String).String
	rightVal := right.(*String).String

	switch node.Operator {
	case "+":
		return NewString(leftVal + rightVal)
	case "<":
		result := nativeBoolToBooleanObject(leftVal < rightVal)
		return evalNextStringInfix(node, result, right, scope)
	case "<=":
		result := nativeBoolToBooleanObject(leftVal <= rightVal)
		return evalNextStringInfix(node, result, right, scope)
	case ">":
		result := nativeBoolToBooleanObject(leftVal > rightVal)
		return evalNextStringInfix(node, result, right, scope)
	case ">=":
		result := nativeBoolToBooleanObject(leftVal >= rightVal)
		return evalNextStringInfix(node, result, right, scope)
	case "==":
		result := nativeBoolToBooleanObject(leftVal == rightVal)
		return evalNextStringInfix(node, result, right, scope)
	case "!=":
		result := nativeBoolToBooleanObject(leftVal != rightVal)
		return evalNextStringInfix(node, result, right, scope)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalNextStringInfix(node *ast.InfixExpression, result *Boolean, right Object, scope *Scope) Object {
	if !node.HasNext { //如果没有后续的比较操作符，直接返回计算结果
		return result
	}
    if result == TRUE { //如果 a < b 条件成立(result == TRUE)，我们需要继续计算 b < c
		infixExpr := &ast.InfixExpression{Token: node.Token, Operator: node.NextOperator}
		r := Eval(node.Next, scope)
		return evalStringInfixExpression(infixExpr, right, r, scope)
	}
	return FALSE
}
```

`evalNextStringInfix`函数是新增的，这里需要注意的是第37-39行的代码，第37行我们重新构造了一个`InfixExpression`，第38行我们解释`node.Next`， 第39行调用`evalStringInfixExpression`函数来解释这个字符串比较。

`evalNumberInfixExpression`函数的更改和`evalStringInfixExpression`比较相似，所以我就不多做解释了，只列出代码：

```go
//eval.go
func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...

	case "<":
		result := nativeBoolToBooleanObject(leftVal < rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	case "<=":
		result := nativeBoolToBooleanObject(leftVal <= rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	case ">":
		result := nativeBoolToBooleanObject(leftVal > rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	case ">=":
		result := nativeBoolToBooleanObject(leftVal >= rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	case "==":
		result := nativeBoolToBooleanObject(leftVal == rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	case "!=":
		result := nativeBoolToBooleanObject(leftVal != rightVal)
		return evalNextNumberInfix(node, result, right, scope)
	default:
		return newError(node.Pos().Sline(), ERR_INFIXOP, left.Type(), node.Operator, right.Type())
	}
}

func evalNextNumberInfix(node *ast.InfixExpression, result *Boolean, right Object, scope *Scope) Object {
	if !node.HasNext {
		return result
	}
	if result == TRUE {
		infixExpr := &ast.InfixExpression{Token: node.Token, Operator: node.NextOperator}
		r := Eval(node.Next, scope)
		return evalNumberInfixExpression(infixExpr, right, r, scope)
	}
	return FALSE
}
```

通过上面的更改，我们实际上只实现了`比较操作符`链式支持。例如：

```javascript
 1 < 2 <= 3
 "ab" <= "bc" < "cd"
```

但是，对于下面的脚本代码：

```javascript
if 10 % 2 == 0 {
    println(" 10 % 2 == 0")
} else {
    println("BAD")
}
```

实际运行结果会打印`BAD`。这个是什么原因呢？让我来解释一下。对于`10 % 2 == 0`这个链式表达式，它的抽象语法表示形式如下：

```javascript
//   10       %        2            ==         0
<left-expr>  op   <right-expr>   next-op   <next-expr>
```

而我们的解释代码如下：

```go
//eval.go
func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...

	case "%":
		v := math.Mod(leftVal, rightVal)
		return &Number{Value: v}
	//...
	}
}
```

仔细看第9-11行的代码，它计算完`10 % 2 `(即`<left-expr> op <right-expr>`)的结果后，就将其结果返回了（10 % 2的结果当然为0），而并没有处理` == 0`部分。因此上面的脚本代码实际上就变成了如下：

```javascript
if 10 % 2 { //10 % 2的结果为0， 也就是说if的条件变成了`if 0`
    println(" 10 % 2 == 0")
} else {
    println("BAD")
}
```

因此我们还需要处理剩下的部分。处理逻辑也非常简单：

```go
//eval.go
func evalNumberInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*Number).Value
	rightVal := right.(*Number).Value

	switch node.Operator {
	//...

	case "%":
		v := math.Mod(leftVal, rightVal)
		//return &Number{Value: v} //这是原始的旧代码
		n := &Number{Value: v}
		if node.HasNext { //如果有next部分，则构造一个新的`InfixExpression`
			infixExpr := &ast.InfixExpression{Token: node.Token, Operator: node.NextOperator}
			r := Eval(node.Next, scope) //计算next表达式
			return evalNumberInfixExpression(infixExpr, n, r, scope) //调用自身
		}
		return n //没有next部分的情况下，和原来代码一样，返回两个数字取模的结果
	//...
	}
}
```

我们不仅仅要对`%`中缀符进行更改，其它的中缀符，比如`+`、`-`、`*`、`/`、`**`都需要做相应的更改。更改的内容和这里更改的内容几乎是一样的，这里就不列出代码了。



同时，对于字符串相加的情况，我们也需要做更改，更改内容如下：

```go
//eval.go
func evalStringInfixExpression(node *ast.InfixExpression, left, right Object, scope *Scope) Object {
	leftVal := left.(*String).String
	rightVal := right.(*String).String

	switch node.Operator {
	case "+":
		// return NewString(leftVal + rightVal)  //这个是原先的旧代码
		s := NewString(leftVal + rightVal)
		if node.HasNext { //如果有next部分
			infixExpr := &ast.InfixExpression{Token: node.Token, Operator: node.NextOperator}
			r := Eval(node.Next, scope)
			return evalStringInfixExpression(infixExpr, s, r, scope)
		}
		return s
	//...
	}
}
```



## 测试

```javascript
# 数组
fn sum (x,y) { x + y }
println(10 < sum(10, 15) <= 25)

a = 12
if 10 != a < 13 {
    println("10 != a < 13")
}

if 10 % 2 == 0 {
    println("10 % 2 == 0")
}
```



下一节，我们会加入`in`操作符的支持。



