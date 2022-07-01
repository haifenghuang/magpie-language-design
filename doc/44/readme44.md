# 真正意义上的多重赋值

在第25节中，我们实现了多重赋值，实际上那一节的实现可以称为`多重变量声明赋值`。因此它不支持下面的例子中的`多重赋值`：

```go
arr = [1,2,3]
arr[0], arr[1] = arr[1], arr[0]

hash = {"a": 10, "b": "hello"}
hash["a"], hash["b"] = hash["b"], hash["a"]

a, arr[2] = 100, 10

struct math {
	let a = 10
	let b = 20
	fn Swap() {
		self.a, self.b = self.b, self.a
        printf("a = %d, b = %d\n", self.a, self.b)
	}
}
```

在第25节中，我们将`多重赋值`实现为一个`LetStatement`。对于`Let`语句来说，因为只是变量声明（也可以同时给变量赋初值），所以它只支持等号（`=`）左边全部是变量的情况（其实还支持一个`_`占位符）。也就是说对于上面例子中的第2、5、7及13行，它是无法解析的。在这一节中，我们将实现对上面例子中的第2、5、7及13行的支持。



下面看一下我们需要做哪些更改：

1. 在抽象语法树（AST）源码`ast.go`中一个新的`多重赋值`语法表示。
2. 在语法解析器（Parser）的源码`parser.go`中加入对多重赋值的的解析。
4. 在解释器（Evaluator）的源码`eval.go`中加入对这个新的`多重赋值`的解释。



## 抽象语法树（AST）的更改

如果读者还有印象的话，我们的`AST`表示中，已经有了一个`单重赋值`的抽象语法表示（`即AssignExpression`）：

```go
//ast.go
type AssignExpression struct {
	Token token.Token
	Name  Expression
	Value Expression
}
```
它的`Name`和`Value`都是一个表达式。而我们现在要支持真正意义上的多重赋值，所以这里的`Name`和`Value`需要是一个数组。有的读者可能说，那我们将其更改为数组不就可以了吗？像下面这样：

```go
//ast.go
type AssignExpression struct {
	Token token.Token
	Names  []Expression
	Values []Expression
}
```

这样是没错。但是这样变更的话有一个问题，就是上一节中讲解的`复合赋值运算符`是不适用的。看一下下面的例子就明白了：

```go
a = 10
a += 10
```

对于第2行的这种`复合赋值运算符`（这里是`+=`），我们还是需要使用没有更改过的`单重赋值(AssignExpression)`表达式。因此这节我们需要新增一个`多重赋值`的抽象语法表示，专门用来处理多重赋值的情况：

```go
//ast.go
type MultiAssignStatement struct {
	Token  token.Token
	Names  []Expression
	Values []Expression
}

func (as *MultiAssignStatement) Pos() token.Position {
	return as.Token.Pos
}

func (as *MultiAssignStatement) End() token.Position {
	aLen := len(as.Values)
	if aLen > 0 {
		return as.Values[aLen-1].End()
	}

	return as.Values[0].End()
}

func (as *MultiAssignStatement) statementNode()       {}
func (as *MultiAssignStatement) TokenLiteral() string { return as.Token.Literal }

func (as *MultiAssignStatement) String() string {
	var out bytes.Buffer

	names := []string{}
	for _, name := range as.Names {
		names = append(names, name.String())
	}
	out.WriteString(strings.Join(names, ", "))

	out.WriteString(" = ")

	values := []string{}
	for _, value := range as.Values {
		values = append(values, value.String())
	}
	out.WriteString(strings.Join(values, ", "))

	return out.String()
}
```

实际上就是把`AssignExpression`结构中的`Name`和`Value`变成了数组。



## 语法解析器（Parser）的更改

在第25节中，对于`多重赋值`，我们有如下的代码修改：

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.TOKEN_LET:
		return p.parseLetStatement(true)
	//...
	case token.TOKEN_IDENTIFIER:
		if p.peekTokenIs(token.TOKEN_COMMA) {
			return p.parseLetStatement(false)
		}
		fallthrough
	default:
		return p.parseExpressionStatement()
	}
}
```

这一节中，我们需要更改第7行的`case`分支。上面的代码第7行，如果遇到了标识符，再接着判断标识符的下一个词元类型是否为`逗号`。如果是，就认为是一个`多重变量声明赋值`语句。而这一节我们需要更改这种判断。如果遇到标识符，我们需要先解析一下，然后再判断解析后的词元类型是否为`逗号`：

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
	case token.TOKEN_IDENTIFIER:
		stmt := p.parseExpressionStatement()
		if p.peekTokenIs(token.TOKEN_COMMA) {
			return p.parseMultiAssignStatement(stmt.Expression)
		}
		return stmt
	default:
		return p.parseExpressionStatement()
	}
}
```

第6行，我们调用`parseExpressionStatement`函数来解析表达式语句，这个函数就能够正确的解析类似`arr[i]`、`hash[a]`之类的表达式（这些都被归并到`ExpressionStatement`）。解析完成后，我们接着判断下一个词元类型是否为`逗号`，如果是，则调用`parseMultiAssignStatement`函数来处理多重赋值语句的解析。如果不是则直接返回（代码第10行）。

现在再来看一下`parseMultiAssignStatement`函数的实现：

```go
//parser.go
func (p *Parser) parseMultiAssignStatement(expr ast.Expression) *ast.MultiAssignStatement {
	tok := token.Token{Pos: p.curToken.Pos, Type: token.TOKEN_ASSIGN, Literal: "="}
	stmt := &ast.MultiAssignStatement{Token: tok}

	stmt.Names = append(stmt.Names, expr) //先把第一个表达式加入Names
	p.nextToken()
	p.nextToken()

	//解析剩余的names
	for {
		n := p.parseExpression(ASSIGN)
		stmt.Names = append(stmt.Names, n)
		if p.peekTokenIs(token.TOKEN_ASSIGN) { //如果下一个词元类型是'='则break
			p.nextToken()
			p.nextToken()
			break
		}
		if !p.peekTokenIs(token.TOKEN_COMMA) { //如果下一个词元类型不是'，'则break
			break
		}

		p.nextToken()
		p.nextToken()
	}

	//解析values
	for {
		v := p.parseExpressionStatement().Expression
		stmt.Values = append(stmt.Values, v)

		if !p.peekTokenIs(token.TOKEN_COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	//fmt.Printf("MultiAssignStatement=%s\n", stmt)
	return stmt
}
```

这里需要特别注意的是第12行。我们给`parseExpression`函数传递的优先级是`ASSIGN`（即赋值语句的优先级）。这个非常重要，如果我们传递的优先级是`LOWEST`的话，语法解析就会解析错误。

细心的读者可能会说，我们是不是需要处理一下左边的`Names`及其等号右边的`Values`的个数是否一致的情况？说实话，在语法解析阶段，在一些特殊的情况下，我们还不能准确的判断等号右边的`Values`的个数。举个例子吧：

```go
fn math(x,y) {
	return x + y, x - y
}
sum_result, sub_result = math(10, 2)
```

就是说如果等式右边有函数（这个函数返回多个值）的情况，我们在解析阶段就无能为力了。我们会在代码`解释(Evaluating)`阶段来处理这种情况。

上面的代码其实主要就是解析等号左边的`Names`及其等号右边的`Values`。

解析器的代码变更就这些。是不是比想象的简单？

哦。忘记了，我们的`parseLetStatemnt`函数需要变回原来的样子，具体的更改请参照源码。这里就不列出了。

> 当然，你也完全可以不用更改`parseLetStatemnt`函数。



## 解释器（Evaluator）的更改

由于我们新加入了一个`MultiAssignStatement`抽象语法表示，因此我们需要在`Eval`函数中新增一个`case`分支：

````go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	//...
	
	switch node := node.(type) {
	//...
	case *ast.MultiAssignStatement:
		return evalMultiAssignStatement(node, scope)
	}
}
````

在实现`evalMultiAssignStatement`函数之前，让我们先对`evalAssignExpression`函数做一下小小的修改：

```go
//eval.go
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	val := Eval(a.Value, scope)
	if val.Type() == ERROR_OBJ {
		return val
	}

    //一堆复杂的代码
}

```

我们需要将上面所谓的`一堆复杂的代码`提取出来，生成一个新的函数`_evalAssignExpression`。这样`evalAssignExpression`函数就变成了下面这样：

```go
//eval.go
func evalAssignExpression(a *ast.AssignExpression, scope *Scope) Object {
	val := Eval(a.Value, scope)
	if val.Type() == ERROR_OBJ {
		return val
	}

	return _evalAssignExpression(a, val, scope)
}
```

这样做的目的是什么呢？主要是在`evalMultiAssignStatement`函数的实现中，利用这个`_evalAssignExpression`函数。

下面让我们来看一下`evalMultiAssignStatement`函数的实现：

```go
//eval.go
func evalMultiAssignStatement(ma *ast.MultiAssignStatement, scope *Scope) Object {
	values := []Object{}
	valuesLen := 0

	for _, value := range ma.Values { //解释`Values`，并判断`Values`的个数
		val := Eval(value, scope)
		if val.Type() == ERROR_OBJ {
			return val
		}

		//元祖的情况需要特殊处理（因为对于返回多个值的函数，我们将返回值放在一个元祖中）
		if val.Type() == TUPLE_OBJ {
			tupleObj := val.(*Tuple)
			if tupleObj.IsMulti { //是一个返回多个值的函数
				valuesLen += len(tupleObj.Members)
				values = append(values, tupleObj.Members...)
			} else { //是一个实际的元祖(tuple)
				valuesLen += 1
				values = append(values, tupleObj)
			}
		} else { //不是元祖的情况
			valuesLen += 1
			values = append(values, val)
		}
	}

	//为了下面的处理简便起见，这里判断`Names`和`Values`的个数是否相等，不相等的话则报错。
	if len(values) != len(ma.Names) {
		return newError(ma.Pos().Sline(), ERR_MULTIASSIGN)
	}

	for idx, name := range ma.Names {
        if name.TokenLiteral() == "_" { // 如果是占位符(_)则继续
			continue
		}

		//为了利用`_evalAssignExpression`,这里收到生成了一个`AssignExpression`结构
		a := &ast.AssignExpression{Token: ma.Token, Name: name}
		_evalAssignExpression(a, values[idx], scope)
	}

	return NIL
}
```

第30行我们新增了一个`ERR_MULTIASSIGN`，我们需要在`errors.go`中新增一个：

```go
//errors.go
var (
	//...
	ERR_MULTIASSIGN     = "the number of names and values are not equal"
)
```

解释器的更改也就完成了。下面让我们简单测试一下。



## 测试

```javascript
//数组
arr = [10, 12]
arr[0] = 25
printf("arr before: %s\n", arr)
arr[0], arr[1] = arr[1], arr[0]
printf("arr after: %s\n", arr)

//哈希
hash = {"a": 10, "b": "hello"}
printf("hash before: %s\n", hash)
hash["a"], hash["b"] = hash["b"], hash["a"]
printf("hash after: %s\n", hash)

//带有占位符(_)的情况
a, bc, _, d = 12, "hello", "world", [1,2,3]
printf("a=%d, bc=%s, d=%s\n", a, bc, d)

//函数有多个返回值的情况
fn math(x, y) {
   return x + y, x -y
}

arr2=[110, 200]
arr2[0], sub_result = math(10, 2)
printf("arr2=%s, sub_result=%d\n", arr2, sub_result)

//简单的交换两个变量的值
w1 = 10
w2 = 20
w1, w2 = w2, w1
printf("w1=%d, w2=%d\n", w1, w2)

```



下一节我们加入尾调用优化（Tail Call Optimization）的支持。
