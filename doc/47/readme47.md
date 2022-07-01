# `函数装饰器（Decorator）`支持

这一节中，我们会给`magpie`语言加入类似`python`语言的`函数装饰器（Decorator）`的功能，当然我们这里实现的`函数装饰器（Decorator）`远没有`python`的功能强大。我们先看一个`函数装饰器（Decorator）`的例子：

```go
//装饰器函数
fn log(otherfn) {
    return fn() {
        println("otherfn start")
        otherfn($_)
        println("otherfn end")
    }
}

@log //使用【@装饰器函数名】来声明一个装饰器
fn sum(x, y) { //sum函数是被装饰器函数修饰的函数，即【被装饰函数】
    printf("%d + %d = %d\n", x, y, x+y)
}

sum(1,2)
/* 结果：
otherfn start
1 + 2 = 3
otherfn end
*/
```

对于没有写过`python`语言的读者，可能对`装饰器`这个术语比较陌生。这里简单解释一下：

> 装饰器（Decorator）本质上是一个函数（例如上面例子中的`log`函数），作用是给其他函数（例如上例中的`sum`函数）添加附加功能。

如果上例中的`sum`函数没有装饰器的话，第15行运行后应该仅输出`1 + 2 = 3`。有了`log`这个装饰器（第10行），它相当于给`sum`函数添加了功能。

在实现这个`函数装饰器（Decorator）`的功能之前，需要明确一下，这里实现的装饰器有一定的限制：

1. 不支持装饰器所装饰的函数（上例中的`sum`函数）有可变参数，就是说`sum`函数不可以有可变参数。
2. 装饰器函数（上例中的`log`函数）中调用被装饰的函数的语法必须是`functionName($_)`这种形式（上例中的第5行），即括号中有且仅有一个参数，而且参数名必须是`$_`。这里`$_`表示传递给被装饰函数的所有参数。



有了上面的讲解，现在让我们开始着手实现装饰器。还是老一套，先来看一下我们需要做哪些更改：

1. 在词元（Token）源码`token.go`加入新增的词元类型（`$`）。
2. 在词法分析器（Lexer）源码`lexer.go`加入对`$`的词法分析。
3. 在抽象语法树(AST)源码`ast.go`中加入装饰器的抽象语法表示。
4. 在语法解析器（Parser）的源码`parser.go`中实现装饰器的语法解析。
5. 在解释器（Evaluator）的源码`eval.go`中加入对装饰器的解释。



## 词元（Token）的更改

```go
//token.go

const (
	//...
	TOKEN_AT // @
)

//词元的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_AT:
		return "@"
	}
}
```



## 词法解析器（Lexer）的更改

```go
//lexer.go
func (l *Lexer) NextToken() token.Token {
	//...

	switch l.ch {
	//...
	case '@':
		tok = newToken(token.TOKEN_AT, l.ch)
	}
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}
```

需要强调的是第13行的代码，在`isLetter`函数中，我们增加了`ch == '$'`这个判断。因为本节开始的例子中，我提过使用`$_`来表示传给【被装饰函数】的所有参数。



## 抽象语法树（AST）的更改

从本节开始的例子中，我们可以看出，装饰器的形式如下：

```
@decorator1
@decorator2(1)
fn xxx() {}
```

这里要提及的有亮点：

1. `xxx`函数可以有多个装饰器
2. 装饰器因为是函数，所以第2行我们可以给装饰器`decorator2`传参数。

来看一下装饰器的抽象语法表示：

```go
//ast.go
//@Func DecoratedFunction
//例如 @logger fn demo(xx, xx) {}
type DecoratorExpr struct {
	Token     token.Token // '@'
	Decorator Expression  //装饰器函数
	Decorated Expression  //装饰器修饰的函数或者另一个装饰器
}

func (dc *DecoratorExpr) Pos() token.Position {
	return dc.Token.Pos
}

func (dc *DecoratorExpr) End() token.Position {
	return dc.Decorated.End()
}

func (dc *DecoratorExpr) expressionNode()      {}
func (dc *DecoratorExpr) TokenLiteral() string { return dc.Token.Literal }
func (dc *DecoratorExpr) String() string {
	var out bytes.Buffer

	out.WriteString("@")
	out.WriteString(dc.Decorator.String())
	out.WriteString(" ")
	out.WriteString(dc.Decorated.String())

	return out.String()
}
```



## 语法解析器（Parser）的更改

首先，我们需要给`@`符增加一个前缀回调函数：

```go
//parser.go
func (p *Parser) registerAction() {
	//...

	p.registerPrefix(token.TOKEN_AT, p.parseDecorator)
}
```

我们给`@`符号增加了一个前缀回调函数`parseDecorator`。来看一下它的实现：

```go
//parser.go
//
func (p *Parser) parseDecorator() ast.Expression {
	dc := &ast.DecoratorExpr{Token: p.curToken}
	p.nextToken() //skip the '@'
	dc.Decorator = p.parseExpressionStatement().Expression //解析【装饰器】

	p.nextToken()
	expr := p.parseExpressionStatement().Expression //解析【被装饰函数】
	//判断【被装饰函数】的类型，必须是一个函数或者另一个装饰器
	switch nodeType := expr.(type) {
	case *ast.FunctionLiteral:
		if nodeType.Name == "" {
			msg := fmt.Sprintf("Syntax Error:%v- decorator must be followed by a named function or another decorator", p.curToken.Pos)
			p.errors = append(p.errors, msg)
			p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
			return nil
		}
		dc.Decorated = nodeType
	case *ast.DecoratorExpr:
		dc.Decorated = nodeType
	default:
		msg := fmt.Sprintf("Syntax Error:%v- decorator must be followed by a named function or another decorator", p.curToken.Pos)
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}
	return dc
}
```

代码比较简单，就不多做解释了。



## 解释器（Evaluator）的更改

因为我们新增了一个装饰器的抽象语法表示`DecoratorExpr`，因此我们需要在`Eval`函数的`switch`语句中追加一个`case`分支：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {

	switch node := node.(type) {
	//..
	case *ast.DecoratorExpr:
		return evalDecorator(node, scope)
	}

	return nil
}
```

在实现第7行的`evalDecorator`函数之前，让我们先实现本节开头说的`$_`（它表示传递给函数的所有参数）的逻辑。我们先在`eval.go`文件中定义这个变量：

```go
//eval.go
var ALL_ARGS = "$_"
```



我们知道，函数调用的实际代码在`applyFunction`函数中，而`applyFunction`中，给函数赋参数的地方在`extendFunctionScope`函数中，因此我们需要在这个函数中将`$_`的逻辑加进去：

```go
//eval.go
func extendFunctionScope(fn *Function, args []Object) *Scope {
	scope := NewScope(fn.Scope, nil)

	//已知逻辑

	scope.Set(ALL_ARGS, &Array{Members: args}) //设置"$_"表示函数的所有参数
	return scope
}
```

还有一个地方需要更改，就是`applyFunction`函数中处理`TailCall`的部分：

```go
//eval.go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function:
		extendedScope := extendFunctionScope(fn, args)
		evaluated := Eval(fn.Literal.Body, extendedScope)
		if evaluated.Type() == TAIL_OBJ {
				//...

				extendedScope.Set(ALL_ARGS, &Array{Members: args2})
				o = Eval(fn2.Literal.Body, extendedScope)
	}
}
```

第10行是新增的代码。在解释函数体（`body`）之前， 我们将所有的函数参数放入`$_`中。

讲解了上面增加`$_`相关逻辑的代码后，我们再来看一下本节开头说的一个其中一个限制：

```
装饰器函数中调用被装饰的函数的语法必须是`functionName($_)`这种形式，即括号中有且仅有一个参数，而且参数名必须是`$_`。
```

我们知道函数调用的逻辑在`evalCallExpression`函数中：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
    args := evalExpressions(node.Arguments, scope)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	//...
	return applyFunction(node.Pos().Sline(), scope, function, args)
}
```

代码第3行的`evalExpressions`就是解释函数参数的部分，我们可以在这里加入这个限制：

```go
//eval.go
func evalCallExpression(node *ast.CallExpression, scope *Scope) Object {
	var args []Object
	if len(node.Arguments) == 1 && node.Arguments[0].TokenLiteral() == ALL_ARGS {
		if arr, ok := scope.Get(ALL_ARGS); ok {
			for _, v := range arr.(*Array).Members {
				args = append(args, v)
			}
		}
	} else {
		args = evalExpressions(node.Arguments, scope)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
	}

	//...
	return applyFunction(node.Pos().Sline(), scope, function, args)
}
```

第4-10行的`if`判断是新增的逻辑，我们判断函数的参数个数是1，且参数名是`$_`，则将`$_`所表示的所有参数追加到`args`数组变量中。

讲解完上面这些，现在轮到真正的主角，即`evalDecorator`函数的实现了：

```go
//eval.go
func evalDecorator(node *ast.DecoratorExpr, scope *Scope) Object {
	name, fn, err := _evalDecorator(node, scope)
	if isError(err) {
		return err
	}
	scope.Set(name, fn)
	return NIL
}

/* 函数有三个返回值：
   第一个返回值： 【被装饰函数】的函数名
   第二个返回值： 解释装饰器的返回值
   第三个返回值： 错误对象（Error Object）
*/
func _evalDecorator(node *ast.DecoratorExpr, scope *Scope) (string, Object, Object) {
    decorator := Eval(node.Decorator, scope) //解释【修饰器(decorator)】自身
	if isError(decorator) {
		return "", nil, decorator
	}

	if _, ok := decorator.(*Function); !ok { //如果【修饰器(decorator)】不是函数，则报错
		return "", nil, newError(node.Pos().Sline(), ERR_DECORATOR, decorator.Inspect())
	}

	name, ok := getDecoratedFuncName(node.Decorated) //获取【被装饰函数】的函数名
	if !ok {
		return "", nil, newError(node.Pos().Sline(), ERR_DECORATED_NAME)
	}

	//判断【装饰器】的类型（必须是一个函数或者另一个装饰器）
	decoratorFn := decorator.(*Function)
	switch decorated := node.Decorated.(type) {
	case *ast.FunctionLiteral:
		decoratedFn := &Function{Literal: decorated, Scope: scope}
		return name, 
				applyFunction(decorated.Pos().Sline(), scope, decoratorFn, []Object{decoratedFn}),
				nil
	case *ast.DecoratorExpr:
		//先解释最后一个装饰器（注意：这里是递归调用）
		name, decoratedFn, err := _evalDecorator(decorated, scope)
		if isError(err) {
			return "", nil, err
		}

		return name,
               applyFunction(node.Pos().Sline(), scope, decoratorFn, append([]Object{decoratedFn})),
               nil
	}

	//这里正常来说应该不会走到。但是由于编译的时候会报告函数少了返回值，所以这里加上。
	return "", nil, newError(node.Pos().Sline(), ERR_DECORATOR_FN)
}

// 得到实际【被装饰函数】的函数名
func getDecoratedFuncName(decorated ast.Expression) (string, bool) {
	switch d := decorated.(type) {
	case *ast.FunctionLiteral:
		return d.Name, true
	case *ast.DecoratorExpr:
		return getDecoratedFuncName(d.Decorated)
	}

	return "", false
}
```

这里比较难理解的实际是37和47行的`applyFunction`函数的参数。举个例子可能有助于理解：

```javascript
@decorator1
fn demo() {}
```

对于上面这个例子，我们希望解释后是下面这样的：

```go
decorator1(demo)
```

即【被装饰函数】作为【装饰器】的参数。我们再来看37和47行的`applyFunction`：

```go
applyFunction(decorated.Pos().Sline(), scope, decoratorFn, []Object{decoratedFn}),
```

第三个参数`decoratorFn`即【装饰器函数】，第四个参数为【装饰器函数】的参数，即【被装饰函数】（`decoratedFn`）。`_evalDecorator`函数返回后，我们需要将【被装饰函数】的函数名重新赋值成新的`_evalDecorator`的第二个返回值（即`applyFunction`的返回值），放到Scope中（代码第7行）。为什么这样做呢？因为当我们调用【被装饰函数】的时候，我们不是调用【被装饰函数】自身，而是应该调用【装饰器函数】。还以本节开头的例子来说明：

```javascript
//装饰器函数
fn log(otherfn) {
    return fn() {
        println("otherfn start")
        otherfn($_)
        println("otherfn end")
    }
}

@log //使用【@装饰器函数名】来声明一个装饰器
fn sum(x, y) { //sum函数是被装饰器函数修饰的函数，即【被装饰函数】
    printf("%d + %d = %d\n", x, y, x+y)
}

sum(1,2)
```

第15行我们调用【被装饰函数】`sum`，实际上是调用【装饰器函数】。因此，我们需要将第15行这个调用，在解释器内部的`Scope`中变成下面这样：

```
"sum" -> 【装饰器函数】
```

即key为`sum`，value为【装饰器函数】。最后还有一点需要说明，【被装饰函数】的【装饰器函数】可以有多个，像下面这样：

```
@decorator1
@decorator2
fn demo() {}
```

我们需要将其变成下面这样的表示：

```
demo = decorator1(decorator2(demo))
```

这就是`_evalDecorator`函数的第41行代码中，递归调用自身而实现的。

最后，在`_evalDecorator`函数的实现中，我们使用了三个错误常量，它们是定义在`errors.go`文件中的：

```go
//errors.go
var (
	//...
	ERR_DECORATOR       = "decorator '%s' is not a function"
	ERR_DECORATED_NAME  = "can not find the name of the decorated function"
	ERR_DECORATOR_FN    = "a decorator must decorate a named function or another decorator"
)
```





## 测试

```javascript
fn timer(otherfn) {
    return fn() {
        println("timer start")
        otherfn($_)
        println("timer end")
    }
}

fn log(otherfn) {
    return fn() {
        println("otherfn start")
        otherfn($_)
        println("otherfn end")
    }
}

@log
@timer
fn sum(x, y) {
    printf("%d + %d = %d\n", x, y, x+y)
}

sum(1,2)
```



下一节，我们会加入`命令执行（Command Execution）`的支持
