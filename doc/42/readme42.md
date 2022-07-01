# `变参函数(Variadic Functions)`支持

这一节中，我们将加入`变参函数(Variadic Functions)`的支持。所谓的变参函数，就是函数支持可变参数。先来看一下使用例：

```go
fn add(x, y, args...) {
    w = x + y
    for i in args {
        w = w + i
    }
    return w
}

sum1 = add(2,3,4,5)
println(sum1)

sum2 = add(2,3,4,5,6,7)
println(sum2)
```

第一行`add`函数的最后一个参数`args...`就表示这个函数可以接受一个可变参数。这里需要注意的一点就是可变参数必须是函数的最后一个参数。

我们再看下一个例子：

```go
fn _add(x, args...) {
    w = x
    for i in args {
        w = w + i
    }
    return w
}

fn add(x, y, args...) {
    return _add(x+y, args...)
}

println(add(1, 2, 3, 4, 5))
```

从第10行的代码可以看到，我们不仅可以在函数的声明中使用`...`，函数调用的地方我们也可以使用`...`。



下面看一下我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（Token）类型(`...`)
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`...`的识别
2. 在抽象语法树（AST）源码`ast.go`中修改`函数字面量`和`函数调用`的抽象语法表示。
3. 在语法解析器（Parser）的源码`parser.go`中修改对`函数字面量`和`函数调用`的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中修改对`函数字面量`和`函数调用`的解释。



## 词元（Token）的更改

```go
//token.go
const (
	//...

	TOKEN_ELLIPSIS   //...

	//...
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...

	case TOKEN_ELLIPSIS:
		return "..."

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

	case '.':
		if l.peek() == '.' {
			l.readNext()
			if l.peek() == '.' {
				tok = token.Token{Type: token.TOKEN_ELLIPSIS, Literal: "..."}
				l.readNext()
			} else {
				tok = token.Token{Type: token.TOKEN_DOTDOT, Literal: ".."}
			}
		} else {
			tok = newToken(token.TOKEN_DOT, l.ch)
		}

	//...
	}

	//...
}
```

第8-19行的`case`分支是修改及新增的代码。



## 抽象语法树(AST)的更改

从本节开头的第二个例子中可以看到，函数声明和函数调用都可以使用`...`。因此我们需要更改`函数字面量`和`函数调用`的抽象语法表示。

先来看一下，对`函数字面量`的抽象语法表示的更改：

```go
//ast.go
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Variadic   bool
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
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	if fl.Variadic {
		out.WriteString("...")
	}
	out.WriteString(") {")
	out.WriteString(fl.Body.String())
	out.WriteString("}")

	return out.String()
}
```

`函数字面量`的抽象语法表示结构中，我们增加了一个`Variadic`的布尔型的变量（代码第5行）。同时，在`函数字面量`的字符串表示中，我们加入了相关的逻辑（代码21-23行）。



接下来，我们看一下函数调用的更改：

```go
//parser.go
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
	Variadic  bool
}

//函数调用的字符串表示
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	if ce.Variadic {
		out.WriteString("...")
	}
	out.WriteString(")")
	return out.String()
}
```

同样的，我们在`函数调用`的抽象语法表示结构中，加入了一个`Variadic`的布尔型的变量（代码第6行）。在`函数调用`的字符串表示中，我们加入了相关的逻辑（代码21-23行）。



## 语法解析器（Parser）的更改

从本节开头的第二个例子中可以知道，我们需要更改`函数字面量`和`函数调用`的语法解析，追加对`...`参数的解析。

在更改之前，让我们写一个简单的函数，这个函数用来检查`...`参数是否为函数或者调用的最后一个参数，不是的话，就报错并返回失败：

```go
//parser.go
//第一个返回的布尔值表示是否找到了'...'，找到返回true，没找到返回false
//第二个返回的布尔值表示是否成功或者失败。如果'...'不是最后一个参数，则返回失败（false），否则返回成功（true）
func (p *Parser) checkEllipsis() (bool, bool) {
	gotEllipsis := false
	if p.peekTokenIs(token.TOKEN_ELLIPSIS) {
		gotEllipsis = true
		p.nextToken()
		if !p.peekTokenIs(token.TOKEN_RPAREN) {
			msg := fmt.Sprintf("Syntax Error:%v- can only have '...' after last parameter", 
                               p.curToken.Pos)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return false, false
		}
	}
	return gotEllipsis, true
}
```

第6行的`if`判断下一个词元类型是否为`...`。如果是的话，则继续判断`...`后是否跟着的是一个`)`，如果不是跟着`)`，表明可变参数不是最后一个参数，则报错并返回。

接下来来看一下，我们对`函数字面量`的语法解析的更改

```go
//parser.go
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.TOKEN_LPAREN) {
		return nil
	}
	lit.Parameters, lit.Variadic = p.parseFunctionParameters()
	if !p.expectPeek(token.TOKEN_LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, bool) {
	gotEllipsis := false
	success: = false

	identifiers := []*ast.Identifier{}
	if p.peekTokenIs(token.TOKEN_RPAREN) {
		p.nextToken()
		return identifiers, false
	}
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)
	gotEllipsis, success = p.checkEllipsis() //e.g. fn xxx(args...)
	if !success {
		return nil, false
	}
    
	for p.peekTokenIs(token.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
		gotEllipsis, success = p.checkEllipsis()
		if !success {
			return nil, false
		}
	}

	if !p.expectPeek(token.TOKEN_RPAREN) {
		return nil, false
	}
	return identifiers, gotEllipsis
}
```

代码看起来有点多（我贴了完整的函数代码），但是实际上增加的代码并不是很多。首先是代码第7行，我们给`parseFunctionParameters`函数增加了一个布尔型的返回值（表示函数是否有可变参数），这个函数的第27-30和第37-40行是新追加的代码。



接着来看一下，我们对`函数调用`的语法解析的更改：

```go
//parser.go
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments, exp.Variadic = p.parseExpressionList(token.TOKEN_RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) ([]ast.Expression, bool) {
	gotEllipsis := false
    success := false

    list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list, false
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	gotEllipsis, success = p.checkEllipsis() //e.g. call(args...)
	if !success {
		return nil, false
	}

	for p.peekTokenIs(token.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))

		gotEllipsis, success = p.checkEllipsis()
		if !success {
			return nil, false
		}
	}

	if !p.expectPeek(end) {
		return nil, false
	}

	return list, gotEllipsis
}

```

第4行，我们给`parseExpressionList`函数增加了一个布尔型的返回值（表示函数是否有可变参数），这个函数的第20-23及30-33行是新追加的代码。是不是和`parseFunctionLiteral`的更改很类似？

由于我们给`parseExpressionList`函数增加了一个布尔型的返回值，而`parseArrayLiteral`函数也会调用`parseExpressionList`函数，所以我们也需要对其做个小小的更改：

```go
//parser.go
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Members, _ = p.parseExpressionList(token.TOKEN_RBRACKET)
	return array
}
```

代码第4行，我们加入了一个占位符`_`，因为数组字面量用不到第二个参数，我们也并不关心它。



## 解释器（Evaluator）的更改

在修改解释器的代码之前，我们还是来看一下本节开头的第二个例子：

````go
fn _add(x, args...) {
    w = x
    for i in args {
        w = w + i
    }
    return w
}

fn add(x, y, args...) {
    return _add(x+y, args...)
}

println(add(1, 2, 3, 4, 5))
````

从第3行的`for`循环可以看到，函数内部，我们是将可变参数（`args...`)当成了一个数组来处理。就是说我们需要将传递的可变参数打包成一个数组。拿上面的例子来说明，就是我们在第13行给`add`函数传递了5个值，我们需要将其当成3个值来传给`add`函数，即`1`、`2`和`[3,4,5]`。这称为装箱(Boxing)。

对于第10行的`_add`函数调用，我们需要将`args...`参数拆开成一个个独立的参数。就是说将`args...`（即`[3,4,5]`）解开。这称之为拆箱(Unboxing)。为什么这里要拆箱呢？其实很好理解，我们不能传递一个不确定的参数给函数，对吧（简单点说，就是函数的实参个数必须确定）。其实就是将第10行的`_add(1,2,[3,4,5]...)`函数调用变成`_add(1,2,3,4,5)`这样的普通函数调用，就像第13行对函数`add`的调用一样。

如果理解了上面的说明，那么我们就可以着手修改相关的解释代码了。先来看一下对上面例子的第10行的拆箱更改：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
	args := evalExpressions(node.Arguments, scope)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	if node.Variadic {
		args = getVariadicArgs(node, args)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
	}

	//...

	function := Eval(node.Function, scope)
	if isError(function) {
		return function
	}

	return applyFunction(node.Pos().Sline(), scope, function, args)
}

func getVariadicArgs(call *ast.CallExpression, args []Object) []Object {
	lastArg := args[len(args)-1] //取最后一个参数
	iterObj, ok := lastArg.(Iterable) //判断最后一个参数是否可遍历
	if !ok {
		errObj := newError(call.Pos().Sline(), ERR_NOTITERABLE)
		return []Object{errObj}
	}
	if !iterObj.iter() {
		errObj := newError(call.Pos().Sline(), ERR_NOTITERABLE)
		return []Object{errObj}
	}

	var members []Object
    //判断最后一个参数的类型(实际上就是取可变参数)
	if lastArg.Type() == STRING_OBJ {
		aStr, _ := lastArg.(*String)
		runes := []rune(aStr.String)
		for _, rune := range runes {
			members = append(members, NewString(string(rune)))
		}
	} else if lastArg.Type() == ARRAY_OBJ {
		arr, _ := lastArg.(*Array)
		members = arr.Members
	} else if lastArg.Type() == TUPLE_OBJ {
		tuple, _ := lastArg.(*Tuple)
		members = tuple.Members
	} else if lastArg.Type() == GO_OBJ { //go object
		goObj := lastArg.(*GoObject)
		arr := goValueToObject(goObj.obj).(*Array)
		members = arr.Members
	}

	args = args[:len(args)-1] //取可变参数前的固定参数
    for _, m := range members { //将可变参数一个一个加入的args数组中
		args = append(args, m)
	}

	return args
}

```

第8-13行的`if`分支是新追加的代码。我们调用了一个新追加的函数（`getVariadicArgs`）。当函数调用含有`...`，我们就进行拆箱操作，将最后一个可变参数（实际上在函数内部是数组）拆成一个个独立的参数。这里重点关注57-60行的代码。第57行取可变参数前面的固定参数，第58-59行将拆箱的一个个参数附加到`args`数组中。



接下来看一下对上面例子的第13行`add(1, 2, 3, 4, 5)`的装箱（Boxing）更改。即将`3`、`4`、`5`装箱成一个`[3,4,5]`数组。其实就是更改`applyFunction`函数的内部实现：

```go
//eval.go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function:
		extendedScope := extendFunctionScope(fn, args)
		evaluated := Eval(fn.Literal.Body, extendedScope)
		return unwrapReturnValue(evaluated)
	case *Builtin:
		return fn.Fn(line, scope, args...)
	default:
		return newError(line, ERR_NOTFUNCTION, fn.Type())
	}
}
```

我们需要修改的是上面代码第5行的`extendFunctionScope`函数（即在第6行解释函数体之前完成装箱操作）：

```go
//eval.go
func extendFunctionScope(fn *Function, args []Object) *Scope {
	scope := NewScope(fn.Scope, nil)
	if fn.Literal.Variadic {
		ellipsisArgs := args[len(fn.Literal.Parameters)-1:]
		newArgs := make([]Object, 0, len(fn.Literal.Parameters)+1)
		newArgs = append(newArgs, args[:len(fn.Literal.Parameters)-1]...)
		args = append(newArgs, &Array{Members: ellipsisArgs})
		for i, arg := range args {
			scope.Set(fn.Literal.Parameters[i].String(), arg)
		}
	} else {
		for paramIdx, param := range fn.Literal.Parameters {
			scope.Set(param.Value, args[paramIdx])
		}
	}
	return scope
}
```

第4-12行是新增的`if`分支。如果函数声明中包含`...`，我们就进行装箱操作。第5行将可变参数变成一个数组（严格的说是分片，即array slice）。第6-7行实际是将固定参数取出。这里需要注意的是第8行，我们将固定参数（`newArgs`）和可变参数（`ellipsisArgs`）重新组装（即装箱）成一个新的数组。这里的重点是我们将可变参数（`ellipsisArgs`）放到了一个数组对象中（这就完成了装箱操作）。



我们还需要更改所有函数调用之前，解析参数的地方，即`evalExpressions`函数（它仅仅处理固定参数）的后面，加上处理可变参数的代码，像下面这样：

```go
args := evalExpressions(method.Arguments, scope)
if len(args) == 1 && isError(args[0]) {
	return args[0]
}

if method.Variadic {
	args = getVariadicArgs(method, args)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}
}
```

类似上面代码的6-10行，就是我们需要追加的代码。程序中有好几处需要追加这种逻辑，所以这里就不一一列出代码了，读者可以参看本节的源码。



## 测试

```javascript
plus = fn(nums...) {
    sum = 0
    for n in nums {
        sum = sum + n
    }
    return sum
}

println(plus(1, 2, 3))

lst = [4, 5, 6]
println(plus(lst...))
```



下一节，我们会加入`复合赋值运算符`的支持。



