# 尾递归调用优化（Tail Recursive Call Optimization）

这一节中，我们会给`magpie`语言加入`尾递归调用优化`的功能。什么是尾递归调用（`tail recursive call`）?具体来说就是指某个函数的最后一步是调用自己。举个例子，读者就明白了:

```go
fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    return TailRecursive(number-1, product)
}
```

注意这里说的`调用自己`，这个条件非常苛刻。像下面的例子就不符合条件：

```go
//情况1
fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    return TailRecursive(number-1, product) + 2
}

//情况2
fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    result = TailRecursive(number-1, product)
    return result
}
```

学过编程的人应该都比较清楚，对于这种调用，如果递归的次数非常多的话，会造成`Stack Overflow`的致命异常（即栈溢出）。因为每次函数调用，都会将函数的参数及函数中的变量等入栈，函数返回后释放掉。当函数递归调用层级非常大的时候，导致栈没有更多的空间存放这些信息的话，栈就会溢出。

来看一下网上找到的说明。对于`go`语言来说，如果条件达到上限，就可能会造成`Stack Overflow`的异常：

> In Golang, stacks grow as needed. It makes the effective recursion limits relatively large. The initial setting is 1 GB on 64-bit systems and 250 MB on 32-bit systems. The default can be changed by [SetMaxStack](https://golang.org/pkg/runtime/debug/#SetMaxStack) in the runtime/debug package.
>
> =>（翻译一下）
>
> 在go语言中，堆栈会根据需要增长。这使得有效递归限制相对较大。在64位系统中，它的初始设置为1GB，32位系统的初始设置为250MB。我们可以通过修改`runtime/debug`包中的`SetMaxStack`方法来更改。

实现更改之前，我需要重点强调以下几点：

1. 函数最后一步必须是__直接调用自己__（不能有任何变形或者其它形式）

2. 不支持非函数调用。

   > ```go
   > struct {
   > 	fn TailRecursive(number, product) {
   > 		product = product + number
   > 		if number == 1 {
   > 			return product
   > 		}
   > 
   >    		//这里的self.TailRecursive是方法调用(method call),不是
   > 		//函数调用。所以也不支持这种。
   > 		return self.TailRecursive(number-1, product)
   > 	}
   > }
   > ```

说实话，在实现这一节的内容前，我做了非常多的尝试，结果都是以失败告终。如果`go`语言自身支持尾调用优化（`tail call optimization`）的话，也许我们能够利用`go`语言的这种功能。但是很遗憾，`go`语言自身也不支持尾调用优化。

因为我们的解释器是`递归解释`的（具体来说就是解释器的`Eval`方法会递归调用自己），所以需要考虑太多的因素，修改起来非常的麻烦。最终的结果：递归把我自己都给搞糊涂了:smile:。

最终我想到了一个方法，就是加入一个新的关键字`tailcall`。然后尝试实现后，解决了问题。我们来看一下具体的例子：

```go
fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    //return TailRecursive(number-1, product)
    tailcall TailRecursive(number-1, product)
}
```

具体来说就是将`return`关键字更改成`tailcall`关键字。如果脚本中遇到`tailcall`关键字，它也会跟`return`一样，返回调用端，而不会接着执行`tailcall`之后的语句，只不过它需要做一些额外的工作。



下面看一下我们需要做哪些更改：

1. 在词元（Token）源码`token.go`中加入`tailcall`关键字
2. 在抽象语法树（AST）的源码`ast.go`中加入`tailcall`语句对应的抽象语法表示
3. 在语法解析器（Parser）的源码`parser.go`中加入对`tailcall`语句的语法解析。
4. 在解释器（Evaluator）的源码`eval.go`中修改相关的解释逻辑。



## 词元（TOKEN）的更改

直接看代码。俗话说`代码如金`嘛，抵得上一千句话：

```go
//token.go
const (
	//...
	TOKEN_TAIL        //tail call
)

//词元类型的字符串表示
func (tt TokenType) String() string {
	switch tt {
	//...
	case TOKEN_TAIL:
		return "TAILCALL"
    }
}

var keywords = map[string]TokenType{
	//...
	"tailcall":    TOKEN_TAIL,
}
```


## 抽象语法树（AST）的更改

我们需要加入一个`TailCallStatement`的结构用来表示`tailcall funcCall(param1, param2, ...)`这种形式的语法。

```go
//语法：tailcall <function-call>
//     举例：tailcall funcCall(param1, param2, ...)
type TailCallStatement struct {
	Token token.Token // the 'tailcall' token
	Call  Expression
}

func (ts *TailCallStatement) Pos() token.Position {
	return ts.Token.Pos
}

func (ts *TailCallStatement) End() token.Position {
	return ts.Call.End()
}

func (ts *TailCallStatement) statementNode()       {}
func (ts *TailCallStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TailCallStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ts.TokenLiteral() + " ")
	out.WriteString(ts.Call.String())
	out.WriteString(";")

	return out.String()
}
```

这里需要注意的是`tailcall`后面必须跟函数调用。这个限制我们会在语法解析阶段来检查。

## 语法解析器（Parser）的更改

同样的只列出代码：

```go
//parser.go
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	//...
	case token.TOKEN_TAIL:
		return p.parseTailCallStatement()
	}
}

func (p *Parser) parseTailCallStatement() *ast.TailCallStatement {
	stmt := &ast.TailCallStatement{Token: p.curToken}

	p.nextToken()
	stmt.Call = p.parseExpressionStatement().Expression
	//检查Call的类型
	switch stmt.Call.(type) {
	case *ast.CallExpression:
	default:
		msg := fmt.Sprintf("Syntax Error:%v- 'tailcall' must be followed by a function call")
		p.errors = append(p.errors, msg)
		p.errorLines = append(p.errorLines, p.curToken.Pos.Sline())
		return nil
	}

	return stmt
}
```

需要注意的一点就是18行的`default`分支，如果`tailcall`关键字后跟的不是函数调用，则报错。

## 解释器（Evaluator）的更改

这一部分是最难的部分。我们之前说了，只需要将`return`关键字换成`tailcall`关键字。所以这里我们的对象系统（`Object System`）中新增的一个对象（`TailCall`）。它和`return`对象非常相似。`return`后面跟的是一个值，而`TailCall`后面跟的是一个函数调用。

````go
//object.go
const (
	//...
	TAIL_OBJ = "TAIL_OBJ"
)

type TailCall struct {
	tail *ast.TailCallStatement //这个字段主要是为了保存`tailcall`关键字后面跟着的函数调用
}

func (tc *TailCall) Inspect() string  { return "tailcall" }
func (tc *TailCall) Type() ObjectType { return TAIL_OBJ }
func (tc *TailCall) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return newError(line, ERR_NOMETHOD, method, tc.Type())
}

````

因为我们增加了一个`TailCallStatement`这样一个抽象语法表示，因此解释器的`Eval`函数也要加入一个新的`case`分支来处理这个抽象语法表示：

```go
//eval.go
func Eval(node ast.Node, scope *Scope) (val Object) {
	//...
	switch node := node.(type) {
	//...
	case *ast.TailCallStatement:
		return &TailCall{tail: node}
	}

	return nil
}
```

遇见`tailcall`语句的时候，我们会生成一个`TailCall`对象返回。

在解释`Block块`的`evalBlockStatement`函数中，如果解释后的结果是一个`TailCall`对象，则和`return`语句一样，返回这个`TailCall`对象：

```go
//eval.go
func evalBlockStatement(block *ast.BlockStatement, scope *Scope) Object {
	var result Object
	for _, statement := range block.Statements {
		result = Eval(statement, scope)
		if result != nil {
			rt := result.Type()
			if rt == RETURN_VALUE_OBJ || rt == ERROR_OBJ || rt == THROW_OBJ || 
				rt == TAIL_OBJ {
				return result
			}
		}
		//...
	}
	return result
}
```

第9行是新增的代码。

接下来是我们的核心处理逻辑了。从之前的文章中，读者已经知道的处理函数调用的实际逻辑是`applyFunction`方法，这里贴出部分代码：

```go
//eval.go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function: //如果是普通函数调用
		extendedScope := extendFunctionScope(fn, args) //扩展scope
		evaluated := Eval(fn.Literal.Body, extendedScope) //在扩展scope中解释函数体
		return unwrapReturnValue(evaluated)
	case *Builtin: //如果是内置函数调用
		return fn.Fn(line, scope, args...)
	default:
		return newError(line, ERR_NOTFUNCTION, fn.Type())
	}
}
```

再来看以下`extendFunctionScope`函数：

```go
//eval.go
func extendFunctionScope(fn *Function, args []Object) *Scope {
	scope := NewScope(fn.Scope, nil)
	//...
	return scope
}
```

这里列出`extendFunctionScope`函数的目的就是让大家注意第3行的代码`scope := NewScope(fn.Scope, nil)`。从这行代码中可以看到，每次函数调用，都会生成一个新的`scope`（因为每个函数都有自己的`scope`）。那么如果对于非常大的递归函数调用来说，就可能会产生`stack overflow`这种栈溢出的情况。

那么我们怎么来改变这种情况呢？如果我们可以在每次循环中重复利用已经存在的`scope`，那么是不是就可以解决这个问题？答案是肯定的。那么如何实现呢？还是来看代码吧：

```go
//eval.go
func applyFunction(line string, scope *Scope, fn Object, args []Object) Object {
	switch fn := fn.(type) {
	case *Function:
		extendedScope := extendFunctionScope(fn, args)
		evaluated := Eval(fn.Literal.Body, extendedScope)
		//如果函数体返回的是一个`TailCall`对象。还记得我们修改的`evalBlockStatement`函数吧。
		//当遇到`TailCall`对象时候，会提前返回。
		if evaluated.Type() == TAIL_OBJ {
			//把返回对象（即TailCall对象）中的CallExpression结构取出来
			call := evaluated.(*TailCall).tail.Call.(*ast.CallExpression)

			var o Object
			needContinue := true
			//将递归函数调用扩展为一个简单的for循环控制结构
			for needContinue {
				args2 := evalExpressions(call.Arguments, extendedScope) //解释函数参数
				if len(args2) == 1 && isError(args2[0]) { //如果出错则返回错误对象
					return args2[0]
				}

				if call.Variadic { //如果有可变参数，需要先拆箱
					args2 = getVariadicArgs(call, args2)
					if len(args2) == 1 && isError(args2[0]) {
						return args2[0]
					}
				}

				function := Eval(call.Function, extendedScope)
				if isError(function) {
					return function
				}

				fn2 := function.(*Function)
				argObjTable := make(map[string]Object)
				for i, identNode := range fn2.Literal.Parameters {
						argObjTable[identNode.Value] = args2[i]
				}

				//这个是最重要的部分，我们重复利用`scope`,而不是每次调用生成一个新的`scope`
				extendedScope.store = argObjTable     //函数参数
				extendedScope.parentScope = fn2.Scope //函数作用域
				extendedScope.Writer = scope.Writer

				if fn2.Literal.Variadic { //如果有可变参数，则装箱
					ellipsisArgs := args2[len(fn2.Literal.Parameters)-1:]
					newArgs := make([]Object, 0, len(fn2.Literal.Parameters)+1)
					newArgs = append(newArgs, args2[:len(fn2.Literal.Parameters)-1]...)
					args2 = append(newArgs, &Array{Members: ellipsisArgs})
					for i, arg := range args2 {
						extendedScope.Set(fn2.Literal.Parameters[i].String(), arg)
					}
				} else {
					for paramIdx, param := range fn2.Literal.Parameters {
						extendedScope.Set(param.Value, args2[paramIdx])
					}
				}

				o = Eval(fn2.Literal.Body, extendedScope) //在extendedScope中解释函数体
				if o.Type() == ERROR_OBJ {
					return o
				}
				if tailcall, ok := o.(*TailCall); ok { //如果是`TailCall`对象则继续循环
					needContinue = true
					call = tailcall.tail.Call.(*ast.CallExpression)
				} else {
					needContinue = false //不是'TailCall'对象，则退出循环
				}
			}
			return unwrapReturnValue(o)
		} else { //如果不是'TailCall'对象
			return unwrapReturnValue(evaluated)
		}
	//...
	}
}
```

核心的注释都写在代码里了。如果读者仔细看代码的话，不难猜到第16行的`for`循环中的很多逻辑，实际就是`evalCallExpression`函数中的处理逻辑。只不过在每次循环调用的时候，我们没有生成新的`scope`，而是重写一个`scope`。

我们写一个简单的程序测试一下：

```perl
fn testTC(n) { // tc:tail call
    if n == 0 {
        println("success! we are tail call optimized")
    } else {
        printf("n=%d\n", n)
        tailcall testTC(n - 1)
    }
}
testTC(2000000)
```

这个测试能够正常通过，而不会产生`stack overflow`的错误。如果我们将第6行的`tailcall testTC(n-1)`更改为`return testTC(n - 1) `则会产生`stack overflow`错误。但是我发现了一个奇怪的问题，就是函数结束后，打印`success! we are tail call optimized`后并没有结束。需要手动按回车键才行（有哪位读者知道是怎么回事嘛？）。为了避免按回车键，所以我加入了一个新的`flushStdout`内嵌函数：

```go
//builtin.go
func init() {
	builtins = map[string]*Builtin{
		//...
		"flushStdout":   flushStdoutBuiltin(),
	}
}

func flushStdoutBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return newError(line, ERR_ARGUMENT, 0, len(args))
			}

			os.Stdout.Sync() //刷新标准输出
			return NIL
		},
	}
}
```

现在代码变成了下面这样：

```perl
fn testTC(n) { // tc:tail call
    if n == 0 {
        println("success! we are tail call optimized")
		flushStdout()
    } else {
        printf("n=%d\n", n)
        
        tailcall testTC(n - 1)
    }
}
testTC(2000000)
```



## 测试

```javascript
fn TailRecursive(number, product) {
    product = product + number
    if number == 1 {
        return product
    }

    //return TailRecursive(number-1, product)  //将会汇报'stack overflow'的错误
    tailcall TailRecursive(number-1, product)
}

answer = TailRecursive(400000, 0)
printf("Recursive: %g\n", answer)
```



下一节，我们会加入`字符串变量内插`支持。



