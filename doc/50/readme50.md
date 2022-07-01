# `管道操作符（|>）`支持

这一节中，我们会将增加类似`Elixir`语言的管道操作符（`|>`） 的支持。先来看一个`管道操作符`的例子：

```go
fn uppercase(s) {
    return s.upper()
}

upper_str1 = "hello, world" |> uppercase
println(upper_str1)

upper_str2 = "hello, world" |> uppercase()
println(upper_str2)
```

第5行（`uppercase`是一个标识符）和第8行（`uppercase()`是个函数调用）的作用是一样的，在我们的实现中，这两种方式都支持。不知道读者是否看出了`管道操作符`的作用？实际上`管道操作符`就是将`|>`左边表达式的结果，作为右边表达式的第一个参数。当然这里右边的表达式必须是一个函数调用或者方法调用（其实还可以是上面例子第2行中的标识符，而这个标识符代表一个函数）。上面的例子是一个函数调用的例子。我们再来看一下方法调用的例子：

```go
struct demo {
    fn Uppercase(s) {
        return s.upper()
    }
}

demo_struct = demo()
upper_str1 = "hello, world" |> demo_struct.Uppercase
println(upper_str1)

upper_str2 = "hello, world" |> demo_struct.Uppercase()
println(upper_str2)
```

第8行和第11行的就是方法调用的例子。

最后让我们看一下在什么情况下使用管道操作符是比较好的（所谓的Best Practice）。假设我们有下面的函数调用：

```go
foo(bar(baz(new_function(other_function()))))
```

对于这样的函数调用来说，是不是看起来很别扭。如果换成`管道操作符`的话，可读性就比较强了：

```
other_function() |> new_function() |> baz() |> bar() |> foo()
```

就是说，对于数据转换（Data Transformation）这种需求来说，使用`管道操作符`是个不错的选择。



有了上面的简短介绍，接下来来看我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入新的词元（`|>`）
2. 在词法分析器（Lexer）源码`lexer.go`中加入对`管道操作符`的识别
3. 在语法解析器（Parser）的源码`parser.go`中给`管道操作符`注册中缀表达式回调函数。
4. 在解释器（Evaluator）的源码`eval.go`中加入对`管道操作符`的解释。



## 词元（Token）的更改

只列出代码，不做解释：

```go
//token.go
const (
	TOKEN_PIPE     // |>
)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_PIPE:
		return "|>"
	}
}
```



## 词法分析器（Lexer）的更改

同样，这里只列出代码，不多做解释：

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '|':
		if l.peek() == '|' {
			tok = token.Token{Type: token.TOKEN_OR, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		} else if l.peek() == '>' {
			tok = token.Token{Type: token.TOKEN_PIPE, Literal: string(l.ch) + string(l.peek())}
			l.readNext()
		}

	//...
	}
}
```



## 语法解析器（Parser）的更改

更改之前，我们再来看一下管道操作符`|>`的一般形式：

```
<left-expression> |> <right-expression>
```

读者可能已经猜到了，管道操作符`|>`是一个中缀操作符。既然是操作符，我们就需要给其赋优先级。还要给这个管道操作符`|>`注册中缀表达式回调函数。对于注册中缀表达式回调函数，这个非常简单：

```go
//parser.go
func (p *Parser) registerAction() {
	//...
	p.registerInfix(token.TOKEN_PIPE, p.parseInfixExpression)
	//...
}
```

而我们给其赋什么优先级呢？这个我参照了`Elixir`语言的操作符优先级，给这个操作符赋予和比较操作符同等的优先级：

```go
//parser.go
var precedences = map[token.TokenType]int{
	//...
	token.TOKEN_LT:   LESSGREATER,
	token.TOKEN_LE:   LESSGREATER,
	token.TOKEN_GT:   LESSGREATER,
	token.TOKEN_GE:   LESSGREATER,
	token.TOKEN_IN:   LESSGREATER,
	token.TOKEN_PIPE: LESSGREATER,
}
```

其中第9行是新增的代码。



## 解释器（Evaluator）的更改

对于管道操作符（`|>`），我们需要在`Eval`函数的`case *ast.InfixExpression`分支中加入处理这个操作符的逻辑代码：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	//...
	switch node := node.(type) {
	//...
	case *ast.InfixExpression:
		if node.Operator == "|>" {
			return evalPipeInfix(node, scope)
		}
	//...
	}
}
```

`evalPipeInfix`函数是实际的处理逻辑，让我们来看一下它的实现：

```go
//eval.go
func evalPipeInfix(node *ast.InfixExpression, scope *Scope) Object {
	switch rightFunc := node.Right.(type) {
	case *ast.MethodCallExpression: //例如：【"x" |> obj.meth】或者【"x" |> obj.meth()】
		switch rightFunc.Call.(type) {
		case *ast.Identifier:
			//e.g.
			//x = "hello" |> xxx.upper    : rightFunc.Call.(type) == *ast.Identifier
			//x = "hello" |> xxx.upper()  : rightFunc.Call.(type) == *ast.CallExpression
			//将*ast.Identifier转换为一个*ast.CallExpression
			rightFunc.Call = &ast.CallExpression{Token: node.Token, 
												 Function: rightFunc.Call}
		}
		//将left作为right的第一个参数
		rightFunc.Call.(*ast.CallExpression).Arguments = 
		    append([]ast.Expression{node.Left}, 
				   rightFunc.Call.(*ast.CallExpression).Arguments...)
		return Eval(rightFunc, scope)

	case *ast.CallExpression: //例如："hello" |> uppercase()
		//将left作为right的第一个参数
		rightFunc.Arguments = append([]ast.Expression{node.Left}, rightFunc.Arguments...)
		return Eval(rightFunc, scope)

	case *ast.Identifier: //例如："hello" |> uppercase
		right := Eval(node.Right, scope) //解释右表达式
		if isError(right) {
			return right
		}
		if right.Type() == FUNCTION_OBJ { //如果是个函数
			//手动构造一个*ast.CallExpression表达式
			call := &ast.CallExpression{Token: node.Token, Function: rightFunc, 
										Variadic: right.(*Function).Literal.Variadic}
			//将节点的left表达式作为right表达式的第一个参数
			call.Arguments = append([]ast.Expression{node.Left}, call.Arguments...)
			return evalCallExpression(call, right, scope)
		} else {
			return newError(node.Pos().Sline(), ERR_PIPE)
		}
	}

	return NIL
}
```

需要注意的是，如果右边的表达式是个标识符（identifier）的话，即类型是`*ast.Identifier`，我们需要手动创建一个`*ast.CallExpression`（代码第11行和第32行）。

上面代码的第38行有个`ERR_PIPE`变量是新增的，它定义在`errors.go`中：

```go
//errors.go
var (
	//...
	ERR_PIPE = "pipe operator's right hand side is not a function"
)
```



## 测试

```javascript
struct demo {
    fn Uppercase(s) {
        return s.upper()
    }
}

fn uppercase(s) {
    return s.upper()
}

//函数调用
upper_str1 = "hello, world" |> uppercase
println(upper_str1)

upper_str2 = "hello, world" |> uppercase()
println(upper_str2)

//方法调用
demo_struct = demo()
upper_str1 = "hello, world" |> demo_struct.Uppercase()
println(upper_str1)

upper_str2 = "hello, world" |> demo_struct.Uppercase
println(upper_str2)


```



下一节，我们将讨论对`有序哈希（ordered hash）`的支持。



